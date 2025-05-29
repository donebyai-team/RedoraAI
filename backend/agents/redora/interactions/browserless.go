package interactions

import (
	"bytes"
	goCtx "context"
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/shank318/doota/datastore"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"strings"
	"time"
)

var cookies = `[
    {
        "domain": ".reddit.com",
        "expirationDate": 1748435404.805688,
        "hostOnly": false,
        "httpOnly": true,
        "name": "token_v2",
        "path": "/",
        "sameSite": null,
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpzS3dsMnlsV0VtMjVmcXhwTU40cWY4MXE2OWFFdWFyMnpLMUdhVGxjdWNZIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyIiwiZXhwIjoxNzQ4NDM1NDA0LjczNzQ0OSwiaWF0IjoxNzQ4MzQ5MDA0LjczNzQ0OSwianRpIjoiNktOQmhoQ1lMMllLTlhKS3BLSXAxX0txVGJ5LWdnIiwiY2lkIjoiMFItV0FNaHVvby1NeVEiLCJsaWQiOiJ0Ml8xY3R5ZHQ2bjhnIiwiYWlkIjoidDJfMWN0eWR0Nm44ZyIsImxjYSI6MTczMTQzNjY1NjkzNSwic2NwIjoiZUp4a2tkR090REFJaGQtRmE1X2dmNVVfbTAxdGNZYXNMUWFvazNuN0RWb2NrNzA3Y0Q0cEhQOURLb3FGRENaWGdxbkFCRmdUclREQlJ1VDluTG0zZzJpTmU4dFlzWm5DQkZtd0ZEcmttTEdzaVFRbWVKSWF5eHNtb0lMTnlGeXV0R05OTFQwUUpxaGNNcmVGSHBjMm9ia2JpNTZkR0ZXNXJEeW9zVmZsMHRqR0ZMWW54amNicXcycHVDNm5Na25MUXZrc1h2VGpOOVczOXZtel9TYTBKOE9LcXVtQjNobEpDRzRzZnBpbTNkOVRrNTZ0Q3hhMTkzcVEydWQ2M0s1OTFpdzBPN2VmNl9sckl4bVhZMmgtSnZ0MzF5LWhBNDg4THpQcUFFYXM0VWNaZG1RZF9sVUhVTG1nSkdNSjR0TUk1TXJsMjM4SnRtdlR2OGJ0RXo5OE0tS21OX3pXRE5SekNlTFFwX0gxR3dBQV9fOFExZVRSIiwicmNpZCI6IlFEWGtOSEVJQUNCZTh6ZER0c3B5NERqRjlsWXpJWVFMblFqcVBVX1E4cm8iLCJmbG8iOjJ9.U2ZmvqjcFZRHtUlM0Qh1NG7f6phK3638mIuOY0bpSOpm6q9V6LHCKQoDpHgm4tcbfX746eqLPgSG7JAJ-hdxazkP7w0PkeBDTKpO75DegQBOmwbwo3ktzgtv-qAfjCWHYJccaFkq34tuxMsy4q3ayLmcFu23ek_XWddYNoN7FbFwrLy5U8XOH2cdClvmD0-OUrGyuY4jNFrzkvwEwwksbE31ikN9MWLbdPvyHKhaUiVWkixfft9eKd0Plp_XqgcoRRXUZVNlkdyJxC5Ps3zsKtW_mGjtlXe2T14slDXQwwhcuc7FMSB8sT8wbVDXRN9aM_O_X8CwB758RFUATMSCqA"
    },
    {
        "domain": ".reddit.com",
        "expirationDate": 1751964826.615271,
        "hostOnly": false,
        "httpOnly": false,
        "name": "csv",
        "path": "/",
        "sameSite": "no_restriction",
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "2"
    },
    {
        "domain": ".reddit.com",
        "hostOnly": false,
        "httpOnly": false,
        "name": "session_tracker",
        "path": "/",
        "sameSite": null,
        "secure": true,
        "session": true,
        "storeId": null,
        "value": "ffibbenemickrfmcho.0.1748349016542.Z0FBQUFBQm9OYkJZNUxnZUFNMzF3TzdtLTMtcDRpSkZvSzFWdzlHLW90aksyc2RNbEU2NHRCRHVNVXBEeDM1SHc5MEM3M3VXSzhLYUdIVXQ4eDZmQWx6WFVtWkVYS29IdGN1UDdzc0tHZldVenhIejVLNTBpV2stRVk4MDhLbVh0alBpOE5Vd2hicGs"
    },
    {
        "domain": "www.reddit.com",
        "expirationDate": 1763460952.912801,
        "hostOnly": true,
        "httpOnly": false,
        "name": "subreddit_sort",
        "path": "/",
        "sameSite": "strict",
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "AZdlhAc="
    },
    {
        "domain": ".reddit.com",
        "hostOnly": false,
        "httpOnly": false,
        "name": "rdt",
        "path": "/",
        "sameSite": "no_restriction",
        "secure": true,
        "session": true,
        "storeId": null,
        "value": "acda907dc700b7f30203261368cc9a94"
    },
    {
        "domain": ".reddit.com",
        "expirationDate": 1779440805,
        "hostOnly": false,
        "httpOnly": false,
        "name": "pc",
        "path": "/",
        "sameSite": null,
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "8v"
    },
    {
        "domain": ".www.reddit.com",
        "expirationDate": 1779882438,
        "hostOnly": false,
        "httpOnly": false,
        "name": "__stripe_mid",
        "path": "/",
        "sameSite": "strict",
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "f3754083-5e0b-40f7-b283-803e69ea2f3ffe04f1"
    },
    {
        "domain": ".reddit.com",
        "expirationDate": 1763987403.423477,
        "hostOnly": false,
        "httpOnly": true,
        "name": "reddit_session",
        "path": "/",
        "sameSite": null,
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpsVFdYNlFVUEloWktaRG1rR0pVd1gvdWNFK01BSjBYRE12RU1kNzVxTXQ4IiwidHlwIjoiSldUIn0.eyJzdWIiOiJ0Ml8xY3R5ZHQ2bjhnIiwiZXhwIjoxNzYzOTg3NDA0LjM3ODc2NywiaWF0IjoxNzQ4MzQ5MDA0LjM3ODc2NywianRpIjoiMXM1ZWhQd2JCakhhYW4zNC1IOXE4QUNMbHg5RS1nIiwiY2lkIjoiY29va2llIiwibGNhIjoxNzMxNDM2NjU2OTM1LCJzY3AiOiJlSnlLamdVRUFBRF9fd0VWQUxrIiwidjEiOiIxMzc3NjA2ODE5OTUyOTYsMjAyNS0wNS0yN1QxMjozMDowNCw1Mzg5ZmRlNDBiYTRkYzUxYzQxMTUxNWQwNmRlMTk0MjQwYjEyNTUwIiwiZmxvIjoyfQ.AFIXUvKg01h_ldsG5nM6UcaGmuqTYyCV2lfcECzF3gJvOqgg9m64Z1S0hVYSEkZ0_dS_pBjVf4k-JhHGnU5ZJjTp_h0oDdH92FrfE7qNJBZuDuwTVmn7WcATqeN_cCeLUUpTlaNO7C0fFc_N7ZuZ45hS899pnux7DpH0RBjlXhA8V3j567Aqse6E8UfxF7mKzB6qlrr32-1k-hpS-2WrXnyOMaZn5bX5LJZIGBMK61_vgXZf3Uz6sbQWwHpR9S7E5V4duzq2vrJeYuLDO5zyz27f0DOjmFyysk64DXrtwerNLq4EerzBLxJEr3VFTYTiFA4iUcruvm-BaoY8QMotVA"
    },
    {
        "domain": ".reddit.com",
        "expirationDate": 1751964826.298759,
        "hostOnly": false,
        "httpOnly": false,
        "name": "edgebucket",
        "path": "/",
        "sameSite": null,
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "zMJyQHhpmMpeQ06gMw"
    },
    {
        "domain": ".reddit.com",
        "hostOnly": false,
        "httpOnly": false,
        "name": "csrf_token",
        "path": "/",
        "sameSite": "strict",
        "secure": true,
        "session": true,
        "storeId": null,
        "value": "e789e3f7863499eca70446badc9530ea"
    },
    {
        "domain": ".reddit.com",
        "expirationDate": 1782909004.805261,
        "hostOnly": false,
        "httpOnly": false,
        "name": "loid",
        "path": "/",
        "sameSite": "no_restriction",
        "secure": true,
        "session": false,
        "storeId": null,
        "value": "000000001ctydt6n8g.2.1731436656935.Z0FBQUFBQm9OYkJNZUh4cllfSFZncGo3RXhRUlEyMUo4b1luVl8yN0NFbG10WVl0OG9OY0RsaUtXOVRjT1R4X3l6UzBCcUFFOWRMUXEwa3oyelFLXy1XX0d2TFJGcnVucnI4UEx2a3JESkhjb1dCOEt0bWVhOVRKcGZneGlFbHhPMGZxRnNraWJHMFQ"
    }
]`

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

