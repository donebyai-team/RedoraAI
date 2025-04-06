package models

import (
	"database/sql/driver"
	"time"
)

type Keyword struct {
	ID        string    `db:"id"`
	OrgID     string    `db:"organization_id"`
	Keyword   string    `db:"keyword"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Reference - https://developers.reddit.com/docs/api/redditapi/classes/models.Subreddit
type SubReddit struct {
	ID                 string            `db:"id"`
	OrganizationID     string            `db:"organization_id"`
	SubRedditID        string            `db:"subreddit_id"`
	Name               string            `db:"name"`
	Description        string            `db:"description"`
	URL                string            `db:"url"`
	SubredditCreatedAt time.Time         `db:"subreddit_created_at"`
	SubRedditMetadata  SubRedditMetadata `db:"metadata"`

	// Optional
	Subscribers       *int64     `db:"subscribers"`
	Title             *string    `db:"title"`
	LastTrackedAt     *time.Time `db:"last_tracked_at"`
	LastPostCreatedAt *time.Time `db:"last_post_created_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         *time.Time `db:"updated_at"`
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

type AugmentedSubReddit struct {
	SubReddit *SubReddit
	Keywords  []*Keyword
}

//go:generate go-enum -f=$GOFILE

// ENUM(COMMENT, POST)
type RedditLeadType string

type RedditLead struct {
	ID                 string             `db:"id"`
	OrganizationID     string             `db:"organization_id"`
	SubRedditID        string             `db:"subreddit_id"`
	User               string             `db:"user"`
	PostID             string             `db:"post_id"`
	Type               RedditLeadType     `db:"type"`
	RelevancyScore     float64            `db:"relevancy_score"`
	PostCreatedAt      time.Time          `db:"post_created_at"`
	CommentID          *string            `db:"comment_id"`
	Title              *string            `db:"title"` // Optional in case of comment
	Description        *string            `db:"description"`
	RedditLeadMetadata RedditLeadMetadata `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type RedditLeadMetadata struct {
	ChainOfThought                   string `json:"chain_of_thought"`
	SuggestedComment                 string `json:"suggested_comment"`
	SuggestedDM                      string `json:"suggested_dm"`
	ChainOfThoughtSuggestedComment   string `json:"chain_of_thought_suggested_comment"`
	ChainOfThoughtCommentSuggestedDM string `json:"chain_of_thought_comment"`
	NoOfComments                     int    `json:"no_of_comments"`
	NoOfLikes                        int    `json:"no_of_likes"`
}

func (b RedditLeadMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "reddit_lead metadata")
}

func (b *RedditLeadMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "reddit_lead metadata")
}
