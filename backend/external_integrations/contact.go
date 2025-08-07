package external_integrations

import (
	"github.com/shank318/doota/models"
)

type Contact struct {
	UserID              string                      `json:"user_id"`
	Email               string                      `json:"email"`
	ProductName         string                      `json:"product_name"`
	SubscriptionPlan    models.SubscriptionPlanType `json:"subscription_plan"`
	SubscriptionExpired bool                        `json:"subscription_expired"`
}
