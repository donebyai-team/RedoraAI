package external_integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type brevo struct {
	ApiKey string `json:"api_key"`
}

func NewBrevoIntegration(apiKey string) *brevo {
	return &brevo{ApiKey: apiKey}
}

func (b *brevo) CreateContact(contact *Contact) error {
	url := "https://api.brevo.com/v3/contacts"

	// Build the payload
	payload := map[string]interface{}{
		"email":         contact.Email,
		"updateEnabled": true,
		"ext_id":        contact.UserID,
		"attributes": map[string]interface{}{
			"PRODUCT_NAME":         contact.ProductName,
			"SUBSCRIPTION_PLAN":    string(contact.SubscriptionPlan),
			"SUBSCRIPTION_EXPIRED": contact.SubscriptionExpired,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal contact payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", b.ApiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo contact creation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo: unexpected status %d - response: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (b *brevo) UpdateContact(contact *Contact) error {
	url := fmt.Sprintf("https://api.brevo.com/v3/contacts/identifier?identifierType=%s", contact.UserID)

	// Build the payload
	payload := map[string]interface{}{
		"EMAIL":                contact.Email,
		"PRODUCT_NAME":         contact.ProductName,
		"SUBSCRIPTION_PLAN":    string(contact.SubscriptionPlan),
		"SUBSCRIPTION_EXPIRED": contact.SubscriptionExpired,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal contact payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", b.ApiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo contact creation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo: unexpected status %d - response: %s", resp.StatusCode, string(body))
	}

	return nil
}
