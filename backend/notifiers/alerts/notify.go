package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AlertNotifier interface {
	Send(ctx context.Context, message string) error
}

type SlackNotifier struct {
	WebhookURL string
	Client     *http.Client
}

func NewSlackNotifier(webhookURL string) AlertNotifier {
	return &SlackNotifier{
		WebhookURL: webhookURL,
		Client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackNotifier) Send(ctx context.Context, message string) error {
	payload := map[string]string{
		"text": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create Slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Slack returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
