package models

import "time"

type Keyword struct {
	ID        string    `db:"id"`
	OrgID     string    `db:"organization_id"`
	Keyword   string    `db:"keyword"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Reference - https://developers.reddit.com/docs/api/redditapi/classes/models.Subreddit
type SubReddit struct {
	ID                 string    `db:"id"`
	OrganizationID     string    `db:"organization_id"`
	SubRedditID        string    `db:"subreddit_id"`
	Name               string    `db:"name"`
	Description        string    `db:"description"`
	URL                string    `db:"url"`
	SubredditCreatedAt time.Time `db:"subreddit_created_at"`

	// Optional
	Subscribers   *int64     `db:"subscribers"`
	Title         *string    `db:"title"`
	LastTrackedAt *time.Time `db:"last_tracked_at"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type AugmentedSubReddit struct {
	SubReddit *SubReddit
	Keywords  []*Keyword
}
