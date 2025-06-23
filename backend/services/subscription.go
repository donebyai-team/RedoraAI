package services

import (
	"context"
	"fmt"
	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/datastore/psql"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/alerts"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"strings"
)

type SubscriptionService interface {
	CreatePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID, redirectURL string) (*models.Subscription, error)
	Verify(ctx context.Context, orgID string) (*models.Subscription, error)
	UpgradePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID string) (*models.Subscription, error)
	UpdateSubscriptionByExternalID(ctx context.Context, data []byte) (*models.Subscription, error)
}

type dodoSubscriptionService struct {
	db              datastore.Repository
	client          *dodopayments.Client
	logger          *zap.Logger
	productIDToPlan map[string]models.SubscriptionPlanType
	planToProductID map[models.SubscriptionPlanType]string
	notifier        alerts.AlertNotifier
}

const (
	liveProPlanID     = "pdt_h2V8lsZsVRO88Q19pCjU8"
	liveFounderPlanID = "pdt_1Arsz78Oy4pyi4MJbYeSb"

	testProPlanID     = "pdt_HcHajJOaRun8JfZwdBuNR"
	testFounderPlanID = "pdt_EEaOOJXUcgej57Jzl4O5w"
)

var livePlanToProductID = map[models.SubscriptionPlanType]string{
	models.SubscriptionPlanTypeFOUNDER: liveFounderPlanID,
	models.SubscriptionPlanTypePRO:     liveProPlanID,
}

var testPlanToProductID = map[models.SubscriptionPlanType]string{
	models.SubscriptionPlanTypeFOUNDER: testFounderPlanID,
	models.SubscriptionPlanTypePRO:     testProPlanID,
}

var liveProductIDToPlan = map[string]models.SubscriptionPlanType{
	liveFounderPlanID: models.SubscriptionPlanTypeFOUNDER,
	liveProPlanID:     models.SubscriptionPlanTypePRO,
}

var testProductIDToPlan = map[string]models.SubscriptionPlanType{
	testFounderPlanID: models.SubscriptionPlanTypeFOUNDER,
	testProPlanID:     models.SubscriptionPlanTypePRO,
}

var addOnIDMap = map[string]string{
	"adn_yIJQyUyFuX5tn2GYqqns5": "source",
	"adn_GQZ66G74wNJUH9yEuNxMG": "keyword",
}

func NewDodoSubscriptionService(db datastore.Repository, notifier alerts.AlertNotifier, token string, logger *zap.Logger, isTest bool) *dodoSubscriptionService {
	client := dodopayments.NewClient()
	if isTest {
		client.Options = []option.RequestOption{
			option.WithBearerToken(token),
			option.WithEnvironmentTestMode(),
		}
	} else {
		client.Options = []option.RequestOption{
			option.WithBearerToken(token),
			option.WithEnvironmentLiveMode(),
		}
	}

	client.Subscriptions = dodopayments.NewSubscriptionService(client.Options...)
	service := &dodoSubscriptionService{db: db, notifier: notifier, client: client, logger: logger}

	if isTest {
		service.planToProductID = testPlanToProductID
		service.productIDToPlan = testProductIDToPlan
	} else {
		service.planToProductID = livePlanToProductID
		service.productIDToPlan = liveProductIDToPlan
	}

	return service
}

func (d dodoSubscriptionService) UpgradePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID string) (*models.Subscription, error) {
	organization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}
	existingSub := organization.FeatureFlags.GetSubscription()
	if existingSub.ExternalID == nil || *existingSub.ExternalID == "" || !organization.FeatureFlags.IsSubscriptionActive() {
		return nil, fmt.Errorf("no subscription exits to upgrade")
	}

	if existingSub.PlanID == plan {
		return nil, fmt.Errorf("plan %s already exists", plan)
	}

	productId, ok := d.planToProductID[plan]
	if !ok {
		return nil, fmt.Errorf("invalid plan type: %s", plan)
	}

	externalSubExternal, err := d.client.Subscriptions.Get(ctx, *existingSub.ExternalID)
	if err != nil {
		d.logger.Error("error verifying subscription", zap.Error(err))
		return nil, fmt.Errorf("error getting existing subscription")
	}

	if externalSubExternal == nil || externalSubExternal.Status != dodopayments.SubscriptionStatusActive {
		return nil, fmt.Errorf("no subscription exits to upgrade")
	}

	// upgrade
	err = d.client.Subscriptions.ChangePlan(ctx, *existingSub.ExternalID, dodopayments.SubscriptionChangePlanParams{
		ProductID:            dodopayments.F(productId),
		ProrationBillingMode: dodopayments.F(dodopayments.SubscriptionChangePlanParamsProrationBillingModeProratedImmediately),
		Quantity:             dodopayments.F(int64(1)),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upgrade subscription: %w", err)
	}

	return d.Verify(ctx, orgID)
}

