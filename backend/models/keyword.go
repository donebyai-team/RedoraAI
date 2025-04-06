package models

import (
	"time"
)

type Keyword struct {
	ID        string    `db:"id"`
	OrgID     string    `db:"organization_id"`
	Keyword   string    `db:"keyword"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
