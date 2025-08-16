package browser_automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type steelBrowserClient struct {
	Token  string
	logger *zap.Logger
}

func NewSteelBrowser(token string, logger *zap.Logger) BrowserAutomationProvider {
	return &steelBrowserClient{Token: token, logger: logger}
}

type createSession struct {
	UserAgent string `json:"userAgent"`
	UseProxy  struct {
		GeoLocation struct {
			Country string `json:"country"`
		} `json:"geolocation"`
	} `json:"useProxy"`
	//UseProxy     bool `json:"useProxy"`
	SolveCaptcha bool `json:"solveCaptcha"`
	//Region        string `json:"region"`
	Timeout       int `json:"timeout"` // ms
	StealthConfig struct {
		HumanizeInteractions     bool `json:"humanizeInteractions"`
		SkipFingerprintInjection bool `json:"skipFingerprintInjection"`
	} `json:"stealthConfig"`
}

type session struct {
	Id               string `json:"id"`
	Status           string `json:"status"`
	CreditsUsed      int    `json:"creditsUsed"`
	WebsocketUrl     string `json:"websocketUrl"`
	DebugUrl         string `json:"debugUrl"`
	SessionViewerUrl string `json:"sessionViewerUrl"`
	ProxyBytesUsed   int    `json:"proxyBytesUsed"`
	SolveCaptcha     bool   `json:"solveCaptcha"`
}

func (r steelBrowserClient) GetCDPInfo(ctx context.Context, input CDPInput) (*CDPInfo, error) {
	const maxRetries = 3
	var backoff = 200 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		r.logger.Info("creating steel browser session", zap.Int("attempt", attempt+1))

		payload := createSession{
			SolveCaptcha: true,
			//UseProxy:     input.UseProxy,
			Timeout: 3 * 60 * 1000, // 3 minutes in ms
		}
		payload.UseProxy.GeoLocation.Country = input.GetCountryCode()
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

		var sessionResp session
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

func (r steelBrowserClient) releaseSession(ctx context.Context, sessionID string) error {
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
