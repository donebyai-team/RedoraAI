package google

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
)

type OauthClient struct {
	config *oauth2.Config
	logger *zap.Logger
}

func NewOauthClient(
	clientID, clientSecret, redirectURL string,
	logger *zap.Logger,
) *OauthClient {

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		Endpoint: google.Endpoint,
	}

	return &OauthClient{
		config: config,
		logger: logger,
	}
}

type GoogleUser struct {
	Email string `json:"email"`
}

// ExchangeCodeForEmail exchanges auth code for a token and fetches user's email.
func (c *OauthClient) Authorize(ctx context.Context, code string) (string, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		c.logger.Error("failed to exchange code for token", zap.Error(err))
		return "", err
	}

	client := c.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.logger.Error("failed to get user info", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("non-200 from userinfo", zap.Int("status", resp.StatusCode))
		return "", fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var userInfo GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.logger.Error("failed to decode user info", zap.Error(err))
		return "", err
	}

	return userInfo.Email, nil
}

func (c *OauthClient) AuthorizeURL(hash string) string {
	return c.config.AuthCodeURL(hash, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}
