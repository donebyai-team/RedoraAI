package interactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/shank318/doota/errorx"
	"net/http"
	"strings"
)

type browserless struct {
	Token string
}

func NewBrowserlessClient(token string) *browserless {
	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	fmt.Println("playwrite", err)
	return &browserless{Token: token}
}

func (r browserless) SendDM(params DMParams) error {
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

	if err := r.tryLogin(page, params); err != nil {
		return err
	}

	chatURL := "https://chat.reddit.com/user/" + params.To
	if _, err := page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
		return fmt.Errorf("chat page navigation failed: %w", err)
	}

	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg != "" {
			return fmt.Errorf("chat error: %s", msg)
		}
		return fmt.Errorf("chat error: invalid user")
	}

	textarea, err := page.WaitForSelector("rs-message-composer textarea[name='message']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(4000),
	})
	if err != nil {
		return fmt.Errorf("message textarea not found: %w", err)
	}

	if err := textarea.Fill(params.Message); err != nil {
		return fmt.Errorf("filling message failed: %w", err)
	}

	sendBtn, err := page.WaitForSelector("rs-message-composer button[aria-label='Send message']", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(2500),
	})
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	if err := sendBtn.Click(); err != nil {
		return fmt.Errorf("clicking send failed: %w", err)
	}

	page.WaitForTimeout(1500)
	return nil
}

func (r browserless) CheckIfLogin(params DMParams) error {
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

	if err := r.tryLogin(page, params); err != nil {
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

	selectors := map[string]string{
		"username": "#login-username input[name='username']",
		"password": "#login-password input[name='password']",
		"button":   "button[type='button'] span:has-text('Log In')",
	}
	for name, selector := range selectors {
		if _, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(5000),
		}); err != nil {
			return fmt.Errorf("%s selector wait failed: %w", name, err)
		}
	}

	if err := page.Fill(selectors["username"], params.Username); err != nil {
		return fmt.Errorf("fill username failed: %w", err)
	}
	if err := page.Fill(selectors["password"], params.Password); err != nil {
		return fmt.Errorf("fill password failed: %w", err)
	}

	page.WaitForTimeout(1000)
	if err := page.Click(selectors["button"]); err != nil {
		return fmt.Errorf("login button click failed: %w", err)
	}

	page.WaitForTimeout(3000)

	if loginMsg := extractLoginErrors(page); loginMsg != "" {
		return &errorx.LoginError{Reason: loginMsg}
	}
	return nil
}

func extractLoginErrors(page playwright.Page) string {
	var errors []string
	helperTexts, err := page.QuerySelectorAll("faceplate-form-helper-text")
	if err != nil {
		return ""
	}

	for _, helper := range helperTexts {
		shadow, err := helper.EvaluateHandle("el => el.shadowRoot?.querySelector('#helper-text')?.innerText")
		if err != nil {
			continue
		}
		if txt, err := shadow.JSONValue(); err == nil && txt != nil {
			str := strings.TrimSpace(fmt.Sprintf("%v", txt))
			if str != "" {
				errors = append(errors, str)
			}
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

func (r browserless) getCDPUrl() (string, error) {
	query := `mutation {
		proxy(type: [document, xhr], country: US, sticky: true) { time }
		goto(url: "https://www.reddit.com/login", waitUntil: firstContentfulPaint) {
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
