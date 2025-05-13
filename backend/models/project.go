package models

import (
	"database/sql/driver"
	"time"
)

type Project struct {
	ID                 string          `db:"id"`
	OrganizationID     string          `db:"organization_id"`
	Name               string          `db:"name"`
	ProductDescription string          `db:"description"`
	CustomerPersona    string          `db:"customer_persona"`
	EngagementGoals    string          `db:"goals"`
	WebsiteURL         string          `db:"website"`
	IsActive           bool            `db:"is_active"`
	Metadata           ProjectMetadata `db:"metadata"`
	CreatedAt          time.Time       `db:"created_at"`
	UpdatedAt          *time.Time      `db:"updated_at"`
}

type ProjectMetadata struct {
	SuggestedKeywords   []string `json:"suggested_keywords"`
	SuggestedSubReddits []string `json:"suggested_subreddits"`
}

func (b ProjectMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "project metadata")
}

func (b *ProjectMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "project metadata")
}
