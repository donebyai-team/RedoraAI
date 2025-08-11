package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrUnAuthorized = errors.New("unauthorized")
var AccountBanned = errors.New("your connected Reddit account either suspended or banned")
var AllAccountBanned = errors.New("Your connected Reddit accounts either suspended or banned")
var AllAccountNotEstablished = errors.New("Your Reddit accounts isn't established yet — it needs things like a verified email, some posting history, and a clean track record to qualify.")

func (r *Client) doRequest(ctx context.Context, method, url string, rawBody interface{}) (*http.Response, error) {
	// Helper to execute and validate request
	execute := func() (*http.Response, error) {
		req, err := r.buildRequest(ctx, rawBody, method, url)
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %w", err)
		}

		resp, err := r.DoWithRateLimit(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		r.logger.Info("request done with user agent", zap.String("user_agent", req.Header.Get("User-Agent")))

		if err := validateResponse(resp); err != nil {
			resp.Body.Close()
			return nil, err
		}
		return resp, nil
	}

	// Initial request attempt
	resp, err := execute()
	if err == nil {
		return resp, nil
	}

	// Check if error is unauthorized and retry logic is applicable
	if errors.Is(err, ErrUnAuthorized) && r.unAuthorizedErrorCallback != nil {
		r.logger.Warn("unauthorized response, attempting token refresh", zap.String("account", r.config.Name))

		if refreshErr := r.refreshToken(ctx); refreshErr != nil {
			r.logger.Error("token refresh failed", zap.String("account", r.config.Name), zap.Error(refreshErr))
			r.unAuthorizedErrorCallback(ctx)
			return nil, err
		}

		// Retry after successful token refresh
		r.logger.Info("retrying request after successful token refresh", zap.String("account", r.config.Name))
		resp, retryErr := execute()
		if retryErr != nil {
			if errors.Is(retryErr, ErrUnAuthorized) {
				r.logger.Error("unauthorized again after token refresh", zap.String("account", r.config.Name))
				r.unAuthorizedErrorCallback(ctx)
			}
			return nil, retryErr
		}
		return resp, nil
	}

	return nil, err
}

func (r *Client) buildRequest(ctx context.Context, rawBody interface{}, method, url string) (*retryablehttp.Request, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, rawBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if r.config != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))
		//req.Header.Set("User-Agent", fmt.Sprintf("com.redoraai:v0.1 by (/u/%s)", r.config.Name))
	} else {
		//req.Header.Set("User-Agent", "com.redoraai:v0.1 by (redora)")
	}

	return req, nil
}

func validateResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnAuthorized
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func decodeJSON(body io.Reader, out any) error {
	if err := json.NewDecoder(body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

var redditLimiter = rate.NewLimiter(rate.Every(500*time.Millisecond), 1) // 2 req/sec

func (r *Client) DoWithRateLimit(req *retryablehttp.Request) (*http.Response, error) {
	err := redditLimiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	return r.httpClient.Do(req)
}
