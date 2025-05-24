package interactions

import (
	"bytes"
	goCtx "context"
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/shank318/doota/errorx"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"os"
	"strings"
)

type browserless struct {
	Token          string
	logger         *zap.Logger
	debugFileStore dstore.Store
}

func NewBrowserlessClient(token string, debugFileStore dstore.Store, logger *zap.Logger) *browserless {
	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		logger.Warn("failed to install playwright", zap.Error(err))
	}
	return &browserless{Token: token, logger: logger, debugFileStore: debugFileStore}
}

func (r browserless) SendDM(params DMParams) (err error) {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	url, err := r.getCDPUrl()
	if err != nil {
		return fmt.Errorf("CDP url fetch failed: %w", err)
	}

	browser, err := pw.Chromium.ConnectOverCDP(url)
	if err != nil {
		return fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	// Create a unique temporary directory for video
	//tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("video_run_%s", params.ID))
	//if err := os.MkdirAll(tmpDir, 0755); err != nil {
	//	return fmt.Errorf("failed to create temp video dir: %w", err)
	//}

	pageContext, err := browser.NewContext(playwright.BrowserNewContextOptions{
		//RecordVideo: &playwright.RecordVideo{
		//	Dir: tmpDir,
		//},
	})
	if err != nil {
		return fmt.Errorf("context creation failed: %w", err)
	}

	page, err := pageContext.NewPage()
	if err != nil {
		_ = pageContext.Close() // clean up context if page creation fails
		return fmt.Errorf("page creation failed: %w", err)
	}

	// Defer cleanup and video save after context is closed
	defer func() {
		//closeErr := pageContext.Close() // finalize video recording
		//if closeErr != nil {
		//	r.logger.Warn("failed to close context", zap.Error(closeErr))
		//}
		//
		//r.saveVideo(params.ID, page)

		if err != nil {
			r.storeScreenshot("defer", params.ID, page)
		}

		// Remove temp directory and video files
		//if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
		//	r.logger.Warn("failed to remove temp video directory", zap.Error(rmErr))
		//}
	}()

	// Login flow
	if err = r.tryLogin(page, params); err != nil {
		return err
	}

	// Navigate to chat page
	chatURL := "https://chat.reddit.com/user/" + params.To
	if _, err = page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("chat page navigation failed: %w", err)
	}

	// Screenshot after chat page load (optional)
	r.storeScreenshot("chat", params.ID, page)

	// Check for error banner on chat page
	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg != "" {
			return fmt.Errorf("chat error: %s", msg)
		}
		return fmt.Errorf("chat error: invalid user")
	}

	// Wait for message textarea to load
	locator := page.Locator("rs-message-composer textarea[name='message']")
	if err = locator.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(20000),
	}); err != nil {
		return fmt.Errorf("message textarea not found: %w", err)
	}

	if err := locator.Fill(params.Message); err != nil {
		return fmt.Errorf("filling message failed: %w", err)
	}

	sendBtn := page.Locator("rs-message-composer button[aria-label='Send message']")
	if err = sendBtn.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	if err := sendBtn.Click(playwright.LocatorClickOptions{
		Delay: playwright.Float(100), // Delay before mouseup (in ms)
	}); err != nil {
		return fmt.Errorf("clicking send failed: %w", err)
	}

	page.WaitForTimeout(1500)
	return nil
}

func (r browserless) saveVideo(id string, page playwright.Page) {
	if video := page.Video(); video != nil {
		videoPath, err := video.Path()
		if err == nil {
			file, err := os.Open(videoPath)
			if err == nil {
				defer file.Close()

				var buf bytes.Buffer
				if _, err := io.Copy(&buf, file); err == nil {
					objectName := fmt.Sprintf("video_%s.webm", id)
					if err := r.debugFileStore.WriteObject(context.Background(), objectName, &buf); err != nil {
						r.logger.Warn("failed to upload video", zap.Error(err))
					}
				}
			}
		}
	}
}

