package models

import (
	"database/sql/driver"
	"time"

	"go.uber.org/zap/zapcore"
)

type Organization struct {
	ID           string                   `db:"id"`
	Name         string                   `db:"name"`
	FeatureFlags OrganizationFeatureFlags `db:"feature_flags"`
	CreatedAt    time.Time                `db:"created_at"`
	UpdatedAt    *time.Time               `db:"updated_at"`
}

func (o *Organization) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", o.ID)
	encoder.AddString("name", o.Name)
	return nil
}

type OrganizationFeatureFlags struct {
	QuoteTTL *time.Duration `json:"quote_ttl,omitempty"`
	// Quote proposals template name
	// - quick_reply
	// - quick_reply_amt
	QuoteProposalEmailTemplate *string  `json:"quote_proposal_email_template,omitempty"`
	DefaultQuoteMarkupPercent  *float32 `json:"default_quote_markup_percent,omitempty"`
	EnableLoadDiffEmail        bool     `json:"enable_load_diff_email"`
	EnableAccessorials         bool     `json:"enable_accessorials"`
	EnableCarrierRates         bool     `json:"enable_carrier_rates"`
}

func (b OrganizationFeatureFlags) IsAccessorialsEnabled() bool {
	return b.EnableAccessorials
}

func (b OrganizationFeatureFlags) CanSendQuoteProposalEmail() bool {
	return b.QuoteProposalEmailTemplate != nil
}

func (b OrganizationFeatureFlags) CanAddAccessorial() bool {
	return b.EnableAccessorials
}

func (b OrganizationFeatureFlags) CanViewCarrierRates() bool {
	return b.EnableCarrierRates
}

func (b OrganizationFeatureFlags) GetDefaultQuoteMarkupPercent() float32 {
	if b.DefaultQuoteMarkupPercent == nil {
		return 0.25
	}
	return *b.DefaultQuoteMarkupPercent
}
func (b OrganizationFeatureFlags) Value() (driver.Value, error) {
	return valueAsJSON(b, "organization feature flags")
}

func (b *OrganizationFeatureFlags) Scan(value interface{}) error {
	return scanFromJSON(value, b, "organization feature flags")
}
