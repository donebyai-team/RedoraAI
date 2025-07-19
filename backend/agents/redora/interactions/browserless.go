package interactions

import (
	"bytes"
	goCtx "context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"strings"
	"time"
)

var cookiesTest = `[]`

type CookieFromJSON struct {
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	Domain         *string  `json:"domain"`
	Path           *string  `json:"path"`
	URL            *string  `json:"url"`
	ExpirationDate *float64 `json:"expirationDate"`
	HttpOnly       *bool    `json:"httpOnly"`
	Secure         *bool    `json:"secure"`
	SameSite       *string  `json:"sameSite"`
}

func ParseCookiesFromJSON(jsonStr string, isUserProvidedCookies bool) ([]playwright.OptionalCookie, error) {
	var rawCookies []playwright.OptionalCookie
	if err := json.Unmarshal([]byte(jsonStr), &rawCookies); err != nil {
		return nil, fmt.Errorf("failed to parse cookie JSON: %w", err)
	}

	if len(rawCookies) == 0 {
		return nil, fmt.Errorf("no cookies found in JSON")
	}

	if !isUserProvidedCookies {
		return rawCookies, nil
	}

	var cookies []playwright.OptionalCookie
	for _, rc := range rawCookies {
		cookie := playwright.OptionalCookie{
			Name:     rc.Name,
			Value:    rc.Value,
			Domain:   rc.Domain,
			Path:     rc.Path,
			URL:      rc.URL,
			HttpOnly: rc.HttpOnly,
			Secure:   rc.Secure,
		}

		if rc.Expires != nil {
			cookie.Expires = rc.Expires
		}

		if rc.SameSite != nil {
			switch *rc.SameSite {
			case "strict":
				s := playwright.SameSiteAttributeStrict
				cookie.SameSite = s
			case "lax":
				s := playwright.SameSiteAttributeLax
				cookie.SameSite = s
			case "no_restriction", "none":
				s := playwright.SameSiteAttributeNone
				cookie.SameSite = s
			}
		}

		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

type browserless struct {
	Token          string
	logger         *zap.Logger
	debugFileStore dstore.Store
	db             datastore.Repository
}

func NewBrowserlessClient(token string, debugFileStore dstore.Store, logger *zap.Logger) *browserless {
	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		logger.Warn("failed to install playwright", zap.Error(err))
	}
	return &browserless{Token: token, logger: logger, debugFileStore: debugFileStore}
}

func (r browserless) ValidateCookies(ctx context.Context, cookiesJSON string) (config *models.RedditDMLoginConfig, err error) {
	optionalCookies, err := ParseCookiesFromJSON(cookiesJSON, true)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.getCDPUrl(ctx, chatURL, false, true)
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	browser, err := pw.Chromium.ConnectOverCDP(info.BrowserWSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	err = pageContext.AddCookies(optionalCookies)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	chatURL := "https://chat.reddit.com/"
	if _, err = page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
		return nil, fmt.Errorf("chat page navigation failed: %w", err)
	}

	currentURL := page.URL()
	if strings.Contains(currentURL, "/login") {
		return nil, fmt.Errorf("unable to login, please check your credentials or cookies and try again")
	}

	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg != "" {
			return nil, fmt.Errorf("chat error: %s", msg)
		}
		return nil, fmt.Errorf("chat error: invalid user")
	}

	displayName, err := page.Locator("rs-current-user").GetAttribute("display-name")
	if err != nil {
		r.logger.Error("failed to get display name")
	} else {
		r.logger.Error("logged in as user", zap.String("display_name", displayName))
	}

	if displayName == "" {
		return nil, fmt.Errorf("unable to login, please check your credentials or cookies and try again")
	}

	// extract the browser cookies and save it
	// IMP: Do not save the one provided by the user, as it may be invalid format
	updatedCookies, err := pageContext.Cookies()
	if err != nil {
		return nil, err
	}

	marshal, err := json.Marshal(updatedCookies)
	if err != nil {
		return nil, err
	}

	config = &models.RedditDMLoginConfig{
		Cookies:  string(marshal),
		Username: displayName,
	}

	return config, nil
}

