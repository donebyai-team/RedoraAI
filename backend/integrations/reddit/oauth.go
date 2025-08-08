package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/errorx"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"strings"
	"sync"
)

type OauthClient struct {
	logger       *zap.Logger
	clientID     string
	clientSecret string
	config       *oauth2.Config
	db           datastore.Repository
	httpClient   *http.Client

	mu          sync.Mutex
	clientCache map[string]*Client // orgID -> RedditClient
}

type userAgentTransport struct {
	userName string
	base     http.RoundTripper
}

func (u *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if u.userName != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("com.redoraai:v0.1 by (/u/%s)", u.userName))
	} else {
		req.Header.Set("User-Agent", "com.redoraai:v0.1.0 by (redora)")
	}
	return u.base.RoundTrip(req)
}

func NewRedditOauthClient(logger *zap.Logger, db datastore.Repository, clientID, clientSecret, redirectURL string) *OauthClient {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"identity", "read", "mysubreddits", "submit", "subscribe", "flair"},
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

	oauthClient := &OauthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config,
		logger:       logger,
		db:           db,
		httpClient:   client,
		clientCache:  make(map[string]*Client),
	}
	return oauthClient
}

func (c *OauthClient) WithRotatingAccounts(
	ctx context.Context,
	orgID string,
	integrationType models.IntegrationType,
	fn func(integration *models.Integration) error,
	strategy IntegrationSelectionStrategy,
) error {
	return c.withRotatingIntegrations(ctx, orgID, integrationType,
		nil,
		nil,
		fn,
		strategy,
	)
}

func (c *OauthClient) WithRotatingAPIClient(
	ctx context.Context,
	orgID string,
	fn func(client *Client) error,
	strategy IntegrationSelectionStrategy,
) error {
	return c.withRotatingIntegrations(ctx, orgID, models.IntegrationTypeREDDIT,
		func(integration *models.Integration) (*Client, error) {
			return c.buildRedditClient(ctx, integration)
		},
		fn,
		nil,
		strategy,
	)
}

