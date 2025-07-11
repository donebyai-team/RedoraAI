package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrUnAuthorized = errors.New("unauthorized")

func redditAccountSuspendedError(username string) error {
	return fmt.Errorf("Your connected Reddit account [%s] is either suspended or banned, please contact us via chat", username)
}

func (r *Client) doRequest(ctx context.Context, method, url string, rawBody interface{}) (*http.Response, error) {
	req, err := r.buildRequest(ctx, rawBody, method, url)
	if err != nil {
		return nil, err
	}

	resp, err := r.DoWithRateLimit(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if err := validateResponse(resp); err != nil {
		if errors.Is(err, ErrUnAuthorized) && r.unAuthorizedErrorCallback != nil {
			r.unAuthorizedErrorCallback(ctx)
		}

		resp.Body.Close() // make sure caller isn't left with unclosed body
		return nil, err
	}
	return resp, nil
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