func (r browserless) SendDM(ctx context.Context, params DMParams) (cookies []byte, err error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.getCDPUrl(ctx, chatURL, false, true)
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	browser, err := pw.Chromium.ConnectOverCDP(info.BrowserWSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	//pageContext, err := browser.NewContext(playwright.BrowserNewContextOptions{})
	//if err != nil {
	//	return fmt.Errorf("context creation failed: %w", err)
	//}
	//
	//page, err := pageContext.NewPage()
	//if err != nil {
	//	_ = pageContext.Close() // clean up context if page creation fails
	//	return fmt.Errorf("page creation failed: %w", err)
	//}

	// Defer cleanup and video save after context is closed
	defer func() {
		if err != nil {
			r.storeScreenshot("defer", params.ID, page)
		}
	}()

	// cookie flow
	if params.Cookie != "" {
		optionalCookies, err := ParseCookiesFromJSON(params.Cookie, false)
		if err != nil {
			return nil, fmt.Errorf("cookie injection failed: %w", err)
		}

		err = pageContext.AddCookies(optionalCookies)
		if err != nil {
			return nil, fmt.Errorf("cookie injection failed: %w", err)
		}
	} else {
		// Login flow
		//if err = r.tryLogin(page, params); err != nil {
		//	return err
		//}
	}

	// Navigate to chat page
	chatURL := "https://chat.reddit.com/user/" + params.To
	if _, err = page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
		return nil, fmt.Errorf("chat page navigation failed: %w", err)
	}

	// Screenshot after chat page load (optional)
	r.storeScreenshot("chat", params.ID, page)

	// verify if logged in
	currentURL := page.URL()
	if strings.Contains(currentURL, "/login") {
		return nil, fmt.Errorf("unable to login, please check your credentials or cookies and try again")
	}

	// Check for error banner on chat page
	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg != "" {
			return nil, fmt.Errorf("chat error: %s", msg)
		}
		return nil, fmt.Errorf("chat error: invalid user")
	}

	displayName, err := page.Locator("rs-current-user").GetAttribute("display-name")
	if err != nil {
		r.logger.Error("failed to get display name")
	} else {
		r.logger.Error("logged in as user", zap.String("display_name", displayName))
	}

	// Wait for message textarea to load
	selectors := []string{
		"textarea[name='message']",
		"textarea[aria-label='Write message']",
		"rs-message-composer-old textarea[name='message']",
		"rs-message-composer textarea[name='message']",
	}

	var locator playwright.Locator
	found := false

	for _, sel := range selectors {
		locator = page.Locator(sel)
		err = locator.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(20000), // short timeout per selector
		})
		if err == nil {
			found = true
			r.logger.Info("found text area", zap.String("selector", sel))
			break
		}
	}

	if !found || locator == nil {
		return nil, fmt.Errorf("message textarea not found using any of the selectors")
	}

	if err := locator.Fill(params.Message); err != nil {
		return nil, fmt.Errorf("filling message failed: %w", err)
	}

	sendBtn := page.Locator("button[aria-label='Send message']")
	if err = sendBtn.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return nil, fmt.Errorf("send button not found: %w", err)
	}

	if err := sendBtn.Click(playwright.LocatorClickOptions{
		Delay: playwright.Float(100), // Delay before mouseup (in ms)
	}); err != nil {
		return nil, fmt.Errorf("clicking send failed: %w", err)
	}

	// Screenshot after chat page load (optional)
	r.storeScreenshot("click_send", params.ID, page)

	// Check if page navigated unexpectedly
	redirectedURL := page.URL()
	if !strings.Contains(redirectedURL, "/user/") {
		r.logger.Warn("Unexpected navigation after sending message",
			zap.String("interaction", params.ID),
			zap.String("redirected_to", redirectedURL))
	}

	// Check for error banner on chat page
	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg == "" {
			return nil, fmt.Errorf("chat error: unknown error with no message")
		}

		if !strings.Contains(strings.ToLower(msg), "unable to show the room") {
			return nil, fmt.Errorf("%s", msg)
		}

		r.logger.Warn("Reddit chat warning (ignorable)",
			zap.String("interaction", params.ID),
			zap.String("error_message", msg))
	}

	page.WaitForTimeout(1500)

	updatedCookies, err := pageContext.Cookies()
	if err != nil {
		return nil, err
	}

	r.logger.Info("updated cookies", zap.String("interaction", params.ID), zap.Int("cookies", len(updatedCookies)))

	return json.Marshal(updatedCookies)
}

func (r browserless) storeScreenshot(stage, id string, page playwright.Page) {
	filePath := fmt.Sprintf("%s_%s.png", stage, id)
	byteData, screenShotErr := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true), // Optional: capture full page
	})
	if screenShotErr != nil {
		r.logger.Error("failed to take chat screenshot", zap.Error(screenShotErr))
	} else {
		buf := bytes.NewBuffer(byteData)
		if errFileStore := r.debugFileStore.WriteObject(goCtx.Background(), filePath, buf); errFileStore != nil {
			r.logger.Error("failed to save chat screenshot", zap.Error(errFileStore), zap.String("output_name", filePath))
		}
	}
}

