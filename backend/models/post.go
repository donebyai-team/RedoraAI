package models

import (
	"database/sql/driver"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(CREATED, SENT, FAILED, SCHEDULED)
type PostStatus string

type Post struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	SourceID    string       `db:"source_id"`
	ReferenceID *string      `db:"reference_id"`
	PostID      *string      `db:"post_id"`
	Status      PostStatus   `db:"status"`
	Reason      string       `db:"reason"`
	ScheduleAt  *time.Time   `db:"schedule_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   *time.Time   `db:"updated_at"`
	DeletedAt   *time.Time   `db:"deleted_at"`
	Metadata    PostMetadata `db:"metadata"`
}

type AugmentedPost struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	SourceID    string       `db:"source_id"`
	Source      Source       `db:"source"`
	Status      PostStatus   `db:"status"`
	Reason      string       `db:"reason"`
	ReferenceID *string      `db:"reference_id"` // id of the insight
	PostID      *string      `db:"post_id"`
	ScheduleAt  *time.Time   `db:"schedule_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   *time.Time   `db:"updated_at"`
	DeletedAt   *time.Time   `db:"deleted_at"`
	Metadata    PostMetadata `db:"metadata"`
}

type PostMetadata struct {
	Author   string                    `json:"author"`
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
	FlairID     *string `json:"flair_id"`
}
