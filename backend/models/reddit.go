package models

import "time"

type SubReddit struct {
	ID             string     `db:"id"`
	OrganizationID string     `db:"organization_id"`
	SubRedditID    string     `db:"subreddit_id"`
	LastTrackedAt  *time.Time `db:"last_tracked_at"`
	Name           string     `db:"from_phone"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}

type AugmentedSubReddit struct {
	SubReddit *SubReddit
	// keywords
}
