package browser_automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"time"
)

type DMParams struct {
	ID          string
	Cookie      string // json array
	CountryCode string
	To          string
	ToUsername  string
	Message     string
}

type DailyWarmParams struct {
	ID          string
	Cookies     string
	CountryCode string
}

type RedditBrowserAutomation struct {
	provider       BrowserAutomationProvider
	logger         *zap.Logger
	debugFileStore dstore.Store
}

func NewRedditBrowserAutomation(provider BrowserAutomationProvider, logger *zap.Logger, debugFileStore dstore.Store) *RedditBrowserAutomation {
	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		logger.Warn("failed to install playwright", zap.Error(err))
	}
	return &RedditBrowserAutomation{provider: provider, logger: logger, debugFileStore: debugFileStore}
}

func (r RedditBrowserAutomation) ValidateCookies(ctx context.Context, cookiesJSON, alpha2CountryCode string) (config *models.RedditDMLoginConfig, err error) {
	optionalCookies, err := ParseCookiesFromJSON(cookiesJSON, true)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.provider.GetCDPInfo(ctx, CDPInput{
		StartURL:          chatURL,
		UseProxy:          true,
		LiveURL:           false,
		Alpha2CountryCode: alpha2CountryCode,
	})
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	defer func() {
		if info.ReleaseSession != nil {
			err = info.ReleaseSession()
			if err != nil {
				r.logger.Error("failed to release session", zap.Error(err))
				return
			}
		}
	}()

	browser, err := pw.Chromium.ConnectOverCDP(info.WSEndpoint)
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

func (r RedditBrowserAutomation) SendDM(ctx context.Context, params DMParams) (cookies []byte, err error) {
	logger := r.logger.With(zap.String("interaction_id", params.ID))
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.provider.GetCDPInfo(ctx, CDPInput{
		StartURL:          chatURL,
		UseProxy:          true,
		LiveURL:           false,
		Alpha2CountryCode: params.CountryCode,
	})
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	defer func() {
		if info.ReleaseSession != nil {
			err = info.ReleaseSession()
			if err != nil {
				logger.Error("failed to release session", zap.Error(err))
				return
			}
		}
	}()

	browser, err := pw.Chromium.ConnectOverCDP(info.WSEndpoint)
	if err != nil {
		r.logger.Error("failed to connect to browser", zap.Error(err))
		return nil, errors.New("unable to connect to the browser, will be retried in sometime")
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	// Defer cleanup and video save after context is closed
	defer func() {
		if err != nil {
			r.storeScreenshot("defer", params.ID, page)
		}
	}()

	// cookie flow
	optionalCookies, err := ParseCookiesFromJSON(params.Cookie, false)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	err = pageContext.AddCookies(optionalCookies)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	// Navigate to chat page
	chatPageURL := "https://chat.reddit.com/user/" + params.To
	if params.ToUsername != "" {
		chatPageURL = "https://www.reddit.com/user/" + params.ToUsername + "/"
	}
	if _, err = page.Goto(chatPageURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)}); err != nil {
		return nil, fmt.Errorf("chat page navigation failed: %w", err)
	}

	// Screenshot after chat page load (optional)
	r.storeScreenshot("chat", params.ID, page)

	// verify if logged in
	currentURL := page.URL()

	logger.Info("sending DM page",
		zap.String("chat_url", chatPageURL),
		zap.String("current_url", currentURL))

	if strings.Contains(currentURL, "/login") {
		return nil, fmt.Errorf("unable to login, please check your credentials or cookies and try again")
	}

	_, err = page.Reload()
	if err != nil {
		logger.Warn("failed to reload page", zap.Error(err))
	}

	// Screenshot after chat page load (optional)
	r.storeScreenshot("chat", params.ID, page)

	// Check for error banner on chat page
	if alert, _ := page.QuerySelector("faceplate-banner[appearance='error']"); alert != nil {
		msg, _ := alert.GetAttribute("msg")
		if msg != "" {
			return nil, fmt.Errorf("chat error: %s", msg)
		}
		return nil, fmt.Errorf("chat error: invalid user")
	}

	locatorCurrentUser := page.Locator("rs-current-user")
	displayName, err := locatorCurrentUser.GetAttribute("display-name", playwright.LocatorGetAttributeOptions{
		Timeout: playwright.Float(5000), // Optional: Custom timeout for this action
	})
	if err != nil {
		logger.Error("failed to get display name", zap.Error(err))
	}

	if displayName != "" {
		logger.Error("logged in as user", zap.String("display_name", displayName))
	}

	if strings.Contains(currentURL, "www.reddit.com/user") {

		locatorCloseChat := page.Locator("button[aria-label='Close chat window']")

		// Check if the close button exists
		count, err := locatorCloseChat.Count()
		if err != nil {
			return nil, fmt.Errorf("error checking for close chat button: %w", err)
		}

		if count > 0 {
			err := locatorCloseChat.Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(3000), // short timeout for optional close
			})
			if err != nil {
				logger.Error("error clicking close chat button", zap.Error(err), zap.String("display_name", displayName))
			}
		}

		locatorStartChat := page.Locator("a[aria-label='Open chat']")
		err = locatorStartChat.Click(playwright.LocatorClickOptions{
			Delay: playwright.Float(100), // Delay before mouseup (in ms)
		})

		if err != nil {
			logger.Error("error clicking open chat button", zap.Error(err), zap.String("display_name", displayName))
			return nil, fmt.Errorf("unable to start chat: chat options might be disabled by the user")
		}
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
			logger.Info("found text area", zap.String("selector", sel))
			break
		}
	}

	if !found {
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
		logger.Warn("Unexpected navigation after sending message",
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

		logger.Warn("Reddit chat warning (ignorable)",
			zap.String("interaction", params.ID),
			zap.String("error_message", msg))
	}

	page.WaitForTimeout(1500)

	updatedCookies, err := pageContext.Cookies()
	if err != nil {
		return nil, err
	}

	logger.Info("updated cookies",
		zap.String("interaction", params.ID),
		zap.String("display_name", displayName),
		zap.Int("cookies", len(updatedCookies)))

	return json.Marshal(updatedCookies)
}

