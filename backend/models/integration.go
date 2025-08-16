package models

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(VOICE_MILLIS, VOICE_VAPI, REDDIT, SLACK_WEBHOOK, REDDIT_DM_LOGIN)
type IntegrationType string

// ENUM(ACTIVE, AUTH_REVOKED, ACCOUNT_SUSPENDED, AUTH_EXPIRED, NOT_ESTABLISHED)
type IntegrationState string

type Integration struct {
	ID              string           `db:"id"`
	OrganizationID  string           `db:"organization_id"`
	ReferenceID     *string          `db:"reference_id"`
	Type            IntegrationType  `db:"type"`
	State           IntegrationState `db:"state"`
	EncryptedConfig string           `db:"encrypted_config"`
	PlainTextConfig string           `db:"plain_text_config"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       *time.Time       `db:"updated_at"`
}

func (i *Integration) GetIntegrationStatus(isOldEnough bool) string {
	switch i.State {
	case IntegrationStateACCOUNTSUSPENDED:
		return "🚫 Account seem to be suspended or banned"
	case IntegrationStateAUTHREVOKED:
		return "🚫 Auth token has been revoked by the user"
	case IntegrationStateAUTHEXPIRED:
		return "🔑 Account auth is expired, please reconnect"
	case IntegrationStateNOTESTABLISHED:
		return "📭 Account is not fully established. Please verify your Reddit email and reconnect"
	}

	if !isOldEnough {
		return "⏳ This Reddit account is less than 2 weeks old. It will be used for auto-comments only after the warmup period."
	}

	return ""
}

type Serializable interface {
	EncryptedData() []byte
	PlainTextData() []byte
	GetUniqID() *string
}

func SetIntegrationType[T Serializable](integration *Integration, integrationType IntegrationType, t T) *Integration {
	integration.Type = integrationType
	integration.EncryptedConfig = string(t.EncryptedData())
	integration.PlainTextConfig = string(t.PlainTextData())
	integration.ReferenceID = t.GetUniqID()
	return integration
}

var _ Serializable = (*VAPIConfig)(nil)

var _ Serializable = (*RedditConfig)(nil)
var _ Serializable = (*RedditDMLoginConfig)(nil)

type RedditDMLoginConfig struct {
	Username          string `json:"username"`
	Alpha2CountryCode string `json:"alpha_2_country_code"`
	Password          string `json:"-"`
	Cookies           string `json:"-"`
}

func (i *RedditDMLoginConfig) GetUniqID() *string {
	if i.Username == "" {
		panic(fmt.Errorf("username cannot be empty"))
	}
	return &i.Username
}

func (i *RedditDMLoginConfig) EncryptedData() []byte {
	toEncrypt := struct {
		Password string `json:"password"`
		Cookies  string `json:"cookies"`
	}{
		Password: i.Password,
		Cookies:  i.Cookies,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *RedditDMLoginConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetRedditDMLoginConfig() *RedditDMLoginConfig {
	if i.Type != IntegrationTypeREDDITDMLOGIN {
		panic(fmt.Errorf("integration is not a reddit dm login integration"))
	}

	out := RedditDMLoginConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}

	encryptedData := struct {
		Password string `json:"password"`
		Cookies  string `json:"cookies"`
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal reddit config: %w", err))
	}
	out.Password = encryptedData.Password
	out.Cookies = encryptedData.Cookies
	return &out
}

type RedditConfig struct {
	AccessToken      string    `json:"-"`
	RefreshToken     string    `json:"-"`
	Verified         bool      `json:"verified"`
	Coins            float64   `json:"coins"`
	Id               string    `json:"id"`
	OauthClientId    string    `json:"oauth_client_id"`
	IsMod            bool      `json:"is_mod"`
	AwarderKarma     float64   `json:"awarder_karma"`
	HasVerifiedEmail bool      `json:"has_verified_email"`
	IsSuspended      bool      `json:"is_suspended"`
	AwardeeKarma     float64   `json:"awardee_karma"`
	LinkKarma        float64   `json:"link_karma"`
	TotalKarma       float64   `json:"total_karma"`
	InboxCount       int       `json:"inbox_count"`
	Name             string    `json:"name"`
	Created          float64   `json:"created"`
	CreatedUtc       float64   `json:"created_utc"`
	CommentKarma     float64   `json:"comment_karma"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func (r *RedditConfig) GetUniqID() *string {
	if r.Name == "" {
		panic(fmt.Errorf("username cannot be empty"))
	}
	return &r.Name
}

// IsUserOldEnough checks if the Reddit user is at least x weeks old.
func (r *RedditConfig) IsUserOldEnough(weeks int) bool {
	// Convert the CreatedUtc (which is in seconds since epoch) to time.Time
	createdTime := time.Unix(int64(r.CreatedUtc), 0)

	// Calculate how many weeks ago the account was created
	duration := time.Since(createdTime)

	// Compare to the threshold duration
	return duration >= (time.Duration(weeks) * 7 * 24 * time.Hour)
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
	Channel string `json:"channel"`
}

func (r *SlackWebhookConfig) GetUniqID() *string {
	return nil
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

func (i *VAPIConfig) GetUniqID() *string {
	return nil
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
