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
	"net/http"
	"strings"
	"time"
)

type steelBrowser struct {
	Token          string
	logger         *zap.Logger
	debugFileStore dstore.Store
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
			SolveCaptcha: false,
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

	_, err = page.Goto(loginURL, playwright.PageGotoOptions{Timeout: playwright.Float(20000)})
	if err != nil {
		return nil, err
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
	if _, err = page.Goto(chatURL, playwright.PageGotoOptions{Timeout: playwright.Float(30000)}); err != nil {
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
		logger.Error("failed to get display name")
	} else {
		logger.Info("logged in as user", zap.String("display_name", displayName))
	}

	if displayName == "" {
		return nil, fmt.Errorf("unable to login, please check your credentials or cookies and try again")
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

		locatorStartChat := page.Locator("faceplate-tracker[action='click'][noun='chat'] a[href*='chat.reddit.com/user/']")
		err = locatorStartChat.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(20000), // short timeout per selector
		})
		if err != nil {
			logger.Error("error clicking start chat button", zap.Error(err), zap.String("interaction_id", params.ID))
			return nil, fmt.Errorf("chat could not be initiated. Direct messages may be disabled by the user")
		}

		err = locatorStartChat.Click(playwright.LocatorClickOptions{
			Delay: playwright.Float(100), // Delay before mouseup (in ms)
		})

		if err != nil {
			return nil, fmt.Errorf("unable to click start chat: %w", err)
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
	logger.Info("clicked send button")

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

	logger.Info("updated cookies", zap.String("interaction", params.ID), zap.Int("cookies", len(updatedCookies)))

	return json.Marshal(updatedCookies)
}
