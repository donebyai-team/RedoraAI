package models

import (
	"database/sql/driver"
	"time"
)

type PostInsight struct {
	ID             string              `db:"id"`
	RelevancyScore float64             `db:"relevancy_score"`
	ProjectID      string              `db:"project_id"`
	PostID         string              `db:"post_id"`
	Source         SourceType          `db:"source_type"`
	Topic          string              `db:"topic"`
	Sentiment      string              `db:"sentiment"`
	Highlights     string              `db:"highlights"`
	CreatedAt      time.Time           `db:"created_at"`
	UpdatedAt      *time.Time          `db:"updated_at"`
	Metadata       PostInsightMetadata `db:"metadata"`
}

type PostInsightMetadata struct {
	ChainOfThought      string    `json:"chain_of_thought"`
	HighlightedComments []string  `json:"highlighted_comments"`
	Title               string    `json:"title"`
	PostCreatedAt       time.Time `json:"post_created_at"`
}

func (b PostInsightMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "post_insight metadata")
}

func (b *PostInsightMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "post_insight metadata")
}