//func (r browserless) CheckIfLogin(params DMParams) (err error) {
//	pw, err := playwright.Run()
//	if err != nil {
//		return fmt.Errorf("playwright start failed: %w", err)
//	}
//	defer pw.Stop()
//
//	info, err := r.getCDPUrl()
//	if err != nil {
//		return fmt.Errorf("CDP url fetch failed: %w", err)
//	}
//
//	browser, err := pw.Chromium.ConnectOverCDP(info.BrowserWSEndpoint)
//	if err != nil {
//		return fmt.Errorf("CDP connection failed: %w", err)
//	}
//	defer browser.Close()
//
//	pageContext, err := browser.NewContext()
//	if err != nil {
//		return fmt.Errorf("context creation failed: %w", err)
//	}
//
//	page, err := pageContext.NewPage()
//	if err != nil {
//		return fmt.Errorf("page creation failed: %w", err)
//	}
//
//	// Screenshot on error (deferred)
//	defer func() {
//		if err != nil {
//			r.storeScreenshot("defer", params.ID, page)
//		}
//	}()
//
//	// cookie flow
//	if params.Cookie != "" {
//		optionalCookies, err := ParseCookiesFromJSON(params.Cookie)
//		if err != nil {
//			return fmt.Errorf("cookie injection failed: %w", err)
//		}
//
//		err = pageContext.AddCookies(optionalCookies)
//		if err != nil {
//			return fmt.Errorf("cookie injection failed: %w", err)
//		}
//	} else {
//		// Login flow
//		if err = r.tryLogin(page, params); err != nil {
//			return err
//		}
//	}
//
//	// Navigate to chat page
//	chatURL := "https://chat.reddit.com"
//	if _, err = page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
//		return fmt.Errorf("chat page navigation failed: %w", err)
//	}
//
//	// Screenshot after chat page load (optional)
//	r.storeScreenshot("login_verify_chat", params.ID, page)
//
//	// verify if logged in
//	currentURL := page.URL()
//	if strings.Contains(currentURL, "/login") {
//		return fmt.Errorf("unable to login, please check your credentials or cookies and try again")
//	}
//
//	return nil
//}

//func (r browserless) tryLogin(page playwright.Page, params DMParams) error {
//	if _, err := page.Goto("https://www.reddit.com/login", playwright.PageGotoOptions{
//		Timeout: playwright.Float(15000),
//	}); err != nil {
//		return fmt.Errorf("navigate to login failed: %w", err)
//	}
//
//	r.storeScreenshot("before_login", params.ID, page)
//
//	locators := map[string]playwright.Locator{
//		"username": page.Locator("#login-username input[name='username']"),
//		"password": page.Locator("#login-password input[name='password']"),
//		"button":   page.Locator("button.login"),
//	}
//
//	// Wait for all locators
//	for name, locator := range locators {
//		if err := locator.WaitFor(playwright.LocatorWaitForOptions{
//			Timeout: playwright.Float(5000),
//		}); err != nil {
//			return fmt.Errorf("%s locator wait failed: %w", name, err)
//		}
//	}
//
//	// Fill inputs
//	if err := locators["username"].Fill(params.Username); err != nil {
//		return fmt.Errorf("fill username failed: %w", err)
//	}
//
//	if err := locators["password"].Fill(params.Password); err != nil {
//		return fmt.Errorf("fill password failed: %w", err)
//	}
//	// Optional pause (but often unnecessary with locators)
//	page.WaitForTimeout(1000)
//
//	// Click the login button with a small delay to simulate realism
//	if err := locators["button"].Click(playwright.LocatorClickOptions{
//		Delay: playwright.Float(100), // Delay before mouseup (in ms)
//	}); err != nil {
//		return fmt.Errorf("login button click failed: %w", err)
//	}
//
//	page.WaitForTimeout(3000) // You can replace this with a proper navigation wait
//
//	r.storeScreenshot("after_login", params.ID, page)
//
//	if loginMsg := extractLoginErrors(page); loginMsg != "" {
//		return &errorx.LoginError{Reason: loginMsg}
//	}
//	return nil
//}
//
//func extractLoginErrors(page playwright.Page) string {
//	var errors []string
//
//	helpers := page.Locator("faceplate-form-helper-text")
//	count, err := helpers.Count()
//	if err != nil {
//		return ""
//	}
//
//	for i := 0; i < count; i++ {
//		helper := helpers.Nth(i)
//
//		txt, err := helper.Evaluate(`el => el.shadowRoot?.querySelector("#helper-text")?.innerText`, nil)
//		if err != nil {
//			continue
//		}
//
//		if str, ok := txt.(string); ok && strings.TrimSpace(str) != "" {
//			errors = append(errors, strings.TrimSpace(str))
//		}
//	}
//
//	return strings.Join(errors, " | ")
//}

