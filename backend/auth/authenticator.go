package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shank318/doota/datastore"
)

type Authenticator struct {
	tokenValidator TokenValidationFunc
	exemptPaths    map[string]bool
	db             datastore.Repository
	logger         *zap.Logger
}

func NewAuthenticator(
	tokenValidator TokenValidationFunc,
	db datastore.Repository,
	logger *zap.Logger,
) *Authenticator {
	return &Authenticator{
		tokenValidator: tokenValidator,
		db:             db,
		exemptPaths: map[string]bool{
			"/doota.portal.v1.PortalService/GetConfig":           true,
			"/doota.portal.v1.PortalService/AuthState":           true,
			"/doota.portal.v1.PortalService/Issue":               true,
			"/doota.portal.v1.PortalService/PasswordlessStart":   true,
			"/doota.portal.v1.PortalService/PasswordlessVerify":  true,
			"/doota.portal.v1.PortalService/SocialLoginCallback": true,
			"/doota.portal.v1.PortalService/OauthAuthorize":      true,
		},
		logger: logger,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, path string, headers map[string][]string, ipAddress string) (context.Context, error) {
	if _, found := a.exemptPaths[path]; found {
		return ctx, nil
	}

	jwtToken, err := extractToken(headers)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, fmt.Sprintf("unable to extract token: %s", err)).Err()
	}

	if jwtToken == "" {
		return nil, status.New(codes.Unauthenticated, "empty token").Err()
	}

	credentials, err := a.tokenValidator(jwtToken)
	if err != nil {
		return nil, status.New(codes.Unauthenticated, fmt.Sprintf("failed to validate JWT token: %s", err)).Err()
	}

	user, err := a.db.GetUserById(ctx, credentials.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	if isAdminPath(path) {
		if !user.IsPlatformAdmin() {
			return nil, status.New(codes.PermissionDenied, "unauthorized access: not an admin").Err()
		}
	}

	orgID := user.OrganizationID
	if in := getHeader(headers, "X-Organization-Id"); in != "" {
		if !user.IsPlatformAdmin() {
			return nil, status.New(codes.PermissionDenied, "unauthorized org id: not an admin").Err()
		}
		orgID = in
	}

	return WithAuthContext(ctx, &AuthContext{
		User:           user,
		OrganizationID: orgID,
	}), nil
}

func (a *Authenticator) Ready(ctx context.Context) bool {
	return true
}

func extractToken(header map[string][]string) (string, error) {
	authHeaders, ok := header["Authorization"]
	if !ok {
		return "", fmt.Errorf("no authorization header found")
	}

	bearerToken := authHeaders[0]
	splitToken := strings.Split(bearerToken, " ")
	if len(splitToken) != 2 {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	jwtToken := splitToken[1]

	return jwtToken, nil
}
