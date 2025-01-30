package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func (s *Client) CreateCall(ctx context.Context, req *RegisterCallRequest) (*RegisterCallResponse, error) {
	resp, err := s.post(ctx, "/start_outbound_call", req)
	if err != nil {
		return nil, fmt.Errorf("start_outbound_call: %w", err)
	}

	cnt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	out := RegisterCallResponse{}
	if err := json.Unmarshal(cnt, &out); err != nil {
		return nil, fmt.Errorf("decode start_outbound_call response: %w", err)
	}
	return &out, nil
}
