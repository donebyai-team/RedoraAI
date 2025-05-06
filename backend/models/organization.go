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
	EnableAutoComment bool `json:"enable_auto_comment"`
}

func (b OrganizationFeatureFlags) Value() (driver.Value, error) {
	return valueAsJSON(b, "organization feature flags")
}

func (b *OrganizationFeatureFlags) Scan(value interface{}) error {
	return scanFromJSON(value, b, "organization feature flags")
}
