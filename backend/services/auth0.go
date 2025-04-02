package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Auth0Config struct {
	Auth0PortalClientID     string
	Auth0PortalClientSecret string
	Auth0ApiRedirectURL     string
	Auth0Domain             string
}

type auth0 struct {
	oauthProvider *oidc.Provider
	oauthConf     oauth2.Config
	oidcConfig    *oidc.Config
	domain        string
	logger        *zap.Logger
}

func newAuth0(ctx context.Context, auth0Config *Auth0Config, logger *zap.Logger) (*auth0, error) {
	logger.Info("setting up auth0",
		zap.String("auth0_domain", auth0Config.Auth0Domain),
		zap.String("auth0_client_id", auth0Config.Auth0PortalClientID),
		zap.String("auth0_client_secret", strings.Repeat("*", len(auth0Config.Auth0PortalClientSecret))),
	)
	auth0URI := "https://" + auth0Config.Auth0Domain + "/"
	provider, err := oidc.NewProvider(ctx, auth0URI)
	if err != nil {
		return nil, fmt.Errorf("oauth provider %q: %w", auth0URI, err)
	}

	oidcConfig := &oidc.Config{
		ClientID: auth0Config.Auth0PortalClientID,
	}

	conf := oauth2.Config{
		ClientID:     auth0Config.Auth0PortalClientID,
		ClientSecret: auth0Config.Auth0PortalClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  auth0Config.Auth0ApiRedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &auth0{
		oauthProvider: provider,
		oauthConf:     conf,
		oidcConfig:    oidcConfig,
		domain:        auth0Config.Auth0Domain,
		logger:        logger,
	}, nil
}

func (a *auth0) verifyIDToken(ctx context.Context, tokenId string) (*oidc.IDToken, error) {
	return a.oauthProvider.Verifier(a.oidcConfig).Verify(ctx, tokenId)
}

type PasswordlessRequest struct {
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	Connection   string            `json:"connection"`
	Email        string            `json:"email"`
	Send         string            `json:"send"`
	AuthParams   map[string]string `json:"authParams"`
}

func (a *auth0) initiatePasswordlessFlow(email, ipaddress string) error {
	req := &PasswordlessRequest{
		ClientID:     a.oauthConf.ClientID,
		ClientSecret: a.oauthConf.ClientSecret,
		Connection:   "email",
		Email:        email,
		Send:         "code",
		AuthParams: map[string]string{
			"scope": "openid profile",
		},
	}

	url := fmt.Sprintf("https://%s/passwordless/start", a.domain)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("masrshal request: %w", err)
	}

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	r.Header.Add("content-type", "application/json")

	// Ensure we're not rate-limited
	// [Reference] https://auth0.com/docs/get-started/authentication-and-authorization-flow/avoid-common-issues-with-resource-owner-password-flow-and-attack-protection#configure-your-application-to-trust-the-ip-address
	r.Header.Add("auth0-forwarded-for", ipaddress)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	cnt, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if err := gjson.GetBytes(cnt, "error").String(); err != "" {
		return fmt.Errorf("failed to start passwordless: %s", gjson.GetBytes(cnt, "error_description").String())
	}

	type Response struct {
		Id            string `json:"_id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	out := &Response{}
	if err := json.Unmarshal(cnt, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	a.logger.Info("passwordless started",
		zap.String("email", email),
		zap.String("ip_address", ipaddress),
		zap.Reflect("out", out),
	)

	return nil
}

type PasswordlessVerificationRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	OTP          string `json:"otp"`
	Realm        string `json:"realm"`
	Scope        string `json:"scope"`
}

type Auth0Token struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (a *auth0) verifyPasswordlessFlow(code, email, ipaddress string) (*Auth0Token, error) {

	req := &PasswordlessVerificationRequest{
		GrantType:    "http://auth0.com/oauth/grant-type/passwordless/otp",
		ClientID:     a.oauthConf.ClientID,
		ClientSecret: a.oauthConf.ClientSecret,
		Username:     email,
		OTP:          code,
		Realm:        "email",
		Scope:        "openid profile",
	}

	url := fmt.Sprintf("https://%s/oauth/token", a.domain)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("masrshal request: %w", err)
	}

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	r.Header.Add("content-type", "application/json")

	// Ensure we're not rate-limited
	// [Reference] https://auth0.com/docs/get-started/authentication-and-authorization-flow/avoid-common-issues-with-resource-owner-password-flow-and-attack-protection#configure-your-application-to-trust-the-ip-address
	r.Header.Add("auth0-forwarded-for", ipaddress)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	cnt, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if err := gjson.GetBytes(cnt, "error").String(); err != "" {
		return nil, fmt.Errorf(gjson.GetBytes(cnt, "error_description").String())
	}

	out := &Auth0Token{}
	if err := json.Unmarshal(cnt, out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	a.logger.Debug("passwordless verified")

	return out, nil
}