func ParseCookiesFromJSON(jsonStr string) ([]playwright.OptionalCookie, error) {
	var rawCookies []playwright.OptionalCookie
	if err := json.Unmarshal([]byte(jsonStr), &rawCookies); err != nil {
		return nil, fmt.Errorf("failed to parse cookie JSON: %w", err)
	}

	if len(rawCookies) == 0 {
		return nil, fmt.Errorf("no cookies found in JSON")
	}

	//var cookies []playwright.OptionalCookie
	//for _, rc := range rawCookies {
	//	cookie := playwright.OptionalCookie{
	//		Name:     rc.Name,
	//		Value:    rc.Value,
	//		Domain:   rc.Domain,
	//		Path:     rc.Path,
	//		URL:      rc.URL,
	//		HttpOnly: rc.HttpOnly,
	//		Secure:   rc.Secure,
	//	}
	//
	//	if rc.ExpirationDate != nil {
	//		cookie.Expires = rc.ExpirationDate
	//	}
	//
	//	if rc.SameSite != nil {
	//		switch *rc.SameSite {
	//		case "strict":
	//			s := playwright.SameSiteAttributeStrict
	//			cookie.SameSite = s
	//		case "lax":
	//			s := playwright.SameSiteAttributeLax
	//			cookie.SameSite = s
	//		case "no_restriction", "none":
	//			s := playwright.SameSiteAttributeNone
	//			cookie.SameSite = s
	//		}
	//	}
	//
	//	cookies = append(cookies, cookie)
	//}

	return rawCookies, nil
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

func (r browserless) SendDM(ctx context.Context, params DMParams) ([]byte, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.getCDPUrl(ctx, loginURL, false)
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
		optionalCookies, err := ParseCookiesFromJSON(params.Cookie)
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

	// Wait for message textarea to load
	locator := page.Locator("rs-message-composer textarea[name='message']")
	if err = locator.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(20000),
	}); err != nil {
		return nil, fmt.Errorf("message textarea not found: %w", err)
	}

	if err := locator.Fill(params.Message); err != nil {
		return nil, fmt.Errorf("filling message failed: %w", err)
	}

	sendBtn := page.Locator("rs-message-composer button[aria-label='Send message']")
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

	page.WaitForTimeout(1500)

	updatedCookies, err := pageContext.Cookies()
	if err != nil {
		return nil, err
	}

	return json.Marshal(updatedCookies)
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

