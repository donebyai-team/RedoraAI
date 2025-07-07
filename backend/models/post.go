package models

import (
	"database/sql/driver"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(CREATED, PROCESSING, SENT, FAILED, SCHEDULED)
type PostStatus string

type Post struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	SourceID    string       `db:"source_id"`
	ReferenceID *string      `db:"reference_id"`
	Status      PostStatus   `db:"status"`
	Reason      string       `db:"reason"`
	ScheduleAt  *time.Time   `db:"schedule_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   *time.Time   `db:"updated_at"`
	Metadata    PostMetadata `db:"metadata"`
}

type AugmentedPost struct {
	ID          string                `db:"id"`
	ProjectID   string                `db:"project_id"`
	Title       string                `db:"title"`
	Description string                `db:"description"`
	SourceID    string                `db:"source_id"`
	Source      Source                `db:"source"`
	Status      LeadInteractionStatus `db:"status"`
	Reason      string                `db:"reason"`
	ReferenceID *string               `db:"reference_id"` // id of the insight
	ScheduleAt  *time.Time            `db:"schedule_at"`
	CreatedAt   time.Time             `db:"created_at"`
	UpdatedAt   *time.Time            `db:"updated_at"`
	Metadata    PostMetadata          `db:"metadata"`
}

type PostMetadata struct {
	Settings PostSettings              `json:"settings"`
	History  []PostRegenerationHistory `json:"history"`
}

type PostRegenerationHistory struct {
	PostSettings PostSettings `json:"post_settings"`
	Title        string       `db:"title"`
	Description  string       `db:"description"`
}

func (b PostMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "post metadata")
}

func (b *PostMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "post metadata")
}

type PostSettings struct {
	Topic       string  `json:"topic"`
	Context     string  `json:"context"`
	Goal        string  `json:"goal"`
	Tone        string  `json:"tone"`
	ReferenceID *string `json:"reference_id"`
}
