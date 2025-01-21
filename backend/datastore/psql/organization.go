package psql

import (
	"context"
	"fmt"

	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"organization/create_organization.sql",
		"organization/update_organization.sql",
		"organization/update_organization.sql",
		"organization/query_all_organizations.sql",
		"organization/query_organization_by_id.sql",
		"organization/query_organization_by_name.sql",
	})
}

func (r *Database) CreateOrganization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	stmt := r.mustGetStmt("organization/create_organization.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"name":          organization.Name,
		"feature_flags": organization.FeatureFlags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	organization.ID = id
	return organization, nil
}

func (r *Database) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	stmt := r.mustGetStmt("organization/update_organization.sql")

	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"id":            org.ID,
		"feature_flags": org.FeatureFlags,
	})

	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil

}

func (r *Database) GetOrganizations(ctx context.Context) ([]*models.Organization, error) {
	return getMany[models.Organization](ctx, r, "organization/query_all_organizations.sql", nil)
}

func (r *Database) GetOrganizationById(ctx context.Context, organizationID string) (*models.Organization, error) {
	return getOne[models.Organization](ctx, r, "organization/query_organization_by_id.sql", map[string]any{
		"id": organizationID,
	})
}

func (r *Database) GetOrganizationByName(ctx context.Context, organizationName string) (*models.Organization, error) {
	return getOne[models.Organization](ctx, r, "organization/query_organization_by_name.sql", map[string]any{
		"name": organizationName,
	})
}