type loginCallback func() ([]byte, error)

func (r browserless) callback(ctx context.Context, browserURL string) ([]byte, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

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
			if strings.HasPrefix(currentURL, "https://www.reddit.com/") && !strings.Contains(currentURL, "/login") {
				// user has logged in and is redirected
				cookies, err := pageContext.Cookies()
				if err != nil {
					return nil, fmt.Errorf("failed to read cookies: %w", err)
				}
				return json.Marshal(cookies)
			}
		}
	}
}

func (r browserless) StartLogin(ctx context.Context) (string, loginCallback, error) {
	cdp, err := r.getCDPUrl(ctx, loginURL, true)
	if err != nil {
		return "", nil, err
	}

	return cdp.LiveURL, func() ([]byte, error) {
		// we pass the same context so it respects timeout/deadline/cancel
		return r.callback(ctx, cdp.BrowserWSEndpoint)
	}, nil
}

type CDPInfo struct {
	BrowserWSEndpoint string
	LiveURL           string
}

const loginURL = "https://www.reddit.com/login"
const chatURL = "https://chat.reddit.com"

// proxy(type: [document, xhr], country: US, sticky: true) { time }
func (r browserless) getCDPUrl(ctx context.Context, startURL string, includeLiveURL bool) (*CDPInfo, error) {
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`mutation {
		goto(url: "` + startURL + `", waitUntil: firstContentfulPaint) {
			status
		}`)

	if includeLiveURL {
		queryBuilder.WriteString(`
		live: liveURL(timeout: 600000) {
			liveURL
		}`)
	}

	queryBuilder.WriteString(`
		reconnect(timeout: 30000) {
			browserWSEndpoint
		}
	}`)

	reqBody := map[string]string{"query": queryBuilder.String()}
	reqBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		fmt.Sprintf("https://production-sfo.browserless.io/chrome/bql?token=%s", r.Token),
		"application/json",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ReconnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
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
