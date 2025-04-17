package reddit

// SubReddit represents information about a subreddit.
type SubReddit struct {
	ID          string  `json:"id"` // subreddit id
	URL         string  `json:"url"`
	DisplayName string  `json:"display_name_prefixed"` // subreddit name
	Description string  `json:"description"`
	CreatedAt   float64 `json:"created"`
	Subscribers int64   `json:"subscribers"`
	Title       string  `json:"title"`
	Over18      bool    `json:"over_18"`
	// Add other relevant fields from the subreddit API response
}

// Post represents a Reddit post.
type Post struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Score       int     `json:"score"`
	Ups         int     `json:"ups"`   // Number of upvotes
	Downs       int     `json:"downs"` // Number of downvotes (usually not directly exposed in v1 API)
	URL         string  `json:"url"`
	Permalink   string  `json:"permalink"`
	CreatedAt   float64 `json:"created_utc"`
	NumComments int     `json:"num_comments"`
	Selftext    string  `json:"selftext"`
	IsSelf      bool    `json:"is_self"`
	Subreddit   string  `json:"subreddit"`
	AuthorInfo  *User
	Comments    []*Comment
	// Add other relevant fields from the post API response
}

// Comment represents a Reddit comment.
type Comment struct {
	ID        string  `json:"id"`
	Author    string  `json:"author"`
	Body      string  `json:"body"`
	Permalink string  `json:"permalink"`
	CreatedAt float64 `json:"created_utc"`
	Score     int     `json:"score"`
	Ups       int     `json:"ups"`   // Number of upvotes
	Downs     int     `json:"downs"` // Number of downvotes (usually not directly exposed in v1 API)
	ParentID  string  `json:"parent_id"`
	Depth     int     `json:"depth"`
	Comments  []*Comment
	// Add other relevant comment fields
}

// User represents a Reddit user.
type User struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Karma            int     `json:"total_karma"`
	CreatedAt        float64 `json:"created_utc"`
	IsGold           bool    `json:"is_gold"`
	HasVerifiedEmail bool    `json:"has_verified_email"`
	// Add other relevant user-related fields
}
