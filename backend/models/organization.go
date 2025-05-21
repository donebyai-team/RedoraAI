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

// TODO: Move it to a better place
type OrganizationFeatureFlags struct {
	EnableAutoDM      bool          `json:"enable_auto_dm"`
	EnableAutoComment bool          `json:"enable_auto_comment"`
	CommentLLMModel   LLMModel      `json:"comment_llm_model"`
	DMLLMModel        LLMModel      `json:"dm_llm_model"`
	RelevancyLLMModel LLMModel      `json:"relevancy_llm_model"`
	Subscription      *Subscription `json:"subscription"` // storing here for faster access
	Activities        []OrgActivity `json:"activities"`
}

type OrgActivity struct {
	ActivityType OrgActivityType `json:"activity_type"`
	CreatedAt    time.Time       `json:"created_at"`
}

//go:generate go-enum -f=$GOFILE

// ENUM(COMMENT_DISABLED_ACCOUNT_AGE_NEW, COMMENT_DISABLED_LOW_KARMA, COMMENT_ENABLED_WARMED_UP)
type OrgActivityType string

func (b OrganizationFeatureFlags) IsSubscriptionExpired() bool {
	if b.Subscription == nil {
		return false
	}
	return b.Subscription.IsExpired()
}

func (b OrganizationFeatureFlags) Value() (driver.Value, error) {
	return valueAsJSON(b, "organization feature flags")
}

func (b *OrganizationFeatureFlags) Scan(value interface{}) error {
	return scanFromJSON(value, b, "organization feature flags")
}
