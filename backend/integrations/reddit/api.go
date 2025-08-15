package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Client) GetUser(ctx context.Context, userName string) (*User, error) {
	reqURL := fmt.Sprintf("%s/user/%s/about.json", r.baseURL, userName)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, AccountBanned
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Data User `json:"data"`
	}

	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}

	if response.Data.IsSuspended {
		return nil, AccountBanned
	}

	return &response.Data, nil
}

func (r *Client) PostComment(ctx context.Context, thingID, text string) (*Comment, error) {
	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("thing_id", thingID) // e.g. t3_abc123
	form.Set("text", text)

	reqURL := fmt.Sprintf("%s/api/comment", r.baseURL)
	resp, err := r.doRequest(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to post comment: %s", resp.Status)
	}

	// Parse Reddit's response JSON
	var result struct {
		JSON struct {
			Errors [][]interface{} `json:"errors"`
			Data   struct {
				Things []struct {
					Kind string  `json:"kind"`
					Data Comment `json:"data"`
				} `json:"things"`
			} `json:"data"`
		} `json:"json"`
	}

	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("error decoding comment response: %w", err)
	}

	if len(result.JSON.Errors) > 0 {
		return nil, fmt.Errorf("reddit API error: %v", result.JSON.Errors)
	}

	if len(result.JSON.Data.Things) == 0 {
		return nil, errors.New("no comment returned")
	}

	return &result.JSON.Data.Things[0].Data, nil
}

func (r *Client) JoinSubreddit(ctx context.Context, subreddit string) error {
	form := url.Values{}
	form.Set("action", "sub")
	form.Set("sr_name", subreddit)

	reqURL := fmt.Sprintf("%s/api/subscribe", r.baseURL)
	resp, err := r.doRequest(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to join subreddit: %s", resp.Status)
	}
	var result struct {
		JSON struct {
			Errors [][]interface{} `json:"errors"`
		} `json:"json"`
	}

	// Try to decode if content-type is JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := decodeJSON(resp.Body, &result); err != nil {
			return fmt.Errorf("error decoding join response: %w", err)
		}
		if len(result.JSON.Errors) > 0 {
			return fmt.Errorf("reddit API error joining subreddit: %v", result.JSON.Errors)
		}
	}

	// If no errors and no JSON, assume success
	return nil
}

func (r *Client) GetSubRedditByName(ctx context.Context, name string) (*SubReddit, error) {
	if strings.ToLower(strings.TrimSpace(name)) == "all" {
		return &SubReddit{
			DisplayName: "all",
			Description: "all subreddits",
			Title:       "Global",
		}, nil
	}

	reqURL := fmt.Sprintf("%s/r/%s/about.json", r.baseURL, name)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data SubReddit `json:"data"`
	}

	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}

	if response.Data.ID == "" || response.Data.DisplayName == "" {
		return nil, status.Error(codes.NotFound, "subreddit not found")
	}

	// Get RUles
	reqURL = fmt.Sprintf("%s/r/%s/about/rules.json", r.baseURL, name)
	resp, err = r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseRules struct {
		Rules []SubRedditRule `json:"rules"`
	}

	if err := decodeJSON(resp.Body, &responseRules); err != nil {
		return nil, err
	}

	response.Data.Rules = responseRules.Rules

	return &response.Data, nil
}

//go:generate go-enum -f=$GOFILE

// ENUM(TOP, HOT, RELEVANCE, NEW, COMMENT_COUNT, CONFIDENCE)
type SortBy string

// ENUM(ALL, YEAR, WEEK, MONTH, TODAY, HOUR)
type TimeRange string

type QueryFilters struct {
	Keywords    []string
	SortBy      *SortBy
	TimeRage    *TimeRange
	After       *string
	Before      *string
	Limit       int
	MaxComments int
	IncludeMore bool
}

func (r *Client) GetConfig() *models.RedditConfig {
	return r.config
}

func (r *Client) GetPosts(ctx context.Context, subRedditName string, filters QueryFilters) ([]*Post, error) {
	v := url.Values{}
	if len(filters.Keywords) > 0 {
		v.Set("q", strings.Join(filters.Keywords, " "))
	}

	// IMP: make sure the sort by is in lower case else it won't work
	if filters.SortBy != nil {
		v.Set("sort", strings.ToLower(filters.SortBy.String()))
	}

	if filters.TimeRage != nil {
		v.Set("t", strings.ToLower(filters.TimeRage.String()))
	}

	if filters.Limit != 0 {
		v.Set("limit", strconv.Itoa(filters.Limit))
	}

	if filters.After != nil {
		v.Set("after", *filters.After)
	}

	v.Set("restrict_sr", "1")
	reqURL := fmt.Sprintf("%s/r/%s/search.json?%s", r.baseURL, subRedditName, v.Encode())
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Children []struct {
				Data *Post `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}

	var posts []*Post
	for _, child := range response.Data.Children {
		posts = append(posts, child.Data)
	}

	return posts, nil
}

func (r *Client) GetPostByID(ctx context.Context, postID string) (*Post, error) {
	reqURL := fmt.Sprintf("%s/comments/%s.json", r.baseURL, postID)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response []struct {
		Data struct {
			Children []struct {
				Data Post `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}

	if len(response) > 0 && len(response[0].Data.Children) > 0 {
		return &response[0].Data.Children[0].Data, nil
	}

	return nil, ErrNotFound // Post not found in the response
}

