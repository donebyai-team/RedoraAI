package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"time"
)

const (
	redditAPIBase  = "https://oauth.reddit.com"
	redirectURL    = "http://localhost:3000/auth/callback"
	redditAuthURL  = "https://www.reddit.com/api/v1/authorize"
	redditTokenURL = "https://www.reddit.com/api/v1/access_token"
)

type OauthClient struct {
	logger       *zap.Logger
	clientID     string
	clientSecret string
	config       *oauth2.Config
}

func NewRedditOauthClient(clientID, clientSecret string) *OauthClient {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"identity", "read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  redditAuthURL,
			TokenURL: redditTokenURL,
		},
	}

	return &OauthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config,
	}
}

// GetAuthURL returns the authorization URL
func (r *OauthClient) GetAuthURL(state string) string {
	return r.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Authorize exchanges the auth code for access + refresh tokens
func (r *OauthClient) Authorize(ctx context.Context, code string) (*models.RedditConfig, error) {
	token, err := r.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Step 2: Create an authenticated HTTP client
	client := r.config.Client(ctx, token)

	// Step 3: Call Reddit API to get user info
	req, err := http.NewRequestWithContext(ctx, "GET", redditAPIBase+"/api/v1/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reddit API error: %s", string(body))
	}

	// Step 4: Parse JSON response
	var userInfo struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &models.RedditConfig{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		UserName:     userInfo.Name,
	}, nil
}

type Client struct {
	logger *zap.Logger
	config *models.RedditConfig
}

func NewRedditClient(logger *zap.Logger, config *models.RedditConfig) *Client {
	return &Client{logger: logger, config: config}
}

func (r *Client) GetUser(ctx context.Context, userId string) (string, error) {
	panic("implement me")
}

func (r *Client) isTokenExpired() bool {
	// Refresh if within 60 seconds of expiry
	return time.Now().After(r.config.ExpiresAt.Add(-1 * time.Minute))
}
