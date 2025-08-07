package events

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type OrgUserEvent struct {
	UserID              string                      `json:"user_id"`
	Email               string                      `json:"email"`
	ProductName         string                      `json:"product_name"`
	SubscriptionPlan    models.SubscriptionPlanType `json:"subscription_plan"`
	SubscriptionExpired bool                        `json:"subscription_expired"`
}

type EventPublisher struct {
	db               datastore.Repository
	logger           *zap.Logger
	brevoIntegration *Brevo
}

func NewEventPublisher(db datastore.Repository, logger *zap.Logger, brevoIntegration *Brevo) *EventPublisher {
	return &EventPublisher{db: db, logger: logger, brevoIntegration: brevoIntegration}
}

func (b *EventPublisher) BulkInsert(ctx context.Context) error {
	organizations, err := b.db.GetOrganizations(ctx)
	if err != nil {
		return nil
	}

	for _, org := range organizations {
		users, err := b.db.GetUsersByOrgID(context.Background(), org.ID)
		if err != nil {
			return nil
		}
		for _, user := range users {
			err := b.CreateUser(context.Background(), user.Email)
			if err != nil {
				b.logger.Error("failed to create user event", zap.Error(err))
			}
		}
	}
	
	return nil
}

func (b *EventPublisher) UpdateUsers(ctx context.Context, orgID string) error {
	org, err := b.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return err
	}

	users, err := b.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	projects, err := b.db.GetProjects(ctx, orgID)
	if err != nil {
		return err
	}

	projectName := "aaa"
	if len(projects) > 0 {
		projectName = projects[0].Name
	}

	for _, user := range users {
		err = b.brevoIntegration.UpdateContact(ctx, &OrgUserEvent{
			UserID:              user.ID,
			Email:               user.Email,
			ProductName:         projectName,
			SubscriptionPlan:    org.FeatureFlags.Subscription.PlanID,
			SubscriptionExpired: org.FeatureFlags.Subscription.IsExpired(),
		})
		if err != nil {
			b.logger.Error("failed to update contact", zap.Error(err), zap.String("email", user.Email))
		}
	}

	return nil
}

func (b *EventPublisher) CreateUser(ctx context.Context, emailID string) error {
	user, err := b.db.GetUserByEmail(ctx, emailID)
	if err != nil {
		return err
	}

	org, err := b.db.GetOrganizationById(ctx, user.OrganizationID)
	if err != nil {
		return err
	}

	projects, err := b.db.GetProjects(ctx, org.ID)
	if err != nil {
		return err
	}

	projectName := ""
	if len(projects) > 0 {
		projectName = projects[0].Name
	}

	err = b.brevoIntegration.CreateContact(ctx, &OrgUserEvent{
		UserID:              user.ID,
		Email:               user.Email,
		ProductName:         projectName,
		SubscriptionPlan:    org.FeatureFlags.Subscription.PlanID,
		SubscriptionExpired: org.FeatureFlags.Subscription.IsExpired(),
	})
	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	return nil
}