func (d dodoSubscriptionService) CreatePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID, returnURL string) (*models.Subscription, error) {
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
	if existingSub.ExternalID != nil && *existingSub.ExternalID != "" {
		return nil, fmt.Errorf("subscription already exists, please upgrade to change plan")
	}

	users, err := d.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no users found")
	}

	if existingSub.PlanID == plan {
		return nil, fmt.Errorf("plan %s already exists", plan)
	}

	productId, ok := d.planToProductID[plan]
	if !ok {
		return nil, fmt.Errorf("invalid plan type: %s", plan)
	}

	externalSubResponse, err := d.client.Subscriptions.New(ctx, dodopayments.SubscriptionNewParams{
		Billing: dodopayments.F(dodopayments.BillingAddressParam{
			City:    dodopayments.F("Bangalore"),
			Country: dodopayments.F(dodopayments.CountryCodeIn),
			State:   dodopayments.F("Karnataka"),
			Street:  dodopayments.F("Bannergetta"),
			Zipcode: dodopayments.F("560068"),
		}),
		ReturnURL: dodopayments.F(returnURL),
		Customer: dodopayments.F(dodopayments.CustomerRequestUnionParam(dodopayments.CustomerRequestParam{
			CreateNewCustomer: dodopayments.F(true),
			Email:             dodopayments.F(users[0].Email),
			Name:              dodopayments.F(organization.Name),
		})),
		Metadata: dodopayments.F(map[string]string{
			"organization_id": orgID,
		}),
		PaymentLink: dodopayments.F(true),
		ProductID:   dodopayments.F(productId),
		Quantity:    dodopayments.F(int64(1)),
	})

	if err != nil {
		d.logger.Error("error creating subscription", zap.Error(err))
		return nil, fmt.Errorf("error creating subscription")
	}

	if externalSubResponse == nil || externalSubResponse.SubscriptionID == "" {
		return nil, fmt.Errorf("error creating subscription: invalid subscription")
	}
	// update subscription in org
	err = d.db.UpdateOrganizationFeatureFlags(ctx, organization.ID, map[string]any{
		psql.FEATURE_FLAG_SUBSCRIPTION_EXTERNAL_ID_PATH: externalSubResponse.SubscriptionID,
	})
	if err != nil {
		return nil, err
	}

	subscriptionPlan := psql.CreateSubscriptionObject(plan)
	subscriptionPlan.OrganizationID = orgID
	subscriptionPlan.ExternalID = &externalSubResponse.SubscriptionID
	subscriptionPlan.PaymentLink = externalSubResponse.PaymentLink

	return subscriptionPlan, nil
}

func (d dodoSubscriptionService) UpdateSubscriptionByExternalID(ctx context.Context, data []byte) (*models.Subscription, error) {
	eventType := gjson.Get(string(data), "type").String()
	if !strings.Contains(eventType, "subscription") {
		return nil, nil
	}

	d.logger.Info("received subscription webhook", zap.String("data", string(data)))
	subscriptionId := gjson.Get(string(data), "data.subscription_id").String()
	if subscriptionId == "" {
		return nil, fmt.Errorf("invalid subscription id")
	}

	externalSubExternal, err := d.client.Subscriptions.Get(ctx, subscriptionId)
	if err != nil {
		d.logger.Error("error verifying subscription", zap.Error(err))
		return nil, fmt.Errorf("error verifying subscription")
	}

	if externalSubExternal == nil {
		return nil, fmt.Errorf("error verifying subscription: invalid subscription")
	}

	organizationID := externalSubExternal.Metadata["organization_id"]
	if organizationID == "" {
		return nil, fmt.Errorf("error verifying subscription: invalid subscription")
	}

	return d.Verify(ctx, organizationID)
}

