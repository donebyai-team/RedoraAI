package models

import (
	"database/sql/driver"
	"time"
)

type Keyword struct {
	ID        string     `db:"id"`
	ProjectID string     `db:"project_id"`
	Keyword   string     `db:"keyword"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type SubRedditTracker struct {
	ID                string     `db:"id"`
	SubRedditID       string     `db:"subreddit_id"`
	KeywordID         string     `db:"keyword_id"`
	LastTrackedAt     *time.Time `db:"last_tracked_at"`
	NewestTrackedPost *string    `db:"newest_tracked_post"`
	OldestTrackedPost *string    `db:"oldest_tracked_post"`
}

// ENUM(SUBREDDIT)
type SourceType string

// Reference - https://developers.reddit.com/docs/api/redditapi/classes/models.Subreddit
type Source struct {
	ID          string            `db:"id"`
	ProjectID   string            `db:"project_id"`
	ExternalID  *string           `db:"external_id"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
	SourceType  SourceType        `db:"source_type"`
	DeletedAt   *time.Time        `db:"deleted_at"`
	Metadata    SubRedditMetadata `db:"metadata"`

	// Optional
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// Store fields required to show in UI
// Eg. Guidelines, rules, karma points etc
type SubRedditMetadata struct {
	Title     *string   `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

func (b SubRedditMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "subreddit metadata")
}

func (b *SubRedditMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "subreddit metadata")
}

//type AugmentedSubRedditTracker struct {
//	Tracker   *SubRedditTracker
//	Source *Source
//	Keyword   *Keyword
//	Project   *Project
//}

type AugmentedKeywordTracker struct {
	Tracker *KeywordTracker
	Source  *Source
	Keyword *Keyword
	Project *Project
}

func (a *AugmentedKeywordTracker) GetID() string {
	return a.Tracker.ID
}

type KeywordTracker struct {
	ID            string                 `db:"id"`
	SourceID      string                 `db:"source_id"`
	KeywordID     string                 `db:"keyword_id"`
	Metadata      KeywordTrackerMetadata `db:"metadata"`
	LastTrackedAt *time.Time             `db:"last_tracked_at"`
	CreatedAt     time.Time              `db:"created_at"`
	UpdatedAt     *time.Time             `db:"updated_at"`
	DeletedAt     *time.Time             `db:"deleted_at"`
}

type KeywordTrackerMetadata struct {
}

func (b KeywordTrackerMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "tracker metadata")
}

func (b *KeywordTrackerMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "tracker metadata")
}

//go:generate go-enum -f=$GOFILE

// ENUM(COMMENT, POST)
type LeadType string

// ENUM(NEW, COMPLETED, NOT_RELEVANT)
type LeadStatus string

type Lead struct {
	ID             string       `db:"id"`
	ProjectID      string       `db:"project_id"`
	SourceID       string       `db:"source_id"`
	KeywordID      string       `db:"keyword_id"`
	Author         string       `db:"author"`
	PostID         string       `db:"post_id"`
	Type           LeadType     `db:"type"`
	Status         LeadStatus   `db:"status"`
	RelevancyScore float64      `db:"relevancy_score"`
	PostCreatedAt  time.Time    `db:"post_created_at"`
	CommentID      *string      `db:"comment_id"`
	Title          *string      `db:"title"` // Optional in case of comment
	Description    string       `db:"description"`
	LeadMetadata   LeadMetadata `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type LeadsData struct {
	Count uint32 `db:"count"`
}

type LeadMetadata struct {
	ChainOfThought                 string `json:"chain_of_thought"`
	SuggestedComment               string `json:"suggested_comment"`
	SuggestedDM                    string `json:"suggested_dm"`
	ChainOfThoughtSuggestedComment string `json:"chain_of_thought_suggested_comment"`
	ChainOfThoughtSuggestedDM      string `json:"chain_of_thought_dm"`
	Ups                            int64  `json:"ups"`
	NoOfComments                   int64  `json:"no_of_comments"`
	PostURL                        string `json:"post_url"`
	AuthorURL                      string `json:"author_url"`
	DmURL                          string `json:"dm_url"`
	SelfTextHTML                   string `json:"description_html"`
	SubRedditPrefixed              string `json:"subreddit_prefixed"`
}

func (b LeadMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "reddit_lead metadata")
}

func (b *LeadMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "reddit_lead metadata")
}
