package reddit

import (
	"regexp"
	"strings"
)

// SubReddit represents information about a subreddit.
type SubReddit struct {
	ID          string  `json:"id"`           // subreddit id eg. 2qib3
	URL         string  `json:"url"`          // eg. r/sales
	DisplayName string  `json:"display_name"` // subreddit name eg. sales
	Description string  `json:"public_description"`
	CreatedAt   float64 `json:"created_utc"`
	Subscribers int64   `json:"subscribers"`
	Title       string  `json:"title"`
	Over18      bool    `json:"over_18"`
	Rules       []SubRedditRule
	// Add other relevant fields from the subreddit API response
}

type SubRedditRule struct {
	Description string `json:"description"`
	ShortName   string `json:"short_name"`
}

// Post represents a Reddit post.
type Post struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Author            string  `json:"author"`
	AuthorFullName    string  `json:"author_fullname"`
	Score             int     `json:"score"`
	Ups               int64   `json:"ups"`   // Number of upvotes
	Downs             int64   `json:"downs"` // Number of downvotes (usually not directly exposed in v1 API)
	URL               string  `json:"url"`
	Permalink         string  `json:"permalink"`
	CreatedAt         float64 `json:"created_utc"`
	NumComments       int64   `json:"num_comments"`
	Selftext          string  `json:"selftext"`
	SelftextHTML      string  `json:"selftext_html"`
	SubRedditPrefixed string  `json:"subreddit_name_prefixed"`
	SubRedditType     string  `json:"subreddit_type"`
	IsSelf            bool    `json:"is_self"`
	Subreddit         string  `json:"subreddit"`
	Archived          bool    `json:"archived"`
	AuthorInfo        *User
	Comments          []*Comment
	// Add other relevant fields from the post API response
}

// Comment represents a Reddit comment.
type Comment struct {
	ID         string  `json:"id"`
	Author     string  `json:"author"`
	Body       string  `json:"body"`
	Permalink  string  `json:"permalink"`
	CreatedAt  float64 `json:"created_utc"`
	Score      int     `json:"score"`
	Ups        int     `json:"ups"`   // Number of upvotes
	Downs      int     `json:"downs"` // Number of downvotes (usually not directly exposed in v1 API)
	ParentID   string  `json:"parent_id"`
	Depth      int     `json:"depth"`
	Comments   []*Comment
	AuthorInfo *User
	// Add other relevant comment fields
}

var lowSignalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*$`),
	regexp.MustCompile(`(?i)^thanks\b`),
	regexp.MustCompile(`(?i)^lol$`),
	regexp.MustCompile(`^👍+$`),
}

const (
	MinScore  = 1
	MinLength = 30
)

func (f Comment) ShouldInclude() bool {
	if f.Score < MinScore {
		return false
	}
	if len(f.Body) < MinLength {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(f.Body))
	for _, re := range lowSignalPatterns {
		if re.MatchString(lower) {
			return false
		}
	}
	return true
}

// User represents a Reddit user.
type User struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	IsSuspended      bool    `json:"is_suspended"`
	Karma            int     `json:"total_karma"`
	CreatedAt        float64 `json:"created_utc"`
	IsGold           bool    `json:"is_gold"`
	HasVerifiedEmail bool    `json:"has_verified_email"`
	// Add other relevant user-related fields
}

const DummyFlair = "FLAIR_NOT_SPECIFIED"
