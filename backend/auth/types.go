package auth

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
)

// Credentials jwt credentials object
type Credentials struct {
	jwt.StandardClaims

	Version int    `json:"v"`
	UserId  string `json:"ui"`

	IP string `json:"-"`
}

type AuthContext struct {
	*models.User
	OrganizationID string
}

func (a *AuthContext) Identity() *pbcore.Identity {
	return &pbcore.Identity{
		UserId:         a.User.ID,
		OrganizationId: a.OrganizationID,
	}
}

func (a *AuthContext) UserID() string {
	return a.User.ID
}

// Valid validate standard claims on credentials
func (c *Credentials) Valid() error {
	return c.StandardClaims.Valid()
}

const authenticatedCredentialKey int = 0

func WithAuthContext(ctx context.Context, user *AuthContext) context.Context {
	return context.WithValue(ctx, authenticatedCredentialKey, user)
}

func FromContext(ctx context.Context) (*AuthContext, bool) {
	authUser, ok := ctx.Value(authenticatedCredentialKey).(*AuthContext)
	return authUser, ok
}
