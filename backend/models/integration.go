package models

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(VOICE_MILLIS, VOICE_VAPI, REDDIT, SLACK_WEBHOOK)
type IntegrationType string

// ENUM(ACTIVE, AUTH_REVOKED)
type IntegrationState string

type Integration struct {
	ID              string           `db:"id"`
	OrganizationID  string           `db:"organization_id"`
	Type            IntegrationType  `db:"type"`
	State           IntegrationState `db:"state"`
	EncryptedConfig string           `db:"encrypted_config"`
	PlainTextConfig string           `db:"plain_text_config"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       *time.Time       `db:"updated_at"`
}

type Serializable interface {
	EncryptedData() []byte
	PlainTextData() []byte
}

func SetIntegrationType[T Serializable](integration *Integration, integrationType IntegrationType, t T) *Integration {
	integration.Type = integrationType
	integration.EncryptedConfig = string(t.EncryptedData())
	integration.PlainTextConfig = string(t.PlainTextData())
	return integration
}

var _ Serializable = (*VAPIConfig)(nil)

var _ Serializable = (*RedditConfig)(nil)

type RedditConfig struct {
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	UserName     string    `json:"user_name"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (i *RedditConfig) EncryptedData() []byte {
	toEncrypt := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  i.AccessToken,
		RefreshToken: i.RefreshToken,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *RedditConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetRedditConfig() *RedditConfig {
	if i.Type != IntegrationTypeREDDIT {
		panic(fmt.Errorf("integration is not a reddit integration"))
	}

	out := RedditConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}

	encryptedData := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}
	out.AccessToken = encryptedData.AccessToken
	out.RefreshToken = encryptedData.RefreshToken
	return &out
}

type SlackWebhookConfig struct {
	Webhook string `json:"-"`
}

func (i *SlackWebhookConfig) EncryptedData() []byte {
	toEncrypt := struct {
		Webhook string `json:"webhook"`
	}{
		Webhook: i.Webhook,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *SlackWebhookConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetSlackWebhook() *SlackWebhookConfig {
	if i.Type != IntegrationTypeSLACKWEBHOOK {
		panic(fmt.Errorf("integration is not a slack webhook integration"))
	}

	out := SlackWebhookConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}

	encryptedData := struct {
		Webhook string `json:"webhook"`
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}
	out.Webhook = encryptedData.Webhook
	return &out
}

type VAPIConfig struct {
	APIKey   string `json:"-"`
	HostName string `json:"hostname"`
}

func NewVAPIConfig(apiKey string) *VAPIConfig {
	return &VAPIConfig{
		APIKey: apiKey,
	}
}

func (i *VAPIConfig) EncryptedData() []byte {
	toEncrypt := struct {
		APIKey string `json:"api_key"`
	}{
		APIKey: i.APIKey,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *VAPIConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetVAPIConfig() *VAPIConfig {
	if i.Type != IntegrationTypeVOICEVAPI {
		panic(fmt.Errorf("integration is not a vapi integration"))
	}

	out := VAPIConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal vapi config: %w", err))
	}

	encryptedData := struct {
		APIKey string `json:"api_key"`
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal vapi config: %w", err))
	}
	out.APIKey = encryptedData.APIKey
	return &out
}
