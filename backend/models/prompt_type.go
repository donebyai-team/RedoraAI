package models

import (
	"encoding/json"
	"time"
)

type PromptType struct {
	Name           string          `db:"name"`
	Description    string          `db:"description"`
	OrganizationId string          `db:"organization_id"`
	Config         json.RawMessage `db:"config"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      *time.Time      `db:"updated_at"`
}
