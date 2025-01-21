package models

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(MICROSOFT, GOOGLE, DAT, REVENOVA)
type IntegrationType string

// ENUM(ACTIVE)
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

var _ Serializable = (*MicrosoftConfig)(nil)

type MicrosoftConfig struct {
	TenantId string `json:"tenant_id"`
}

func NewMicrosoftConfig(tenantId string) *MicrosoftConfig {
	return &MicrosoftConfig{
		TenantId: tenantId,
	}
}

func (i *MicrosoftConfig) EncryptedData() []byte {
	return []byte("{}")
}

func (i *MicrosoftConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetMicrosoftConfig() *MicrosoftConfig {
	if i.Type != IntegrationTypeMICROSOFT {
		panic(fmt.Errorf("integration is not a microsoft integration"))
	}
	out := MicrosoftConfig{}
	// we know that msoft the whole object is store in plaintext nothing to encrypt?
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal microsoft config: %w", err))
	}
	return &out
}

var _ Serializable = (*GoogleConfig)(nil)

type GoogleConfig struct {
	RefreshToken string `json:"refresh_token"`
}

func NewGoogleConfig(refreshToken string) *GoogleConfig {
	return &GoogleConfig{
		RefreshToken: refreshToken,
	}
}

func (i *GoogleConfig) EncryptedData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *GoogleConfig) PlainTextData() []byte {
	return []byte("{}")
}

func (i *Integration) GetGoogleConfig() *GoogleConfig {
	if i.Type != IntegrationTypeGOOGLE {
		panic(fmt.Errorf("integration is not a google integration"))
	}
	out := GoogleConfig{}
	if err := json.Unmarshal([]byte(i.EncryptedConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal google config: %w", err))
	}
	return &out
}

type RevenovaConfig struct {
	OrgID    string `json:"org_id"`
	LoginURL string `json:"login_url"`
	Username string `json:"username"`
	Password string `json:"-"`
}

func (i *RevenovaConfig) EncryptedData() []byte {
	toEncrypt := struct {
		Password string `json:"password"` // CHANGE_PASSWORD [ENCRYPTED]
	}{
		Password: i.Password,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *RevenovaConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetRevenovaConfig() *RevenovaConfig {
	if i.Type != IntegrationTypeREVENOVA {
		panic(fmt.Errorf("integration is not a revenova integration"))
	}

	out := RevenovaConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal revenova config: %w", err))
	}

	encryptedData := struct {
		Password string `json:"password"` // CHANGE_PASSWORD [ENCRYPTED]
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal google config: %w", err))
	}
	out.Password = encryptedData.Password
	return &out
}

type DATConfig struct {
	AuthHost    string `json:"auth_host"` // identity.api.staging.dat.com
	APIHost     string `json:"api_host"`  // api.staging.dat.com
	OrgUser     string `json:"org_user"`  //dat@loadlogic.ai
	OrgPassword string `json:"-"`
	UserAccount string `json:"user_account"` // mdm@streamingfast.io
}

func (i *DATConfig) EncryptedData() []byte {
	toEncrypt := struct {
		OrgPassword string `json:"org_password"` // CHANGE_PASSWORD [ENCRYPTED]
	}{
		OrgPassword: i.OrgPassword,
	}
	data, err := json.Marshal(toEncrypt)
	if err != nil {
		panic(err)
	}
	return data

}

func (i *DATConfig) PlainTextData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return data
}

func (i *Integration) GetDATConfig() *DATConfig {
	if i.Type != IntegrationTypeDAT {
		panic(fmt.Errorf("integration is not a dat integration"))
	}

	out := DATConfig{}
	if err := json.Unmarshal([]byte(i.PlainTextConfig), &out); err != nil {
		panic(fmt.Errorf("unable to unmarshal dat config: %w", err))
	}

	encryptedData := struct {
		OrgPassword string `json:"org_password"` // CHANGE_PASSWORD [ENCRYPTED]
	}{}

	if err := json.Unmarshal([]byte(i.EncryptedConfig), &encryptedData); err != nil {
		panic(fmt.Errorf("unable to unmarshal google config: %w", err))
	}
	out.OrgPassword = encryptedData.OrgPassword
	return &out
}
