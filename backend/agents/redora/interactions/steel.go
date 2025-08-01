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

type BrowserAutomation interface {
	StartLogin(ctx context.Context) (*CDPInfo, error)
	WaitAndGetCookies(ctx context.Context, browserURL string) (*models.RedditDMLoginConfig, error)
}

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
	UserAgent     string `json:"userAgent"`
	UseProxy      bool   `json:"useProxy"`
	SolveCaptcha  bool   `json:"solveCaptcha"`
	Region        string `json:"region"`
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

func (r steelBrowser) getCDPUrl(ctx context.Context, startURL string, includeLiveURL, useProxy bool) (*CDPInfo, error) {
	payload := CreateSession{
		UseProxy:     useProxy,
		SolveCaptcha: true,
		Region:       "lax",
	}

	payload.StealthConfig.HumanizeInteractions = true
	payload.StealthConfig.SkipFingerprintInjection = true

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.steel.dev/v1/sessions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Steel-Api-Key", r.Token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %s - %s", resp.Status, string(bodyBytes))
	}

	var sessionResp Session
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	r.logger.Info("steel browser raw response", zap.Any("body", sessionResp))

	return &CDPInfo{
		SessionID:         sessionResp.Id,
		BrowserWSEndpoint: sessionResp.WebsocketUrl,
		LiveURL:           sessionResp.DebugUrl,
	}, nil
}

func (r steelBrowser) releaseSession(ctx context.Context, sessionID string) error {
	url := fmt.Sprintf("https://api.steel.dev/v1/sessions/%s/release", sessionID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Steel-Api-Key", r.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (r steelBrowser) StartLogin(ctx context.Context) (*CDPInfo, error) {
	cdp, err := r.getCDPUrl(ctx, loginURL, true, true)
	if err != nil {
		return nil, err
	}

	return cdp, nil
}

func (r steelBrowser) WaitAndGetCookies(ctx context.Context, browserURL string) (*models.RedditDMLoginConfig, error) {
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

	_, err = page.Goto(loginURL, playwright.PageGotoOptions{Timeout: playwright.Float(10000)})
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
