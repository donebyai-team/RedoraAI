package psql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"subscription/create_subscription.sql",
		"subscription/query_subscription_by_orgid.sql",
	})
}

func (r *Database) CreateSubscription(ctx context.Context, sub *models.Subscription, tx *sqlx.Tx) (*models.Subscription, error) {
	stmtSub := r.mustGetTxStmt(ctx, "subscription/create_subscription.sql", tx)
	var id string
	err := stmtSub.GetContext(ctx, &id, map[string]interface{}{
		"organization_id": sub.OrganizationID,
		"plan_id":         sub.PlanID,
		"status":          sub.Status,
		"metadata":        sub.Metadata,
		"external_id":     sub.ExternalID,
		"expires_at":      sub.ExpiresAt,
		"amount":          sub.Amount,
	})
	if err != nil {
		return nil, err
	}

	sub.ID = id
	return sub, nil
}

func (r *Database) GetSubscriptionByOrgID(ctx context.Context, orgID string) (*models.Subscription, error) {
	return getOne[models.Subscription](ctx, r, "subscription/query_subscription_by_orgid.sql", map[string]any{
		"organization_id": orgID,
	})
}