func (r *Client) CreatePost(ctx context.Context, subreddit string, post *models.Post) (*Post, error) {
	title := post.Title
	text := post.Description

	form := url.Values{}
	form.Set("sr", subreddit) // Subreddit name
	form.Set("title", title)  // Post title
	form.Set("text", text)    // Post body

	flairID := post.Metadata.Settings.FlairID
	if flairID != nil {
		form.Set("flair_id", *flairID)
	}

	form.Set("kind", "self")        // "self" for text post, "link" for link post
	form.Set("resubmit", "true")    // avoid Reddit duplicate filtering
	form.Set("sendreplies", "true") // enable inbox replies
	form.Set("api_type", "json")

	reqURL := fmt.Sprintf("%s/api/submit", r.baseURL)
	resp, err := r.doRequest(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to submit post: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		JSON struct {
			Errors [][]interface{} `json:"errors"`
			Data   struct {
				URL   string `json:"url"`
				Name  string `json:"name"`  // Fullname of the new post (e.g., t3_abcdef)
				ID    string `json:"id"`    // ID of the new post
				Draft bool   `json:"draft"` // Whether post is a draft
			} `json:"data"`
		} `json:"json"`
	}

	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create post response: %w", err)
	}

	if len(result.JSON.Errors) > 0 {
		return nil, fmt.Errorf("reddit API error while posting: %v", result.JSON.Errors)
	}

	return &Post{
		ID:       result.JSON.Data.ID,
		URL:      result.JSON.Data.URL,
		Title:    title,
		Selftext: text,
	}, nil
}

func (r *Client) GetPostWithAllComments(ctx context.Context, postID string, filters QueryFilters) (*Post, error) {
	v := url.Values{}
	// IMP: make sure the sort by is in lower case else it won't work
	if filters.SortBy != nil {
		v.Set("sort", strings.ToLower(filters.SortBy.String()))
	}

	if filters.TimeRage != nil {
		v.Set("t", strings.ToLower(filters.TimeRage.String()))
	}

	if filters.Limit != 0 {
		v.Set("limit", strconv.Itoa(filters.Limit))
	}

	if filters.After != nil {
		v.Set("after", *filters.After)
	}

	v.Set("raw_json", "1")

	reqURL := fmt.Sprintf("%s/comments/%s.json?%s", r.baseURL, postID, v.Encode())
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResp []json.RawMessage
	if err := decodeJSON(resp.Body, &rawResp); err != nil {
		return nil, err
	}
	return r.parsePostWithComments(ctx, rawResp, filters.MaxComments, filters.IncludeMore)
}

func (r *Client) isTokenExpired() bool {
	const buffer = 10 * time.Minute
	return time.Now().Add(buffer).After(r.config.ExpiresAt)
}

func (r *Client) refreshToken(ctx context.Context) error {
	// Build the current token manually
	oldToken := &oauth2.Token{
		AccessToken:  r.config.AccessToken,
		RefreshToken: r.config.RefreshToken,
		Expiry:       r.config.ExpiresAt,
	}
	// Create a token source that can refresh
	ctx = context.WithValue(ctx, oauth2.HTTPClient, r.httpClient.HTTPClient)
	tokenSource := r.oauthConfig.TokenSource(ctx, oldToken)

	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update the config with new token details
	r.config.AccessToken = newToken.AccessToken
	r.config.ExpiresAt = newToken.Expiry

	// Only update refresh token if it's provided
	if newToken.RefreshToken != "" {
		r.config.RefreshToken = newToken.RefreshToken
	}

	r.logger.Info("token refreshed", zap.String("expiry", r.config.ExpiresAt.String()), zap.String("account", r.config.Name))

	return nil
}

func (r *Client) GetPostRequirements(ctx context.Context, subreddit string) (*models.PostRequirements, error) {
	reqURL := fmt.Sprintf("%s/api/v1/%s/post_requirements", r.baseURL, subreddit)

	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var requirements models.PostRequirements
	if err := json.Unmarshal(bodyBytes, &requirements); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &requirements, nil
}

func (r *Client) GetSubredditFlairs(ctx context.Context, subreddit string) ([]models.Flair, error) {
	reqURL := fmt.Sprintf("%s/r/%s/api/link_flair_v2", r.baseURL, subreddit)

	resp, err := r.doRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var flairs []models.Flair
	if err := json.Unmarshal(bodyBytes, &flairs); err != nil {
		return nil, fmt.Errorf("failed to decode flairs: %w", err)
	}

	return flairs, nil
}
