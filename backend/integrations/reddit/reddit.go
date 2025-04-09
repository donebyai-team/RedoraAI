package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/datastore"
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
	db           datastore.Repository
}

func NewRedditOauthClient(logger *zap.Logger, db datastore.Repository, clientID, clientSecret string) *OauthClient {
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
		logger:       logger,
		db:           db,
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

func (r *OauthClient) NewRedditClient(ctx context.Context, orgID string) (*Client, error) {
	integration, err := r.db.GetIntegrationByOrgAndType(ctx, orgID, models.IntegrationTypeREDDIT)
	if err != nil {
		return nil, err
	}

	client := &Client{logger: r.logger, config: integration.GetRedditConfig()}
	if client.isTokenExpired() {
		err := client.refreshToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to reddit refresh token: %w", err)
		}

		// Update the credentials
		integrationType := models.SetIntegrationType(integration, models.IntegrationTypeREDDIT, client.config)
		integration, err = r.db.UpsertIntegration(ctx, integrationType)
		if err != nil {
			return nil, fmt.Errorf("upsert integration: %w", err)
		}
	}
	return client, nil
}

type Client struct {
	logger      *zap.Logger
	config      *models.RedditConfig
	db          datastore.Repository
	oauthConfig *oauth2.Config
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

type SubReddit struct {
	// Add fields that subreddit api will returns
}

type Post struct {
	// Add fields that subreddit posts api will returns
}

type User struct {
	// add user related data eg. Karma points, name, joined at etc
}

func (r *Client) GetUser(ctx context.Context, userId string) (*User, error) {
	panic("implement me")
}

func (r *Client) GetSubRedditByURL(ctx context.Context, url string) (*SubReddit, error) {
	panic("implement me")
}

func (r *Client) GetPosts(ctx context.Context, subRedditID string, keywords []string) ([]Post, error) {
	panic("implement me")
}

func (r *Client) GetPostByID(ctx context.Context, postID string) (*Post, error) {
	panic("implement me")
}

func (r *Client) isTokenExpired() bool {
	// Refresh if within 60 seconds of expiry
	return time.Now().After(r.config.ExpiresAt.Add(-1 * time.Minute))
}