// it should be called only when the plan is changed
func (d dodoSubscriptionService) Verify(ctx context.Context, orgID string) (*models.Subscription, error) {
	orgnization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}

	existingSub := orgnization.FeatureFlags.GetSubscription()
	if existingSub.ExternalID == nil || *existingSub.ExternalID == "" {
		return existingSub, nil
	}

	externalSubExternal, err := d.client.Subscriptions.Get(ctx, *existingSub.ExternalID)
	if err != nil {
		d.logger.Error("error verifying subscription", zap.Error(err))
		return nil, fmt.Errorf("error verifying subscription")
	}

	if externalSubExternal == nil {
		return nil, fmt.Errorf("error verifying subscription: invalid subscription")
	}

	plan, ok := d.productIDToPlan[externalSubExternal.ProductID]
	if !ok {
		return nil, fmt.Errorf("invalid product id to plan mapping: %s", plan)
	}

	//// if the plan is not changed, do nothing
	//if existingSub.PlanID == plan {
	//	return existingSub, nil
	//}

	d.logger.Info("subscription status received",
		zap.String("orgID", orgID),
		zap.String("external_id", externalSubExternal.SubscriptionID),
		zap.Any("subscription_status", externalSubExternal.Status))

	if externalSubExternal.Status == dodopayments.SubscriptionStatusActive {
		subscriptionPlan := psql.CreateSubscriptionObject(plan)
		subscriptionPlan.OrganizationID = orgID
		subscriptionPlan.ExternalID = &externalSubExternal.SubscriptionID

		for _, addOnID := range externalSubExternal.Addons {
			addOnType, ok := addOnIDMap[addOnID.AddonID]
			if !ok {
				return nil, fmt.Errorf("invalid addOn id: %s", plan)
			}
			if addOnType == "source" {
				subscriptionPlan.Metadata.MaxSources = subscriptionPlan.Metadata.MaxSources * int(addOnID.Quantity)
			} else if addOnType == "keyword" {
				subscriptionPlan.Metadata.MaxKeywords = subscriptionPlan.Metadata.MaxKeywords * int(addOnID.Quantity)
			}
		}

		subscriptionPlan.ExpiresAt = externalSubExternal.NextBillingDate
		subscriptionPlan.Status = models.SubscriptionStatusACTIVE
		err = d.db.UpdateOrganizationFeatureFlags(ctx, orgID, map[string]any{
			psql.FEATURE_FLAG_SUBSCRIPTION_PATH: subscriptionPlan,
		})
		if err != nil {
			d.logger.Error("error verifying subscription", zap.Error(err))
			return nil, err
		}
		d.logger.Info("subscription activated successfully",
			zap.String("orgID", orgID),
			zap.Any("subscription", subscriptionPlan))

		if existingSub.PlanID == models.SubscriptionPlanTypeFREE {
			go d.notifier.SendSubscriptionCreatedEmail(context.Background(), orgID)
		} else if !existingSub.ExpiresAt.Equal(subscriptionPlan.ExpiresAt) {
			go d.notifier.SendSubscriptionRenewedEmail(context.Background(), orgID)
		}

		return subscriptionPlan, nil
	}

	if externalSubExternal.Status == dodopayments.SubscriptionStatusPending {
		existingSub.Status = models.SubscriptionStatusCREATED
	} else if externalSubExternal.Status == dodopayments.SubscriptionStatusExpired {
		existingSub.Status = models.SubscriptionStatusEXPIRED
	} else {
		// move back to free plan to retry subscription
		existingSub.Status = models.SubscriptionStatusFAILED
		existingSub.PlanID = models.SubscriptionPlanTypeFREE
		existingSub.ExternalID = nil
		err = d.db.UpdateOrganizationFeatureFlags(ctx, orgID, map[string]any{
			psql.FEATURE_FLAG_SUBSCRIPTION_PATH: existingSub,
		})
		if err != nil {
			d.logger.Error("error verifying subscription", zap.Error(err))
			return nil, err
		}
		d.logger.Info("subscription cancelled successfully",
			zap.String("orgID", orgID),
			zap.Any("existing subscription", existingSub))
	}

	return existingSub, nil
}
