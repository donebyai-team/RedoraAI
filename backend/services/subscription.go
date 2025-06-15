package services

import (
	"context"
	"fmt"
	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/datastore/psql"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type SubscriptionService interface {
	Create(ctx context.Context, orgID string) (*models.Subscription, error)
}

type dodoSubscriptionService struct {
	db     datastore.Repository
	client *dodopayments.Client
	logger *zap.Logger
}

const (
	proPlanID     = "pdt_h2V8lsZsVRO88Q19pCjU8"
	founderPlanID = "pdt_1Arsz78Oy4pyi4MJbYeSb"
)

var planToProductID = map[models.SubscriptionPlanType]string{
	models.SubscriptionPlanTypeFOUNDER: founderPlanID,
	models.SubscriptionPlanTypePRO:     proPlanID,
}

var addOnIDMap = map[string]string{
	"adn_yIJQyUyFuX5tn2GYqqns5": "source",
	"adn_GQZ66G74wNJUH9yEuNxMG": "keyword",
}

func NewDodoSubscriptionService(db datastore.Repository, token string, logger *zap.Logger, isTest bool) *dodoSubscriptionService {
	client := dodopayments.NewClient(
		option.WithBearerToken(token),
	)
	if isTest {
		client.Options = append(client.Options, option.WithEnvironmentTestMode())
	}

	return &dodoSubscriptionService{db: db, client: client}
}

func (d dodoSubscriptionService) Create(ctx context.Context, plan models.SubscriptionPlanType, orgID, returnURL string) (*models.Subscription, error) {
	// check if externalID exists
	// if yes, call dodo and verify if the plan is active, if active return error
	// if yes, call dodo and verify if the plan is cancelled, then create new subbscription
	// if yes, call dodo and verify if same plan or not, if same plan return error
	// if not same plan, change plan call
	// if externamID not exists then update sub with external ID(temp hack)
	// On verify get the external id and check if active then update org and subscription
	// On verify get the external id and check if not active then ?
	// Listen webhook for renewal, update org and subsription
	// If cancelled via webhook or via us, update org and subsription and disable project
	// Check other webhook status ?
	//

	organization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}
	existingSub := organization.FeatureFlags.GetSubscription()

	if existingSub.PlanID == plan {
		return nil, fmt.Errorf("plan %s already exists", plan)
	}

	productId, ok := planToProductID[plan]
	if !ok {
		return nil, fmt.Errorf("invalid plan type: %s", plan)
	}

	var externalSubResponse *dodopayments.SubscriptionNewResponse

	if existingSub == nil || existingSub.ExternalID == nil {
		// create fresh
		externalSubResponse, err = d.client.Subscriptions.New(context.TODO(), dodopayments.SubscriptionNewParams{
			Billing: dodopayments.F(dodopayments.BillingAddressParam{
				City:    dodopayments.F("Bangalore"),
				Country: dodopayments.F(dodopayments.CountryCodeIn),
				State:   dodopayments.F("Karnataka"),
				Street:  dodopayments.F("Bannergetta"),
				Zipcode: dodopayments.F("560068"),
			}),
			ReturnURL: dodopayments.F(returnURL),
			Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](dodopayments.CustomerRequestParam{
				CustomerID: dodopayments.F(orgID),
			}),
			ProductID: dodopayments.F(productId),
			Quantity:  dodopayments.F(int64(1)),
		})
	} else {
		// upgrade
		err = d.client.Subscriptions.ChangePlan(ctx, *existingSub.ExternalID, dodopayments.SubscriptionChangePlanParams{
			ProductID:            dodopayments.F(productId),
			ProrationBillingMode: dodopayments.F(dodopayments.SubscriptionChangePlanParamsProrationBillingModeProratedImmediately),
			Quantity:             dodopayments.F(int64(1)),
		})
	}

	if err != nil {
		d.logger.Error("error creating subscription", zap.Error(err))
		return nil, fmt.Errorf("error creating subscription: %w", err)
	}

	if externalSubResponse == nil || externalSubResponse.SubscriptionID == "" {
		return nil, fmt.Errorf("error creating subscription: invalid subscription")
	}
	// update subscription in org
	err = d.db.UpdateOrganizationFeatureFlags(ctx, organization.ID, map[string]any{
		psql.FEATURE_FLAG_SUBSCRIPTION_EXTERNAL_ID_PATH: existingSub.ExternalID,
	})
	if err != nil {
		return nil, err
	}

	subscriptionPlan := psql.CreateFreeSubscription(plan)
	subscriptionPlan.OrganizationID = orgID
	subscriptionPlan.ExternalID = &externalSubResponse.SubscriptionID
	subscriptionPlan.PaymentLink = externalSubResponse.PaymentLink

	return subscriptionPlan, nil
}

func (d dodoSubscriptionService) Verify(ctx context.Context, subscriptionID, orgID string) (*models.Subscription, error) {
	subscription, err := d.db.GetSubscriptionByIDAndOrg(ctx, subscriptionID, orgID)
	if err != nil {
		return nil, err
	}

	_, err = d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}

	externalSub, err := d.client.Subscriptions.Get(ctx, *subscription.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("error verifying subscription: %w", err)
	}

	if externalSub == nil {
		return nil, fmt.Errorf("error verifying subscription: invalid subscription")
	}

	if externalSub.Status == dodopayments.SubscriptionStatusActive {
		// update org and subscription
	}

	return subscription, nil
}
