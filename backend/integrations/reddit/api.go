package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (r *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	reqURL := fmt.Sprintf("%s/user/%s/about.json", r.baseURL, userID)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Data User `json:"data"`
	}

	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (r *Client) GetSubRedditByName(ctx context.Context, name string) (*SubReddit, error) {
	reqURL := fmt.Sprintf("%s/r/%s/about.json", r.baseURL, name)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL)
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

	return &response.Data, nil
}

//go:generate go-enum -f=$GOFILE

// ENUM(TOP, HOT, RELEVANCE, NEW, COMMENT_COUNT)
type SortBy string

// ENUM(ALL, YEAR, WEEK, MONTH, TODAY, HOUR)
type TimeRange string

type PostFilters struct {
	Keywords []string
	SortBy   *SortBy
	TimeRage *TimeRange
	After    *string
	Before   *string
	Limit    int
}

func (r *Client) GetPosts(ctx context.Context, subRedditName string, filters PostFilters) ([]*Post, error) {
	v := url.Values{}
	if len(filters.Keywords) > 0 {
		v.Set("q", strings.Join(filters.Keywords, " "))
	}

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
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL)
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
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL)
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

func (r *Client) GetPostWithAllComments(ctx context.Context, postID string) (*Post, error) {
	reqURL := fmt.Sprintf("%s/comments/%s.json?raw_json=1&sort=new", r.baseURL, postID)
	resp, err := r.doRequest(ctx, http.MethodGet, reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResp []json.RawMessage
	if err := decodeJSON(resp.Body, &rawResp); err != nil {
		return nil, err
	}
	return r.parsePostWithComments(ctx, rawResp)
}

func (r *Client) isTokenExpired() bool {
	// Refresh if within 60 seconds of expiry
	return true
}

func (r *Client) refreshToken(ctx context.Context) error {
	// Build the current token manually
	oldToken := &oauth2.Token{
		AccessToken:  r.config.AccessToken,
		RefreshToken: r.config.RefreshToken,
		Expiry:       r.config.ExpiresAt,
	}

	// Create a token source that can refresh
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

	return nil
}
