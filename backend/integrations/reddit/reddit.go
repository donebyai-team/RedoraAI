package reddit

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	redditAPIBase  = "https://oauth.reddit.com"
	redditAuthURL  = "https://www.reddit.com/api/v1/authorize"
	redditTokenURL = "https://www.reddit.com/api/v1/access_token"
)

type unAuthorizedErrorCallback func(ctx context.Context)

type Client struct {
	logger                    *zap.Logger
	config                    *models.RedditConfig
	oauthConfig               *oauth2.Config
	httpClient                *retryablehttp.Client
	baseURL                   string
	unAuthorizedErrorCallback unAuthorizedErrorCallback
}

func NewClientWithConfig(config *models.RedditConfig, logger *zap.Logger) *Client {
	return &Client{
		baseURL:    redditAPIBase,
		config:     config,
		logger:     logger,
		httpClient: newHTTPClient(),
	}
}

func newHTTPClient() *retryablehttp.Client {
	cli := retryablehttp.NewClient()
	cli.Logger = nil
	cli.RetryMax = 1

	// Your existing transport
	baseTransport := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DisableKeepAlives:  false,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 300 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Wrap with user-agent injection
	cli.HTTPClient.Transport = baseTransport
	cli.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, err
		}

		// Retry on 429 or 5xx
		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode != 501) {
			return true, nil
		}
		return false, nil
	}

	cli.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			wait := 5 * time.Second // default wait
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					wait = time.Duration(seconds) * time.Second
				}
			}

			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(wait + jitter) // Add random jitter to reduce sync retries
			return nil, fmt.Errorf("rate limited (429), waited %s", wait+jitter)
		}

		return resp, err
	}
	return cli
}
