package reddit

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
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
func (r *OauthClient) GetAuthURL(state string, redirectURI string) string {
	r.config.RedirectURL = redirectURI
	return r.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Authorize exchanges the auth code for access + refresh tokens
func (r *OauthClient) Authorize(ctx context.Context, code string) (string, error) {
	token, err := r.config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %w", err)
	}

	return token.AccessToken, nil
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
