package interactions

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
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type steelBrowser struct {
	Token          string
	logger         *zap.Logger
	debugFileStore dstore.Store
}

type DailyWarmParams struct {
	Cookies string
}

func (r steelBrowser) DailyWarmup(ctx context.Context, params DailyWarmParams) error {
	// Step 1: Get CDP URL
	r.logger.Info("Fetching CDP URL")
	cdp, err := r.getCDPUrl(ctx, true)
	if err != nil {
		return fmt.Errorf("CDP url fetch failed: %w", err)
	}
	defer cdp.ReleaseSession()

	// Step 2: Start Playwright
	r.logger.Info("Starting Playwright")
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	// Step 3: Connect to browser
	r.logger.Info("Connecting to Chromium over CDP", zap.String("wsEndpoint", cdp.WSEndpoint))
	browser, err := pw.Chromium.ConnectOverCDP(cdp.WSEndpoint)
	if err != nil {
		return fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	// Step 4: Inject cookies
	r.logger.Info("Injecting cookies")
	optionalCookies, err := ParseCookiesFromJSON(params.Cookies, false)
	if err != nil {
		return fmt.Errorf("cookie injection failed: %w", err)
	}
	if err = pageContext.AddCookies(optionalCookies); err != nil {
		return fmt.Errorf("cookie injection failed: %w", err)
	}

	// Step 5: Go to Reddit home
	r.logger.Info("Navigating to Reddit home")
	if err = r.gotoWithRetry(page, "https://www.reddit.com", 30000); err != nil {
		return fmt.Errorf("home page navigation failed: %w", err)
	}

	// Wait for feed to load
	time.Sleep(5 * time.Second)
	r.logger.Info("Initial page load complete")

	// Step 6: Initial scroll to load posts
	r.logger.Info("Performing initial scroll to load more posts")
	for i := 0; i < 1; i++ {
		_, err = page.Evaluate(`window.scrollBy(0, 600)`)
		if err != nil {
			r.logger.Error("failed to scroll", zap.Error(err))
		} else {
			r.logger.Info("Scrolled feed", zap.Int("scrollIteration", i+1))
		}
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}

	// Step 7: Decide how many articles to visit
	rand.Seed(time.Now().UnixNano())
	numVisits := rand.Intn(2) + 4 // 4 or 5
	r.logger.Info("Starting article visits", zap.Int("numVisits", numVisits))

	for i := 0; i < numVisits; i++ {
		// Refresh article list
		r.logger.Info("Fetching latest articles", zap.Int("visit", i+1))
		feed := page.Locator("shreddit-feed")
		articles, err := feed.Locator("article").All()
		if err != nil {
			return fmt.Errorf("failed to get articles: %w", err)
		}
		if len(articles) == 0 {
			return fmt.Errorf("no articles found")
		}

		// Pick random article
		randomIndex := rand.Intn(len(articles))
		r.logger.Info("Selected article", zap.Int("visit", i+1), zap.Int("index", randomIndex+1), zap.Int("totalArticles", len(articles)))

		selectedArticle := articles[randomIndex]

		// Scroll into view before clicking
		if err := selectedArticle.ScrollIntoViewIfNeeded(); err != nil {
			return fmt.Errorf("failed to scroll article into view: %w", err)
		}
		r.logger.Info("Scrolled article into view", zap.Int("visit", i+1))
		time.Sleep(500 * time.Millisecond) // small delay for rendering

		// Click article
		if err := selectedArticle.Click(); err != nil {
			return fmt.Errorf("failed to click article: %w", err)
		}
		r.logger.Info("Clicked article", zap.Int("visit", i+1))

		// Scroll inside post
		time.Sleep(3 * time.Second)
		scrollCount := rand.Intn(3) + 2
		for s := 0; s < scrollCount; s++ {
			if err := page.Keyboard().Press("PageDown"); err != nil {
				return fmt.Errorf("failed to scroll post: %w", err)
			}
			r.logger.Info("Scrolled post", zap.Int("scroll", s+1), zap.Int("totalScrolls", scrollCount))
			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
		}

		// Go back to feed
		r.logger.Info("Returning to home feed", zap.Int("visit", i+1))
		if err, _ := page.GoBack(); err != nil {
			return fmt.Errorf("failed to navigate back: %w", err)
		}

		// Wait for feed to reappear
		if _, err := page.WaitForSelector("shreddit-feed", playwright.PageWaitForSelectorOptions{
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("feed not found after going back: %w", err)
		}
		r.logger.Info("Feed reloaded", zap.Int("visit", i+1))

		// Small delay before next article
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}

	r.logger.Info("Daily warmup complete")
	return nil
}

func NewSteelBrowserClient(token string, debugFileStore dstore.Store, logger *zap.Logger) BrowserAutomation {
	err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	if err != nil {
		logger.Warn("failed to install playwright", zap.Error(err))
	}
	return &steelBrowser{Token: token, logger: logger, debugFileStore: debugFileStore}
}

type CreateSession struct {
	UserAgent string `json:"userAgent"`
	//UseProxy  struct {
	//	GeoLocation struct {
	//		Country string `json:"country"`
	//	} `json:"geolocation"`
	//} `json:"useProxy"`
	UseProxy     bool `json:"useProxy"`
	SolveCaptcha bool `json:"solveCaptcha"`
	//Region        string `json:"region"`
	Timeout       int `json:"timeout"` // ms
	StealthConfig struct {
		HumanizeInteractions     bool `json:"humanizeInteractions"`
		SkipFingerprintInjection bool `json:"skipFingerprintInjection"`
	} `json:"stealthConfig"`
}

type Session struct {
	Id               string `json:"id"`
	Status           string `json:"status"`
	CreditsUsed      int    `json:"creditsUsed"`
	WebsocketUrl     string `json:"websocketUrl"`
	DebugUrl         string `json:"debugUrl"`
	SessionViewerUrl string `json:"sessionViewerUrl"`
	ProxyBytesUsed   int    `json:"proxyBytesUsed"`
	SolveCaptcha     bool   `json:"solveCaptcha"`
}

func (r steelBrowser) getCDPUrl(ctx context.Context, useProxy bool) (*CDPInfo, error) {
	const maxRetries = 3
	var backoff = 200 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		r.logger.Info("creating steel browser session", zap.Int("attempt", attempt+1))

		payload := CreateSession{
			SolveCaptcha: true,
			UseProxy:     useProxy,
			Timeout:      3 * 60 * 1000, // 3 minutes in ms
		}

		payload.StealthConfig.HumanizeInteractions = true

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("error marshalling json: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.steel.dev/v1/sessions", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Steel-Api-Key", r.Token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			r.logger.Warn("request failed, retrying...", zap.Error(err))
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			r.logger.Warn("unexpected response from Steel, retrying...",
				zap.Int("status_code", resp.StatusCode),
				zap.String("body", string(bodyBytes)),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		var sessionResp Session
		if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		r.logger.Info("steel browser raw response", zap.Any("body", sessionResp))

		return &CDPInfo{
			SessionID:  sessionResp.Id,
			WSEndpoint: fmt.Sprintf("%s&apiKey=%s", sessionResp.WebsocketUrl, r.Token),
			LiveURL:    fmt.Sprintf("%s?interactive=true&showControls=false", sessionResp.DebugUrl),
			ReleaseSession: func() error {
				r.logger.Info("releasing steel browser session", zap.String("session_id", sessionResp.Id))
				err := r.releaseSession(context.Background(), sessionResp.Id)
				if err != nil {
					r.logger.Error("failed to release session", zap.Error(err), zap.String("session_id", sessionResp.Id))
				}
				return err
			},
		}, nil
	}

	return nil, errors.New("failed to create steel browser session after retries")
}

func (r steelBrowser) releaseSession(ctx context.Context, sessionID string) error {
	url := fmt.Sprintf("https://api.steel.dev/v1/sessions/%s/release", sessionID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Steel-Api-Key", r.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %s - %s", resp.Status, string(bodyBytes))
	}

	r.logger.Info("steel browser session released", zap.String("session_id", sessionID))

	return nil
}

func (r steelBrowser) StartLogin(ctx context.Context) (*CDPInfo, error) {
	cdp, err := r.getCDPUrl(ctx, true)
	if err != nil {
		return nil, err
	}

	return cdp, nil
}

func (r steelBrowser) WaitAndGetCookies(ctx context.Context, cdp *CDPInfo) (*models.RedditDMLoginConfig, error) {
	defer func() {
		cdp.ReleaseSession()
	}()

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.ConnectOverCDP(cdp.WSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("CDP connection failed: %w", err)
	}
	defer browser.Close()

	pageContext := browser.Contexts()[0]
	page := pageContext.Pages()[0]

	err = r.gotoWithRetry(page, loginURL, 30000)
	if err != nil {
		return nil, fmt.Errorf("login page navigation failed: %w", err)
	}

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

func (r steelBrowser) gotoWithRetry(page playwright.Page, url string, timeout float64) error {
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

func (r steelBrowser) ValidateCookies(ctx context.Context, cookiesJSON string) (config *models.RedditDMLoginConfig, err error) {
	optionalCookies, err := ParseCookiesFromJSON(cookiesJSON, true)
	if err != nil {
		return nil, fmt.Errorf("cookie injection failed: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.getCDPUrl(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	defer func() {
		info.ReleaseSession()
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

	err = r.gotoWithRetry(page, chatURL, 30000)
	if err != nil {
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

func (r steelBrowser) SendDM(ctx context.Context, params DMParams) (cookies []byte, err error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright start failed: %w", err)
	}
	defer pw.Stop()

	info, err := r.getCDPUrl(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("CDP url fetch failed: %w", err)
	}

	logger := r.logger.With(
		zap.String("session_id", info.SessionID),
		zap.String("interaction_id", params.ID))

	defer func() {
		info.ReleaseSession()
	}()

	browser, err := pw.Chromium.ConnectOverCDP(info.WSEndpoint)
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
			logger.Error("failed to send DM", zap.Error(err))
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
	if params.ToUsername != "" {
		chatURL = "https://www.reddit.com/user/" + params.ToUsername + "/"
	}

	err = r.gotoWithRetry(page, chatURL, 30000)
	if err != nil {
		return nil, fmt.Errorf("chat page navigation failed: %w", err)
	}

	logger.Info("chat page loaded", zap.String("chat_url", chatURL))

	// Screenshot after chat page load (optional)
	//r.storeScreenshot("chat", params.ID, page)

	// verify if logged in
	currentURL := page.URL()

	logger.Info("sending DM page",
		zap.String("chat_url", chatURL),
		zap.String("current_url", currentURL))

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

	locatorCurrentUser := page.Locator("rs-current-user")
	displayName, err := locatorCurrentUser.GetAttribute("display-name", playwright.LocatorGetAttributeOptions{
		Timeout: playwright.Float(5000), // Optional: Custom timeout for this action
	})
	if err != nil {
		r.logger.Error("failed to get display name", zap.Error(err))
	}

	if displayName != "" {
		r.logger.Error("logged in as user", zap.String("display_name", displayName))
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
				r.logger.Error("error clicking close chat button", zap.Error(err), zap.String("display_name", displayName))
			}
		}

		locatorStartChat := page.Locator("a[aria-label='Open chat']")
		err = locatorStartChat.Click(playwright.LocatorClickOptions{
			Delay: playwright.Float(100), // Delay before mouseup (in ms)
		})

		if err != nil {
			return nil, fmt.Errorf("unable to start chat: %w", err)
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
			r.logger.Info("found text area", zap.String("selector", sel))
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

	r.logger.Info("updated cookies",
		zap.String("interaction", params.ID),
		zap.String("display_name", displayName),
		zap.Int("cookies", len(updatedCookies)))

	return json.Marshal(updatedCookies)
}