func (r RedditBrowserAutomation) storeScreenshot(stage, id string, page playwright.Page) {
	filePath := fmt.Sprintf("%s_%s.png", stage, id)
	byteData, screenShotErr := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true), // Optional: capture full page
	})
	if screenShotErr != nil {
		r.logger.Error("failed to take chat screenshot", zap.Error(screenShotErr))
	} else {
		buf := bytes.NewBuffer(byteData)
		if errFileStore := r.debugFileStore.WriteObject(context.Background(), filePath, buf); errFileStore != nil {
			r.logger.Error("failed to save chat screenshot", zap.Error(errFileStore), zap.String("output_name", filePath))
		}
	}
}

func (r RedditBrowserAutomation) WaitAndGetCookies(ctx context.Context, cdp *CDPInfo) (*models.RedditDMLoginConfig, error) {
	defer func() {
		if cdp.ReleaseSession != nil {
			err := cdp.ReleaseSession()
			if err != nil {
				r.logger.Error("failed to release session", zap.Error(err))
				return
			}
		}
	}()

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	// added a hack to reconnect wait
	time.Sleep(3 * time.Second)
	browser, err := pw.Chromium.ConnectOverCDP(cdp.WSEndpoint)
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

func (r RedditBrowserAutomation) StartLogin(ctx context.Context, alpha2CountryCode string) (*CDPInfo, error) {
	cdp, err := r.provider.GetCDPInfo(ctx, CDPInput{
		StartURL:          loginURL,
		UseProxy:          true,
		LiveURL:           true,
		Alpha2CountryCode: alpha2CountryCode,
	})
	if err != nil {
		return nil, err
	}

	return cdp, nil
}

func (r RedditBrowserAutomation) DailyWarmup(ctx context.Context, params DailyWarmParams) error {
	logger := r.logger.With(zap.String("integration_id", params.ID))
	warmupStartedAt := time.Now()

	// Step 1: Get CDP URL
	logger.Info("Fetching CDP URL")
	cdp, err := r.provider.GetCDPInfo(ctx, CDPInput{
		StartURL:          redditHomePage,
		UseProxy:          true,
		Alpha2CountryCode: params.CountryCode,
	})
	if err != nil {
		return fmt.Errorf("cDP url fetch failed: %w", err)
	}

	defer func() {
		if cdp.ReleaseSession != nil {
			err = cdp.ReleaseSession()
			if err != nil {
				logger.Error("failed to release session", zap.Error(err))
				return
			}
		}
	}()

	// Step 2: Start Playwright
	logger.Info("Starting Playwright")
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	// Step 3: Connect to browser
	logger.Info("Connecting to Chromium over CDP", zap.String("wsEndpoint", cdp.WSEndpoint))
	browser, err := pw.Chromium.ConnectOverCDP(cdp.WSEndpoint)
	if err != nil {
		return fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	// Step 4: Inject cookies
	logger.Info("Injecting cookies")
	optionalCookies, err := ParseCookiesFromJSON(params.Cookies, false)
	if err != nil {
		return fmt.Errorf("cookie injection failed: %w", err)
	}
	if err = pageContext.AddCookies(optionalCookies); err != nil {
		return fmt.Errorf("cookie injection failed: %w", err)
	}

	// Step 5: Go to Reddit home
	logger.Info("Navigating to Reddit home")
	if err = r.gotoWithRetry(page, "https://www.reddit.com", 30000); err != nil {
		return fmt.Errorf("home page navigation failed: %w", err)
	}

	// Wait for feed to load
	time.Sleep(5 * time.Second)
	logger.Info("Initial page load complete")

	// Step 6: Initial scroll to load posts
	logger.Info("Performing initial scroll to load more posts")
	for i := 0; i < 4; i++ {
		_, err = page.Evaluate(`window.scrollBy(0, 600)`)
		if err != nil {
			logger.Error("failed to scroll", zap.Error(err))
		} else {
			logger.Info("Scrolled feed", zap.Int("scrollIteration", i+1))
		}
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}

	// Step 7: Decide how many articles to visit
	rand.Seed(time.Now().UnixNano())
	numVisits := rand.Intn(2) + 4 // 4 or 5
	logger.Info("Starting article visits", zap.Int("numVisits", numVisits))

	for i := 0; i < numVisits; i++ {
		// Refresh article list
		logger.Info("Fetching latest articles", zap.Int("visit", i+1))
		feed := page.Locator("shreddit-feed")
		articles, err := feed.Locator("article").All()
		if err != nil {
			return fmt.Errorf("failed to get articles: %w", err)
		}
		if len(articles) == 0 {
			return fmt.Errorf("no articles found")
		}

		// Pick random article
		randomIndex := rand.Intn(5)
		logger.Info("Selected article", zap.Int("visit", i+1), zap.Int("index", randomIndex+1), zap.Int("totalArticles", len(articles)))

		selectedArticle := articles[randomIndex]

		// Scroll into view before clicking
		if err := selectedArticle.ScrollIntoViewIfNeeded(); err != nil {
			return fmt.Errorf("failed to scroll article into view: %w", err)
		}
		logger.Info("Scrolled article into view", zap.Int("visit", i+1))
		time.Sleep(500 * time.Millisecond) // small delay for rendering

		// Click article
		if err := selectedArticle.Click(); err != nil {
			return fmt.Errorf("failed to click article: %w", err)
		}
		logger.Info("Clicked article", zap.Int("visit", i+1))

		// Scroll inside post
		time.Sleep(3 * time.Second)
		scrollCount := rand.Intn(3) + 2
		for s := 0; s < scrollCount; s++ {
			if err := page.Keyboard().Press("PageDown"); err != nil {
				return fmt.Errorf("failed to scroll post: %w", err)
			}
			logger.Info("Scrolled post", zap.Int("scroll", s+1), zap.Int("totalScrolls", scrollCount))
			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
		}

		// Go back to feed
		logger.Info("Returning to home feed", zap.Int("visit", i+1))
		if err, _ := page.GoBack(); err != nil {
			return fmt.Errorf("failed to navigate back: %w", err)
		}

		// Wait for feed to reappear
		if _, err := page.WaitForSelector("shreddit-feed", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("feed not found after going back: %w", err)
		}
		logger.Info("Feed reloaded", zap.Int("visit", i+1))

		// Small delay before next article
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}

	logger.Info("Daily warmup complete", zap.Duration("duration", time.Since(warmupStartedAt)))
	return nil
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

func (r RedditBrowserAutomation) gotoWithRetry(page playwright.Page, url string, timeout float64) error {
	maxRetries := 2
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		r.logger.Info("navigating to url", zap.String("url", url))
		_, err := page.Goto(url, playwright.PageGotoOptions{
			Timeout:   playwright.Float(0),
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		if err == nil {
			return nil
		}

		lastErr = err

		if strings.Contains(err.Error(), "CONNECTION_FAILED") && i < maxRetries {
			r.logger.Error(fmt.Sprintf("Tunnel connection failed, retrying... (%d/%d)", i+1, maxRetries))
			time.Sleep(1 * time.Second) // backoff
			continue
		}

		break
	}

	return lastErr
}