func (c *OauthClient) withRotatingIntegrations(
	ctx context.Context,
	orgID string,
	integrationType models.IntegrationType,
	clientBuilder func(integration *models.Integration) (*Client, error),
	clientHandler func(*Client) error,
	integrationHandler func(*models.Integration) error,
	strategy IntegrationSelectionStrategy,
) error {
	integrations, err := c.db.GetIntegrationByOrgAndType(ctx, orgID, integrationType)
	if err != nil {
		return fmt.Errorf("failed to get integrations: %w", err)
	}

	var activeIntegrations []*models.Integration
	for _, integration := range integrations {
		if integration.State == models.IntegrationStateACTIVE {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	if len(activeIntegrations) == 0 {
		return datastore.IntegrationNotFoundOrActive
	}

	// Final ordered list
	finalIntegrations := strategy(activeIntegrations)

	var lastErr error
	banned, notEstablished := 0, 0

	for _, integration := range finalIntegrations {
		if integration.ReferenceID == nil {
			c.logger.Error("reference id is nil", zap.String("integration_id", integration.ID))
			continue
		}

		var client *Client

		if clientBuilder != nil {
			client, err = clientBuilder(integration)
			if err != nil {
				lastErr = err
				continue
			}
		} else {
			client = NewClientWithOutConfig(c.logger)
		}

		// Attempt GetUser call
		// If it gives any other error other than an account banned, skip it
		_, getUserErr := client.GetUser(ctx, *integration.ReferenceID)
		if getUserErr != nil {
			if errors.Is(getUserErr, AccountBanned) {
				lastErr = getUserErr
				banned++
				c.logger.Error("account is banned", zap.String("integration_id", integration.ID), zap.Error(getUserErr))
				c.revokeIntegration(ctx, integration.ID, models.IntegrationStateACCOUNTSUSPENDED)
				continue
			}

			c.logger.Warn("warn:failed to check user", zap.String("integration_id", integration.ID), zap.Error(getUserErr))
		}

		if clientBuilder != nil {
			err = clientHandler(client)
			if err == nil {
				return nil
			}
			lastErr = err
		} else if integrationHandler != nil {
			err = integrationHandler(integration)
			if err == nil {
				return nil
			}
			lastErr = err
		}

		// Revoke integration if account isn't established
		if strings.Contains(lastErr.Error(), "account isn't established") {
			notEstablished++
			c.logger.Error("account isn't established", zap.String("integration_id", integration.ID), zap.Error(err))
			c.revokeIntegration(ctx, integration.ID, models.IntegrationStateNOTESTABLISHED)
		}
	}

	switch {
	case banned == len(finalIntegrations):
		return AllAccountBanned
	case notEstablished == len(finalIntegrations):
		return AllAccountNotEstablished
	default:
		return lastErr
	}
}

func (c *OauthClient) GetActiveIntegrations(ctx context.Context, orgID string, integrationType models.IntegrationType) ([]*models.Integration, error) {
	integrations, err := c.db.GetIntegrationByOrgAndType(ctx, orgID, integrationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	var activeIntegrations []*models.Integration
	for _, integration := range integrations {
		if integration.State == models.IntegrationStateACTIVE {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}

	return activeIntegrations, nil
}

// GetRedditAPIClient gives a random connected account as per below strategy
// 1. Try to find an account for which both integrations type REDDIT and DM exist to let a single user do both comment and DM
// 2. Prioritize the one that is > 2 weeks old
// TODO: Also check is account is suspended
func (c *OauthClient) GetRedditAPIClient(ctx context.Context, orgID string, forceAuth bool) (*Client, error) {
	activeRedditIntegrations, err := c.GetActiveIntegrations(ctx, orgID, models.IntegrationTypeREDDIT)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	if len(activeRedditIntegrations) == 0 {
		if !forceAuth {
			return NewClientWithOutConfig(c.logger), nil
		}
		return nil, datastore.IntegrationNotFoundOrActive
	}

	var candidates []*models.Integration

	if !forceAuth {
		candidates = RandomStrategy(activeRedditIntegrations)
	} else {
		candidates = MostQualifiedAccountStrategy(c.logger)(activeRedditIntegrations)
	}

	// Randomly select one from candidates
	for _, integration := range candidates {
		client, err := c.buildRedditClient(ctx, integration)
		if err == nil {
			return client, nil
		}
		c.logger.Warn("failed to build reddit client from integration", zap.String("integration_id", integration.ID), zap.Error(err))
	}

	if !forceAuth {
		c.logger.Warn("all reddit integrations failed, using unauthenticated client")
		return NewClientWithOutConfig(c.logger), nil
	}

	c.logger.Error("failed to build reddit client from any integrations")
	return nil, datastore.IntegrationNotFoundOrActive
}

func (c *OauthClient) buildRedditClient(ctx context.Context, integration *models.Integration) (*Client, error) {
	redditUserConfig := integration.GetRedditConfig()

	client := &Client{
		logger:      c.logger,
		config:      redditUserConfig,
		httpClient:  newHTTPClient(redditUserConfig.Name),
		oauthConfig: c.config,
		baseURL:     redditAPIBase,
		unAuthorizedErrorCallback: func(ctx context.Context) {
			_ = c.revokeIntegration(ctx, integration.ID, models.IntegrationStateAUTHEXPIRED)
		},
	}

	if client.isTokenExpired() {
		c.logger.Info("token expired, refreshing...", zap.String("integration_id", integration.ID))
		err := client.refreshToken(ctx)
		if err != nil {
			c.logger.Error("failed to refresh token", zap.String("integration_id", integration.ID), zap.Error(err))
			client.unAuthorizedErrorCallback(ctx)
			return nil, &errorx.RefreshTokenError{Reason: err.Error()}
		}

		// Update credentials in DB
		updated := models.SetIntegrationType(integration, models.IntegrationTypeREDDIT, client.config)
		_, err = c.db.UpsertIntegration(ctx, updated)
		if err != nil {
			return nil, fmt.Errorf("failed to update integration after token refresh: %w", err)
		}
	} else {
		c.logger.Info("token exists, using existing token",
			zap.String("expiry", client.config.ExpiresAt.String()),
			zap.String("integration_id", integration.ID))
	}

	return client, nil
}

func (c *OauthClient) revokeIntegration(ctx context.Context, integrationID string, state models.IntegrationState) error {
	integration, err := c.db.GetIntegrationById(ctx, integrationID)
	if err != nil {
		c.logger.Error("failed to fetch integration to revoke", zap.String("integration_id", integrationID), zap.Error(err))
		return err
	}
	integration.State = state
	_, err = c.db.UpsertIntegration(ctx, integration)
	if err != nil {
		c.logger.Error("failed to mark integration as AUTHREVOKED", zap.Error(err))
	}
	c.logger.Info("integration marked as revoked", zap.String("state", state.String()), zap.String("integration_id", integrationID))
	return err
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

	//Verify If Account is active
	_, err = NewClientWithOutConfig(r.logger).GetUser(ctx, userInfo.Name)
	if err != nil {
		return nil, err
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
