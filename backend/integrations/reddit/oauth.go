package reddit

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

type OauthClient struct {
	logger       *zap.Logger
	clientID     string
	clientSecret string
	config       *oauth2.Config
	db           datastore.Repository
	httpClient   *http.Client
}

type userAgentTransport struct {
	base http.RoundTripper
}

func (u *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "linux:com.reddit.scraper:v0.1")
	return u.base.RoundTrip(req)
}

func NewRedditOauthClient(logger *zap.Logger, db datastore.Repository, clientID, clientSecret, redirectURL string) *OauthClient {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"identity", "read", "modconfig", "mysubreddits"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  redditAuthURL,
			TokenURL: redditTokenURL,
		},
	}

	// simple http client to overrride roundtripper
	client := &http.Client{
		Transport: &userAgentTransport{
			base: http.DefaultTransport,
		},
	}
	return &OauthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config,
		logger:       logger,
		db:           db,
		httpClient:   client,
	}
}

func (r *OauthClient) NewRedditClient(ctx context.Context, orgID string) (*Client, error) {
	integration, err := r.db.GetIntegrationByOrgAndType(ctx, orgID, models.IntegrationTypeREDDIT)
	if err != nil && errors.Is(err, datastore.NotFound) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("reddit integration not configured"))
	}

	if err != nil {
		return nil, err
	}

	if integration.State != models.IntegrationStateACTIVE {
		return nil, fmt.Errorf("reddit integration not active")
	}

	client := &Client{
		logger:      r.logger,
		config:      integration.GetRedditConfig(),
		httpClient:  newHTTPClient(),
		oauthConfig: r.config,
		baseURL:     redditAPIBase,
		unAuthorizedErrorCallback: func(ctx context.Context) {
			integration.State = models.IntegrationStateAUTHREVOKED
			integration, err = r.db.UpsertIntegration(ctx, integration)
			if err != nil {
				r.logger.Error("failed to update reddit integration state to auth_revoked", zap.Error(err))
			}
		},
	}
	if client.isTokenExpired() {
		err := client.refreshToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Update the credentials
		integrationType := models.SetIntegrationType(integration, models.IntegrationTypeREDDIT, client.config)
		integration, err = r.db.UpsertIntegration(ctx, integrationType)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert reddit integration: %w", err)
		}
	}
	return client, nil
}

// GetAuthURL returns the authorization URL
func (r *OauthClient) GetAuthURL(state string) string {
	base := r.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return base + "&duration=permanent&approval_prompt=force"
}

// Authorize exchanges the auth code for access + refresh tokens
func (r *OauthClient) Authorize(ctx context.Context, code string) (*models.RedditConfig, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, r.httpClient)
	token, err := r.config.Exchange(ctx, code, oauth2.AccessTypeOffline)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// TODO: Remove it later
	r.logger.Info("reddit token received", zap.String("token", token.AccessToken))

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
		Verified         bool    `json:"verified"`
		Coins            float64 `json:"coins"`
		Id               string  `json:"id"`
		OauthClientId    string  `json:"oauth_client_id"`
		IsMod            bool    `json:"is_mod"`
		AwarderKarma     float64 `json:"awarder_karma"`
		HasVerifiedEmail bool    `json:"has_verified_email"`
		IsSuspended      bool    `json:"is_suspended"`
		AwardeeKarma     float64 `json:"awardee_karma"`
		LinkKarma        float64 `json:"link_karma"`
		TotalKarma       float64 `json:"total_karma"`
		InboxCount       int     `json:"inbox_count"`
		Name             string  `json:"name"`
		Created          float64 `json:"created"`
		CreatedUtc       float64 `json:"created_utc"`
		CommentKarma     float64 `json:"comment_karma"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &models.RedditConfig{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		Verified:         userInfo.Verified,
		Coins:            userInfo.Coins,
		Id:               userInfo.Id,
		OauthClientId:    userInfo.OauthClientId,
		IsMod:            userInfo.IsMod,
		AwarderKarma:     userInfo.AwarderKarma,
		HasVerifiedEmail: userInfo.HasVerifiedEmail,
		IsSuspended:      userInfo.IsSuspended,
		AwardeeKarma:     userInfo.AwardeeKarma,
		LinkKarma:        userInfo.LinkKarma,
		TotalKarma:       userInfo.TotalKarma,
		InboxCount:       userInfo.InboxCount,
		Name:             userInfo.Name,
		Created:          userInfo.Created,
		CreatedUtc:       userInfo.CreatedUtc,
		CommentKarma:     userInfo.CommentKarma,
		ExpiresAt:        token.Expiry,
	}, nil
}
