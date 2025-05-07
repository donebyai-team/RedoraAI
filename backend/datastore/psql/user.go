package psql

import (
	"context"
	"fmt"

	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"user/create_user.sql",
		"user/update_user.sql",
		"user/query_user_by_id.sql",
		"user/query_user_by_org.sql",
		"user/query_user_by_email.sql",
		"user/query_user_by_auth0_id.sql",
	})

}

func (r *Database) GetUsersByOrgID(ctx context.Context, orgID string) ([]*models.User, error) {
	return getMany[models.User](ctx, r, "user/query_user_by_org.sql", map[string]any{
		"organization_id": orgID,
	})
}

func (r *Database) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	stmt := r.mustGetStmt("user/create_user.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"auth0_id":        user.Auth0ID,
		"email":           user.Email,
		"email_verified":  user.EmailVerified,
		"organization_id": user.OrganizationID,
		"role":            user.Role,
		"state":           user.State,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	user.ID = id
	return user, nil
}

func (r *Database) UpdateUser(ctx context.Context, user *models.User) error {
	stmt := r.mustGetStmt("user/update_user.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"id":              user.ID,
		"email":           user.Email,
		"email_verified":  user.EmailVerified,
		"organization_id": user.OrganizationID,
		"role":            user.Role,
		"state":           user.State,
	})
	if err != nil {
		return fmt.Errorf("failed to update user %q: %w", user.ID, err)
	}
	return nil
}

func (r *Database) GetUserById(ctx context.Context, userID string) (*models.User, error) {
	return getOne[models.User](ctx, r, "user/query_user_by_id.sql", map[string]any{
		"id": userID,
	})
}

func (r *Database) GetUserByAuth0Id(ctx context.Context, auth0ID string) (*models.User, error) {
	return getOne[models.User](ctx, r, "user/query_user_by_auth0_id.sql", map[string]any{
		"auth0_id": auth0ID,
	})
}

func (r *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return getOne[models.User](ctx, r, "user/query_user_by_email.sql", map[string]any{
		"email": email,
	})
}
