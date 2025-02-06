package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"net/http"

	"go.uber.org/zap"
)

func (s *Client) post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return s.doRequest(ctx, "POST", path, body)
}

func (s *Client) patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return s.doRequest(ctx, "PATCH", path, body)
}

// Get performs a GET request
func (s *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return s.doRequest(ctx, "GET", path, nil)
}

// doRequest handles the common HTTP request setup and execution
func (s *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	// Create the full URL
	url := "https://" + s.hostname + path

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}

	// Create the request
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))

	s.logger.Debug("performing request", zap.Stringp("url", &url), zap.String("method", method), zap.String("body", string(reqBody)))
	// Perform the request
	resp, err := s.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}

	if resp.StatusCode < 300 {
		return resp, nil
	}

	//errRes := []types.ErrorResponse{}
	//if err := json.NewDecoder(resp.Body).Decode(&errRes); err != nil {
	//	return nil, fmt.Errorf("decode error response: %w", err)
	//}
	//
	//if len(errRes) > 0 {
	//	return nil, &errRes[0]
	//}

	return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
}
