package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	httpClient   *retryablehttp.Client
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
		httpClient:   newHTTPClient(),
	}
}

// GetAuthURL returns the authorization URL
func (r *OauthClient) GetAuthURL(state string) string {
	base := r.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return base + "&duration=permanent&approval_prompt=force"
}

type userAgentTransport struct {
	base http.RoundTripper
}

func (u *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "linux:com.reddit.scraper:v0.1 (by /u/Ashamed-Lime-6816)")
	return u.base.RoundTrip(req)
}

// Authorize exchanges the auth code for access + refresh tokens
func (r *OauthClient) Authorize(ctx context.Context, code string) (*models.RedditConfig, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, r.httpClient.HTTPClient)
	token, err := r.config.Exchange(ctx, code, oauth2.AccessTypeOffline)
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

	client := &Client{
		logger:      r.logger,
		config:      integration.GetRedditConfig(),
		httpClient:  r.httpClient,
		oauthConfig: r.config,
		baseURL:     redditAPIBase,
	}
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
	oauthConfig *oauth2.Config
	httpClient  *retryablehttp.Client
	userAgent   string
	baseURL     string
}

func NewClientWithConfig(config *models.RedditConfig, logger *zap.Logger) *Client {
	return &Client{
		baseURL:    redditAPIBase,
		config:     config,
		logger:     logger,
		httpClient: newHTTPClient(),
	}
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

func newHTTPClient() *retryablehttp.Client {
	cli := retryablehttp.NewClient()
	cli.Logger = nil
	cli.RetryMax = 1

	// Your existing transport
	baseTransport := &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DisableKeepAlives:  false,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 300 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Wrap with user-agent injection
	cli.HTTPClient.Transport = &userAgentTransport{base: baseTransport}

	cli.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		return resp, err
	}
	return cli
}

var ErrNotFound = errors.New("not found")

func (r *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	reqURL := fmt.Sprintf("%s/user/%s/about.json", r.baseURL, userID)
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Data, nil
}

func (r *Client) GetSubRedditByURL(ctx context.Context, urlPath string) (*SubReddit, error) {
	if !strings.HasPrefix(urlPath, "/r/") {
		return nil, fmt.Errorf("invalid subreddit URL path: %s", urlPath)
	}
	reqURL := fmt.Sprintf("%s%sabout.json", r.baseURL, urlPath)
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data SubReddit `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
}

func (r *Client) GetPosts(ctx context.Context, subRedditID string, filters PostFilters) ([]*Post, error) {
	v := url.Values{}
	if len(filters.Keywords) > 0 {
		v.Set("q", strings.Join(filters.Keywords, " "))
	}

	if filters.SortBy != nil {
		v.Set("sort", string(*filters.SortBy))
	}

	if filters.TimeRage != nil {
		v.Set("t", string(*filters.TimeRage))
	}

	v.Set("restrict_sr", "1")
	reqURL := fmt.Sprintf("%s/r/%s/search.json?%s", r.baseURL, subRedditID, v.Encode())
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data struct {
			Children []struct {
				Data *Post `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var posts []*Post
	for _, child := range response.Data.Children {
		posts = append(posts, child.Data)
	}

	return posts, nil
}

func (r *Client) GetPostByID(ctx context.Context, postID string) (*Post, error) {
	reqURL := fmt.Sprintf("%s/comments/%s.json", r.baseURL, postID)
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response []struct {
		Data struct {
			Children []struct {
				Data Post `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response) > 0 && len(response[0].Data.Children) > 0 {
		return &response[0].Data.Children[0].Data, nil
	}

	return nil, ErrNotFound // Post not found in the response
}

func (r *Client) GetPostWithAllComments(ctx context.Context, postID string) (*Post, error) {
	reqURL := fmt.Sprintf("%s/comments/%s.json?raw_json=1&sort=new", r.baseURL, postID)
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rawResp []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return r.parsePostWithComments(ctx, rawResp)
}

func (r *Client) isTokenExpired() bool {
	// Refresh if within 60 seconds of expiry
	return time.Now().After(r.config.ExpiresAt.Add(-1 * time.Minute))
}
