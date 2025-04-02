package services

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/auth/crypto"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

type AuthUsecase struct {
	auth0  *auth0
	db     datastore.Repository
	signer crypto.SigningKeyGetter
	logger *zap.Logger
}

func NewAuthUsecase(ctx context.Context, auth0Config *Auth0Config, db datastore.Repository, signingAPIKeyGetter crypto.SigningKeyGetter, logger *zap.Logger) (*AuthUsecase, error) {
	auth0, err := newAuth0(ctx, auth0Config, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth0: %w", err)
	}

	return &AuthUsecase{
		auth0:  auth0,
		db:     db,
		signer: signingAPIKeyGetter,
		logger: logger,
	}, nil
}
func (a *AuthUsecase) StartPasswordless(ctx context.Context, email string, ip string) error {
	if err := a.auth0.initiatePasswordlessFlow(email, ip); err != nil {
		return fmt.Errorf("failed to initiate passwordless flow: %w", err)
	}
	return nil
}

func (a *AuthUsecase) VerifyPasswordless(ctx context.Context, email string, code string, ip string) (*pbportal.JWT, error) {
	logger := logging.Logger(ctx, a.logger)
	auth0Token, err := a.auth0.verifyPasswordlessFlow(code, email, ip)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Note: the auth0User only has an external id, no email attached, unlike standard code flow
	auth0User, err := a.verifyAndDecodeRawIDToken(ctx, auth0Token.IdToken, nil)
	if err != nil {
		return nil, fmt.Errorf("decode id getToken: %w", err)
	}

	jwt, err := a.getUser(ctx, email, auth0User.ExternalAuthProviderID, false, logger)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return jwt, nil
}

func (a *AuthUsecase) verifyAndDecodeRawIDToken(ctx context.Context, idToken string, nonce *string) (*Auth0User, error) {
	token, err := a.auth0.verifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("unable to validate auth0 request: %w", err)
	}

	return a.decodeIDToken(token, nonce)
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (p *AuthUsecase) decodeIDToken(idToken *oidc.IDToken, nonce *string) (*Auth0User, error) {
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if nonce != nil {
		if claims["nonce"] != *nonce {
			return nil, fmt.Errorf("invalid nonce")
		}
	}

	externalAuthProviderID, ok := claims["sub"].(string)
	if !ok || externalAuthProviderID == "" {
		return nil, fmt.Errorf("invalid claim")
	}

	internalID, found := claims["https://id.rarecircles.io/user_id"].(string) // same in staging
	if !found {
		internalID = ""
	}

	var email string
	email, ok = claims["email"].(string)
	if !ok {
		p.logger.Warn("email is missing from Auth0 JWT claims")
	}

	emailVerified, _ := claims["email_verified"].(bool)

	return &Auth0User{
		InternalID:             internalID,
		ExternalAuthProviderID: externalAuthProviderID,
		Email:                  email,
		EmailVerified:          emailVerified,
	}, nil
}

func (a *AuthUsecase) getUser(ctx context.Context, email, externalAuthProviderID string, emailVerified bool, logger *zap.Logger) (*pbportal.JWT, error) {
	user, err := a.db.GetUserByEmail(ctx, email)
	if err != nil && err != datastore.NotFound {
		logger.Warn("failed to get user", zap.Error(err),
			zap.String("external_id", externalAuthProviderID),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("unable to get user: %w", err)
	}

	if err != nil {
		if errors.Is(err, datastore.NotFound) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("User account not found, please contact your system administrator"))
		}

		return nil, fmt.Errorf("unable to retrieved user: %w", err)
	}

	resp, err := a.getJWTToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("unable to create token for user credentials: %w", err)
	}

	logger.Info("jwt issued", zap.String("user_id", user.ID))
	return resp, nil
}

func (a *AuthUsecase) getJWTToken(ctx context.Context, user *models.User) (*pbportal.JWT, error) {
	key := a.signer.GetKey(ctx)

	unsignedToken, signedToken, err := auth.NewTokenFromUser(user, key)
	if err != nil {
		return nil, err
	}

	claims, _ := unsignedToken.Claims.(*auth.Credentials)
	return &pbportal.JWT{
		Token:     signedToken,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

type Auth0User struct {
	InternalID             string
	ExternalAuthProviderID string
	Email                  string
	EmailVerified          bool
}
