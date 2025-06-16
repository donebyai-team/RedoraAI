package psql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	subscription := CreateSubscriptionObject(models.SubscriptionPlanTypeFREE)
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

func CreateSubscriptionObject(planType models.SubscriptionPlanType) *models.Subscription {
	plan := models.RedoraPlans[planType]
	now := time.Now()
	expiresAt := now.Add(time.Duration(plan.Interval) * 24 * time.Hour)
	return &models.Subscription{
		Amount:    plan.Price,
		PlanID:    planType,
		Status:    models.SubscriptionStatusACTIVE,
		Metadata:  plan.Metadata,
		ExpiresAt: expiresAt,
	}
}

const (
	FEATURE_FLAG_DISABLE_AUTOMATED_COMMENT_PATH = "enable_auto_comment"
	FEATURE_FLAG_ACTIVITIES_PATH                = "activities"
	FEATURE_FLAG_SUBSCRIPTION_PATH              = "subscription"
	FEATURE_FLAG_SUBSCRIPTION_EXTERNAL_ID_PATH  = "subscription.external_id"
	FEATURE_FLAG_DISABLE_AUTOMATED_DM_PATH      = "enable_auto_dm"
	FEATURE_FLAG_NOTIFICATION_FREQUENCY_PATH    = "notification_settings.notification_frequency_posts"
	FEATURE_FLAG_NITIFICATION_LAST_SENT_AT_PATH = "notification_settings.last_relevant_post_alert_sent_at"
)

func (r *Database) UpdateOrganizationFeatureFlags(ctx context.Context, orgID string, updates map[string]any) error {
	for flatPath, value := range updates {
		pathParts := strings.Split(flatPath, ".")
		if len(pathParts) == 0 {
			continue
		}

		// Build Postgres-style path strings
		fullPath := "{" + strings.Join(pathParts, ",") + "}"

		valJSON, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("marshal value for path %s: %w", flatPath, err)
		}

		var query string
		var args []any

		// If it's nested, ensure parent object exists
		if len(pathParts) > 1 {
			parentPath := "{" + strings.Join(pathParts[:len(pathParts)-1], ",") + "}"

			query = `
				UPDATE organizations
				SET feature_flags = jsonb_set(
					jsonb_set(
						feature_flags,
						$1,
						coalesce(feature_flags #> $1, '{}'::jsonb),
						true
					),
					$2,
					$3::jsonb,
					true
				),
				updated_at = now()
				WHERE id = $4
			`
			args = []any{parentPath, fullPath, string(valJSON), orgID}
		} else {
			// Flat field update
			query = `
				UPDATE organizations
				SET feature_flags = jsonb_set(
					feature_flags,
					$1,
					$2::jsonb,
					true
				),
				updated_at = now()
				WHERE id = $3
			`
			args = []any{fullPath, string(valJSON), orgID}
		}

		if _, err := r.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("failed to update path %s: %w", flatPath, err)
		}
	}

	return nil
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
