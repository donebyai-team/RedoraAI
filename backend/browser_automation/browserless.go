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
	"strings"
)

type browserLessClient struct {
	Token       string
	WarmUpToken string
	logger      *zap.Logger
}

func NewBrowserLessBrowser(token, warmUpToken string, logger *zap.Logger) BrowserAutomationProvider {
	return &browserLessClient{Token: token, WarmUpToken: warmUpToken, logger: logger}
}

func (b browserLessClient) GetCDPInfo(ctx context.Context, input CDPInput) (*CDPInfo, error) {
	const maxRetries = 3
	//var backoff = 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		var queryBuilder strings.Builder

		queryBuilder.WriteString("mutation {")

		if input.UseProxy {
			queryBuilder.WriteString(`
      proxy(
        type: [document, xhr],
        country: ` + input.GetCountryCode() + `,
        sticky: true
      ) {
        time
      }`)
		}

		queryBuilder.WriteString(`
      goto(
        url: "` + input.StartURL + `",
        waitUntil: firstContentfulPaint
      ) {
        status
      }`)

		if input.LiveURL {
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

		tokenToUse := b.Token
		if input.IsWarmUp {
			tokenToUse = b.WarmUpToken
		}

		resp, err := http.Post(
			fmt.Sprintf("https://production-sfo.browserless.io/chromium/bql?token=%s&humanlike=true&blockConsentModals=true", tokenToUse),
			"application/json",
			bytes.NewBuffer(reqBytes),
		)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		b.logger.Info("browserless raw response",
			zap.String("country_code", input.Alpha2CountryCode),
			zap.ByteString("body", bodyBytes))

		var result reconnectResponse
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}

		// Retry logic if proxy time is 0 and we're using proxy
		//if useProxy && result.Data.Proxy.Time == 0 {
		//	r.logger.Warn("proxy.time is 0, retrying", zap.Int("attempt", attempt+1))
		//	time.Sleep(backoff)
		//	backoff *= 2
		//	continue
		//}

		if result.Data.Reconnect.BrowserWSEndpoint == "" {
			return nil, errors.New("empty browserWSEndpoint - CDP connection failed")
		}

		info := &CDPInfo{
			WSEndpoint: result.Data.Reconnect.BrowserWSEndpoint,
		}

		if input.LiveURL {
			info.LiveURL = result.Data.Live.LiveURL
			b.logger.Info("browserless live url", zap.String("url", info.LiveURL))
		}

		return info, nil
	}

	return nil, errors.New("failed to get CDP URL after retries due to proxy.time = 0")
}

type reconnectResponse struct {
	Data struct {
		Proxy struct {
			Time int `json:"time"`
		} `json:"proxy"`

		Goto struct {
			Status int `json:"status"`
		}

		Reconnect struct {
			BrowserWSEndpoint string `json:"browserWSEndpoint"`
		} `json:"reconnect"`

		Live struct {
			LiveURL string `json:"liveURL"`
		} `json:"live"`
	} `json:"data"`
}
