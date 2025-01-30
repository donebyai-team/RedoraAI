package api

import (
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

type Client struct {
	cli       *retryablehttp.Client
	hostname  string
	token     string
	expiresAt time.Time
	isSandbox bool
	logger    *zap.Logger
}

func (c *Client) IsExpired() bool {
	return time.Now().After(c.expiresAt)
}

func newHTTPClient() *retryablehttp.Client {
	cli := retryablehttp.NewClient()
	cli.Logger = nil
	cli.RetryMax = 1
	cli.HTTPClient.Transport = &http.Transport{
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

	cli.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		return resp, err
	}

	return cli
}

type MillisConfig struct {
	Hostname string `json:"hostname"`
	Token    string `json:"token"`
}

func NewClient(config *MillisConfig, logger *zap.Logger) (*Client, error) {
	client := &Client{
		cli:      newHTTPClient(),
		hostname: config.Hostname,
		token:    config.Token,
		logger:   logger,
	}

	return client, nil
}
