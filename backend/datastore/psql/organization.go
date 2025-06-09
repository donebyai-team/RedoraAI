package psql

import (
	"context"
	"fmt"
	"time"

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
		"subscription/create_subscription.sql",
	})
}

func (r *Database) CreateOrganization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	stmt := r.mustGetTxStmt(ctx, "organization/create_organization.sql", tx)
	var id string

	// Create a free sub and store it in a feature flag for faster access
	subscription := CreateFreeSubscription()
	organization.FeatureFlags.Subscription = subscription

	err = stmt.GetContext(ctx, &id, map[string]interface{}{
		"name":          organization.Name,
		"feature_flags": organization.FeatureFlags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	organization.ID = id

	subscription.OrganizationID = id
	// Create subscription
	_, err = r.CreateSubscription(ctx, subscription, tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return organization, nil
}

func CreateFreeSubscription() *models.Subscription {
	plan := models.RedoraPlans[models.SubscriptionPlanTypeFREE]
	now := time.Now()
	expiresAt := now.Add(time.Duration(plan.Interval) * 24 * time.Hour)
	return &models.Subscription{
		Amount:    plan.Price,
		PlanID:    models.SubscriptionPlanTypeFREE,
		Status:    models.SubscriptionStatusACTIVE,
		Metadata:  plan.Metadata,
		ExpiresAt: expiresAt,
	}
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