func (r browserless) storeScreenshot(stage, id string, page playwright.Page) {
	filePath := fmt.Sprintf("%s_%s.png", stage, id)
	byteData, screenShotErr := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true), // Optional: capture full page
	})
	if screenShotErr != nil {
		r.logger.Warn("failed to take chat screenshot", zap.Error(screenShotErr))
	} else {
		buf := bytes.NewBuffer(byteData)
		if errFileStore := r.debugFileStore.WriteObject(goCtx.Background(), filePath, buf); errFileStore != nil {
			r.logger.Debug("failed to save chat screenshot", zap.Error(errFileStore), zap.String("output_name", filePath))
		}
	}
}

func (r browserless) CheckIfLogin(params DMParams) (err error) {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	url, err := r.getCDPUrl()
	if err != nil {
		return fmt.Errorf("CDP url fetch failed: %w", err)
	}

	browser, err := pw.Chromium.ConnectOverCDP(url)
	if err != nil {
		return fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		return fmt.Errorf("context creation failed: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("page creation failed: %w", err)
	}

	// Screenshot on error (deferred)
	defer func() {
		if err != nil {
			r.storeScreenshot("defer", params.ID, page)
		}
	}()

	if err = r.tryLogin(page, params); err != nil {
		return err
	}
	return nil
}

func (r browserless) tryLogin(page playwright.Page, params DMParams) error {
	if _, err := page.Goto("https://www.reddit.com/login", playwright.PageGotoOptions{
		Timeout: playwright.Float(15000),
	}); err != nil {
		return fmt.Errorf("navigate to login failed: %w", err)
	}

	r.storeScreenshot("before_login", params.ID, page)

	locators := map[string]playwright.Locator{
		"username": page.Locator("#login-username input[name='username']"),
		"password": page.Locator("#login-password input[name='password']"),
		"button":   page.Locator("button.login"),
	}

	// Wait for all locators
	for name, locator := range locators {
		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("%s locator wait failed: %w", name, err)
		}
	}

	// Fill inputs
	if err := locators["username"].Fill(params.Username); err != nil {
		return fmt.Errorf("fill username failed: %w", err)
	}

	if err := locators["password"].Fill(params.Password); err != nil {
		return fmt.Errorf("fill password failed: %w", err)
	}
	// Optional pause (but often unnecessary with locators)
	page.WaitForTimeout(1000)

	// Click the login button with a small delay to simulate realism
	if err := locators["button"].Click(playwright.LocatorClickOptions{
		Delay: playwright.Float(100), // Delay before mouseup (in ms)
	}); err != nil {
		return fmt.Errorf("login button click failed: %w", err)
	}

	page.WaitForTimeout(3000) // You can replace this with a proper navigation wait

	r.storeScreenshot("after_login", params.ID, page)

	if loginMsg := extractLoginErrors(page); loginMsg != "" {
		return &errorx.LoginError{Reason: loginMsg}
	}
	return nil
}

func extractLoginErrors(page playwright.Page) string {
	var errors []string

	helpers := page.Locator("faceplate-form-helper-text")
	count, err := helpers.Count()
	if err != nil {
		return ""
	}

	for i := 0; i < count; i++ {
		helper := helpers.Nth(i)

		txt, err := helper.Evaluate(`el => el.shadowRoot?.querySelector("#helper-text")?.innerText`, nil)
		if err != nil {
			continue
		}

		if str, ok := txt.(string); ok && strings.TrimSpace(str) != "" {
			errors = append(errors, strings.TrimSpace(str))
		}
	}

	return strings.Join(errors, " | ")
}

type ReconnectResponse struct {
	Data struct {
		Reconnect struct {
			BrowserWSEndpoint string `json:"browserWSEndpoint"`
		} `json:"reconnect"`
	} `json:"data"`
}

//proxy(type: [document, xhr], country: US, sticky: true) { time }

func (r browserless) getCDPUrl() (string, error) {
	query := `mutation {
		goto(url: "https://www.reddit.com", waitUntil: firstContentfulPaint) {
			status
		}
		reconnect(timeout: 30000) {
			browserWSEndpoint
		}
	}`

	reqBody := map[string]string{"query": query}
	reqBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(fmt.Sprintf("https://production-sfo.browserless.io/chrome/bql?token=%s", r.Token), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result ReconnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Data.Reconnect.BrowserWSEndpoint, nil
}
