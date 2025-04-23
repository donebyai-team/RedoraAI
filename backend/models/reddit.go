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

// Reference - https://developers.reddit.com/docs/api/redditapi/classes/models.Subreddit
type SubReddit struct {
	ID                 string            `db:"id"`
	ProjectID          string            `db:"project_id"`
	SubRedditID        string            `db:"subreddit_id"`
	Name               string            `db:"name"`
	Description        string            `db:"description"`
	SubredditCreatedAt time.Time         `db:"subreddit_created_at"`
	SubRedditMetadata  SubRedditMetadata `db:"metadata"`
	LastTrackedAt      *time.Time        `db:"last_tracked_at"`
	DeletedAt          *time.Time        `db:"deleted_at"`

	// Optional
	Title     *string    `db:"title"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// Store fields required to show in UI
// Eg. Guidelines, rules, karma points etc
type SubRedditMetadata struct {
}

func (b SubRedditMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "subreddit metadata")
}

func (b *SubRedditMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "subreddit metadata")
}

//type AugmentedSubRedditTracker struct {
//	Tracker   *SubRedditTracker
//	SubReddit *SubReddit
//	Keyword   *Keyword
//	Project   *Project
//}

type AugmentedSubReddit struct {
	SubReddit *SubReddit
	Keywords  []*Keyword
	Project   *Project
}

//go:generate go-enum -f=$GOFILE

// ENUM(COMMENT, POST)
type LeadType string

// ENUM(NEW, COMPLETED, NOT_RELEVANT)
type LeadStatus string

type RedditLead struct {
	ID             string       `db:"id"`
	ProjectID      string       `db:"project_id"`
	SubRedditID    string       `db:"subreddit_id"`
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

type LeadMetadata struct {
	IsRelevant                       bool        `json:"is_relevant"`
	ChainOfThought                   string      `json:"chain_of_thought"`
	SuggestedComment                 string      `json:"suggested_comment"`
	SuggestedDM                      string      `json:"suggested_dm"`
	ChainOfThoughtSuggestedComment   string      `json:"chain_of_thought_suggested_comment"`
	ChainOfThoughtCommentSuggestedDM string      `json:"chain_of_thought_comment"`
	PostURL                          string      `json:"post_url"`
	AuthorInfo                       interface{} `json:"author_info"`
}

func (b LeadMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "reddit_lead metadata")
}

func (b *LeadMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "reddit_lead metadata")
}