func (r browserless) WaitAndGetCookies(ctx context.Context, browserURL string) (*models.RedditDMLoginConfig, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	// added a hack to reconnect wait
	time.Sleep(3 * time.Second)
	browser, err := pw.Chromium.ConnectOverCDP(browserURL)
	if err != nil {
		return nil, fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("login timed out or cancelled: %w", ctx.Err())
		case <-ticker.C:
			currentURL := page.URL()
			if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
				msg, _ := alert.GetAttribute("msg")
				if msg != "" {
					return nil, errors.New(msg)
				}
			}
			
			if (strings.HasPrefix(currentURL, "https://www.reddit.com") || strings.HasPrefix(currentURL, "https://chat.reddit.com")) &&
				!strings.Contains(currentURL, "/login") {

				displayName, err := page.Locator("rs-current-user").GetAttribute("display-name")
				if err != nil {
					r.logger.Error("failed to get display name, while login")
				} else {
					r.logger.Error("logged in as user while login", zap.String("display_name", displayName))
				}

				cookies, err := pageContext.Cookies()
				if err != nil {
					return nil, fmt.Errorf("failed to read cookies: %w", err)
				}

				marshal, err := json.Marshal(cookies)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal cookies: %w", err)
				}

				if len(marshal) == 0 {
					return nil, errors.New("no cookies found")
				}

				loginConfig := &models.RedditDMLoginConfig{
					Username: displayName,
					Cookies:  string(marshal),
				}
				return loginConfig, nil
			}
		}
	}
}

func (r browserless) StartLogin(ctx context.Context) (*CDPInfo, error) {
	cdp, err := r.getCDPUrl(ctx, loginURL, true, true)
	if err != nil {
		return nil, err
	}

	return cdp, nil
}

type CDPInfo struct {
	BrowserWSEndpoint string
	LiveURL           string
}

const loginURL = "https://www.reddit.com/login"
const chatURL = "https://chat.reddit.com"

// proxy(type: [document, xhr], country: US, sticky: true) { time }
func (r browserless) getCDPUrl(ctx context.Context, startURL string, includeLiveURL, useProxy bool) (*CDPInfo, error) {
	var queryBuilder strings.Builder

	queryBuilder.WriteString("mutation {")

	if useProxy {
		queryBuilder.WriteString(`
  proxy(
    type: [document, xhr],
    country: US,
    sticky: true
  ) {
    time
  }`)
	}

	queryBuilder.WriteString(`
  goto(
    url: "` + startURL + `",
    waitUntil: firstContentfulPaint
  ) {
    status
  }`)

	if includeLiveURL {
		queryBuilder.WriteString(`
  live: liveURL(timeout: 600000 quality: 30 type: jpeg) {
    liveURL
  }`)
	}

	queryBuilder.WriteString(`
  reconnect {
    browserWSEndpoint
  }
}`)

	reqBody := map[string]string{"query": queryBuilder.String()}
	reqBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		fmt.Sprintf("https://production-sfo.browserless.io/chromium/bql?token=%s&humanlike=true&blockConsentModals=true", r.Token),
		"application/json",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	r.logger.Info("browserless raw response", zap.ByteString("body", bodyBytes))

	var result ReconnectResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	if result.Data.Reconnect.BrowserWSEndpoint == "" {
		return nil, errors.New("empty browserWSEndpoint - CDP connection failed")
	}

	info := &CDPInfo{
		BrowserWSEndpoint: result.Data.Reconnect.BrowserWSEndpoint,
	}

	if includeLiveURL {
		info.LiveURL = result.Data.Live.LiveURL
		r.logger.Info("browserless live url", zap.String("url", info.LiveURL))
	}

	return info, nil
}

type ReconnectResponse struct {
	Data struct {
		Reconnect struct {
			BrowserWSEndpoint string `json:"browserWSEndpoint"`
		} `json:"reconnect"`

		Live struct {
			LiveURL string `json:"liveURL"`
		} `json:"live"`
	} `json:"data"`
}
