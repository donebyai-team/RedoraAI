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
	Verify(ctx context.Context, orgID, externalID string) (*models.Subscription, error)
	UpgradePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID string) (*models.Subscription, error)
	CancelPlan(ctx context.Context, orgID string) (*models.Subscription, error)
	UpdateSubscriptionByExternalID(ctx context.Context, data []byte) (*models.Subscription, error)
}

type dodoSubscriptionService struct {
	db              datastore.Repository
	client          *dodopayments.Client
	logger          *zap.Logger
	productIDToPlan map[string]models.SubscriptionPlanType
	planToProductID map[models.SubscriptionPlanType]string
	addOnIDMap      map[string]models.AddOnType
	checkoutLink    string
	notifier        alerts.AlertNotifier
}

const (
	liveStarterPlanID = "pdt_vvajqdsO8AjYlrz0PGWkp"
	liveProPlanID     = "pdt_h2V8lsZsVRO88Q19pCjU8"
	liveFounderPlanID = "pdt_1Arsz78Oy4pyi4MJbYeSb"

	testStarterPlanID = "pdt_8ZRMux3EotTaRBj7Y7iXA"
	testProPlanID     = "pdt_HcHajJOaRun8JfZwdBuNR"
	testFounderPlanID = "pdt_EEaOOJXUcgej57Jzl4O5w"
)

var livePlanToProductID = map[models.SubscriptionPlanType]string{
	models.SubscriptionPlanTypeFOUNDER: liveFounderPlanID,
	models.SubscriptionPlanTypePRO:     liveProPlanID,
	models.SubscriptionPlanTypeSTARTER: liveStarterPlanID,
}

var testPlanToProductID = map[models.SubscriptionPlanType]string{
	models.SubscriptionPlanTypeFOUNDER: testFounderPlanID,
	models.SubscriptionPlanTypePRO:     testProPlanID,
	models.SubscriptionPlanTypeSTARTER: testStarterPlanID,
}

var liveProductIDToPlan = map[string]models.SubscriptionPlanType{
	liveFounderPlanID: models.SubscriptionPlanTypeFOUNDER,
	liveProPlanID:     models.SubscriptionPlanTypePRO,
	liveStarterPlanID: models.SubscriptionPlanTypeSTARTER,
}

var testProductIDToPlan = map[string]models.SubscriptionPlanType{
	testFounderPlanID: models.SubscriptionPlanTypeFOUNDER,
	testProPlanID:     models.SubscriptionPlanTypePRO,
	testStarterPlanID: models.SubscriptionPlanTypeSTARTER,
}

var liveAddOnIDMap = map[string]models.AddOnType{
	"adn_yIJQyUyFuX5tn2GYqqns5": models.AddOnTypeSOURCE,
	"adn_GQZ66G74wNJUH9yEuNxMG": models.AddOnTypeKEYWORD,
}

var testAddOnIDMap = map[string]models.AddOnType{
	"adn_cQcg8NyHgyCgikswH5Uk7": models.AddOnTypeSOURCE,
	"adn_xargF2CfzXK0biAY4EDgy": models.AddOnTypeKEYWORD,
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
		service.addOnIDMap = testAddOnIDMap
		service.checkoutLink = "https://test.checkout.dodopayments.com"
	} else {
		service.planToProductID = livePlanToProductID
		service.productIDToPlan = liveProductIDToPlan
		service.addOnIDMap = liveAddOnIDMap
		service.checkoutLink = "https://checkout.dodopayments.com"
	}

	return service
}

func (d dodoSubscriptionService) CancelPlan(ctx context.Context, orgID string) (*models.Subscription, error) {
	organization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}
	existingSub := organization.FeatureFlags.GetSubscription()
	if existingSub.ID == "" || !organization.FeatureFlags.IsSubscriptionActive() {
		return nil, fmt.Errorf("no active subscription exits to cancel")
	}

	_, err = d.client.Subscriptions.Update(ctx, existingSub.ID, dodopayments.SubscriptionUpdateParams{
		Status: dodopayments.F(dodopayments.SubscriptionStatusCancelled),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return d.Verify(ctx, orgID, existingSub.ID)
}

func (d dodoSubscriptionService) UpgradePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID string) (*models.Subscription, error) {
	organization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}
	existingSub := organization.FeatureFlags.GetSubscription()
	if existingSub.ID == "" || !organization.FeatureFlags.IsSubscriptionActive() {
		return nil, fmt.Errorf("no subscription exits to upgrade")
	}

	if existingSub.PlanID == plan {
		return nil, fmt.Errorf("plan %s already exists", plan)
	}

	productId, ok := d.planToProductID[plan]
	if !ok {
		return nil, fmt.Errorf("invalid plan type: %s", plan)
	}

	externalSubExternal, err := d.client.Subscriptions.Get(ctx, existingSub.ID)
	if err != nil {
		d.logger.Error("error verifying subscription", zap.Error(err))
		return nil, fmt.Errorf("error getting existing subscription")
	}

	if externalSubExternal == nil || externalSubExternal.Status != dodopayments.SubscriptionStatusActive {
		return nil, fmt.Errorf("no subscription exits to upgrade")
	}

	// upgrade
	err = d.client.Subscriptions.ChangePlan(ctx, existingSub.ID, dodopayments.SubscriptionChangePlanParams{
		ProductID:            dodopayments.F(productId),
		ProrationBillingMode: dodopayments.F(dodopayments.SubscriptionChangePlanParamsProrationBillingModeProratedImmediately),
		Quantity:             dodopayments.F(int64(1)),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upgrade subscription: %w", err)
	}

	return d.Verify(ctx, orgID, existingSub.ID)
}

func (d dodoSubscriptionService) CreatePlan(ctx context.Context, plan models.SubscriptionPlanType, orgID, returnURL string) (*models.Subscription, error) {
	organization, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}

	existingSub := organization.FeatureFlags.GetSubscription()
	if existingSub.ID != "" {
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

	d.logger.Info("creating subscription", zap.String("orgID", orgID), zap.String("productID", productId))
	//externalSubResponse, err := d.client.Subscriptions.New(ctx, dodopayments.SubscriptionNewParams{
	//	Billing: dodopayments.F(dodopayments.BillingAddressParam{
	//		City:    dodopayments.F("Bangalore"),
	//		Country: dodopayments.F(dodopayments.CountryCodeIn),
	//		State:   dodopayments.F("Karnataka"),
	//		Street:  dodopayments.F("Bannergetta"),
	//		Zipcode: dodopayments.F("560068"),
	//	}),
	//	ReturnURL: dodopayments.F(returnURL),
	//	Customer: dodopayments.F(dodopayments.CustomerRequestUnionParam(dodopayments.CustomerRequestParam{
	//		CreateNewCustomer: dodopayments.F(true),
	//		Email:             dodopayments.F(users[0].Email),
	//		Name:              dodopayments.F(organization.Name),
	//	})),
	//	Metadata: dodopayments.F(map[string]string{
	//		"organization_id": orgID,
	//	}),
	//	PaymentLink: dodopayments.F(true),
	//	ProductID:   dodopayments.F(productId),
	//	Quantity:    dodopayments.F(int64(1)),
	//})
	//
	//if err != nil {
	//	d.logger.Error("error creating subscription", zap.Error(err))
	//	return nil, fmt.Errorf("error creating subscription")
	//}
	//
	//if externalSubResponse == nil || externalSubResponse.SubscriptionID == "" {
	//	return nil, fmt.Errorf("error creating subscription: invalid subscription")
	//}
	//// update subscription in org
	//err = d.db.UpdateOrganizationFeatureFlags(ctx, organization.ID, map[string]any{
	//	psql.FEATURE_FLAG_SUBSCRIPTION_EXTERNAL_ID_PATH: externalSubResponse.SubscriptionID,
	//})
	//if err != nil {
	//	return nil, err
	//}

	subscriptionPlan := psql.CreateSubscriptionObject(plan)
	subscriptionPlan.OrganizationID = orgID
	subscriptionPlan.PaymentLink = fmt.Sprintf("%s/buy/%s?metadata_organization_id=%s&redirect_url=%s", d.checkoutLink, productId, orgID, returnURL)
	d.logger.Info("subscription created successfully", zap.String("orgID", orgID), zap.Any("subscription", subscriptionPlan))

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
		return nil, fmt.Errorf("error verifying subscription: invalid subscription received")
	}

	organizationID := externalSubExternal.Metadata["organization_id"]
	if organizationID == "" {
		return nil, fmt.Errorf("error verifying subscription: no organization_id")
	}

	return d.Verify(ctx, organizationID, subscriptionId)
}

func (d dodoSubscriptionService) Verify(ctx context.Context, orgID, externalID string) (*models.Subscription, error) {
	org, err := d.db.GetOrganizationById(ctx, orgID)
	if err != nil {
		return nil, err
	}

	d.logger.Info("verifying subscription", zap.String("orgID", orgID))

	externalSub, err := d.client.Subscriptions.Get(ctx, externalID)
	if err != nil || externalSub == nil {
		d.logger.Error("failed to fetch external subscription", zap.Error(err))
		return nil, fmt.Errorf("error verifying subscription")
	}

	organizationIdFromSub := externalSub.Metadata["organization_id"]
	if organizationIdFromSub != orgID {
		return nil, fmt.Errorf("error verifying subscription: invalid organization_id")
	}

	plan, ok := d.productIDToPlan[externalSub.ProductID]
	if !ok {
		return nil, fmt.Errorf("invalid product id to plan mapping: %s", externalSub.ProductID)
	}

	sub := org.FeatureFlags.GetSubscription()
	d.logger.Info("subscription status received", zap.String("orgID", orgID), zap.Any("external_subscription", externalSub))

	switch externalSub.Status {
	case dodopayments.SubscriptionStatusActive:
		return d.handleActiveSubscription(ctx, orgID, sub, externalSub, plan)
	case dodopayments.SubscriptionStatusPending:
		sub.Status = models.SubscriptionStatusCREATED
	case dodopayments.SubscriptionStatusExpired:
		sub.Status = models.SubscriptionStatusEXPIRED
	case dodopayments.SubscriptionStatusCancelled:
		return d.changeOrDowngradePlan(ctx, orgID, sub, models.SubscriptionPlanTypeFREE, models.SubscriptionStatusCANCELLED)
	default:
		return d.changeOrDowngradePlan(ctx, orgID, sub, sub.PlanID, models.SubscriptionStatusFAILED)
	}

	d.logger.Info("subscription updated", zap.String("orgID", orgID), zap.String("status", sub.Status.String()))
	return sub, nil
}

func (d dodoSubscriptionService) handleActiveSubscription(
	ctx context.Context,
	orgID string,
	oldSub *models.Subscription,
	externalSub *dodopayments.Subscription,
	plan models.SubscriptionPlanType,
) (*models.Subscription, error) {
	oldSubExpiresAt := oldSub.ExpiresAt
	oldPlanID := oldSub.PlanID

	newPlan := psql.CreateSubscriptionObject(plan)
	oldSub.Metadata = newPlan.Metadata
	oldSub.PlanID = plan
	oldSub.OrganizationID = orgID
	oldSub.ExternalID = nil
	oldSub.ID = externalSub.SubscriptionID
	oldSub.ExpiresAt = externalSub.NextBillingDate
	oldSub.Status = models.SubscriptionStatusACTIVE
	if len(oldSub.Metadata.AddOns) == 0 {
		oldSub.Metadata.AddOns = map[models.AddOnType]int{}
	}

	for _, addOn := range externalSub.Addons {
		switch addOnType := d.addOnIDMap[addOn.AddonID]; addOnType {
		case models.AddOnTypeSOURCE:
			oldSub.Metadata.AddOns[models.AddOnTypeSOURCE] = int(addOn.Quantity)
		case models.AddOnTypeKEYWORD:
			oldSub.Metadata.AddOns[models.AddOnTypeKEYWORD] = int(addOn.Quantity)
		default:
			return nil, fmt.Errorf("invalid addOn id: %s", addOn.AddonID)
		}
	}

	if err := d.db.UpdateOrganizationFeatureFlags(ctx, orgID, map[string]any{
		psql.FEATURE_FLAG_SUBSCRIPTION_PATH: oldSub,
	}); err != nil {
		d.logger.Error("error updating feature flags", zap.Error(err))
		return nil, err
	}

	if err := d.db.UpdateProjectIsActive(ctx, orgID, true); err != nil {
		d.logger.Error("error activating project", zap.Error(err))
		return nil, err
	}

	d.logger.Info("subscription activated", zap.String("orgID", orgID), zap.Any("subscription", oldSub))

	// Send email notifications
	if oldPlanID == models.SubscriptionPlanTypeFREE {
		go d.notifier.SendSubscriptionCreatedEmail(context.Background(), orgID)
	} else if !oldSubExpiresAt.Equal(externalSub.NextBillingDate) {
		go d.notifier.SendSubscriptionRenewedEmail(context.Background(), orgID)
	}

	return oldSub, nil
}

func (d dodoSubscriptionService) changeOrDowngradePlan(
	ctx context.Context,
	orgID string,
	sub *models.Subscription,
	planToChange models.SubscriptionPlanType,
	status models.SubscriptionStatus,
) (*models.Subscription, error) {
	sub.PlanID = planToChange
	sub.ExternalID = nil
	sub.ID = ""

	if err := d.db.UpdateOrganizationFeatureFlags(ctx, orgID, map[string]any{
		psql.FEATURE_FLAG_SUBSCRIPTION_PATH: sub,
	}); err != nil {
		d.logger.Error("error downgrading to free plan", zap.Error(err))
		return nil, err
	}

	// make sure we do it after updating feature flags
	sub.Status = status

	if sub.Status == models.SubscriptionStatusCANCELLED {
		go d.notifier.SendSubscriptionCancelledEmail(context.Background(), orgID)
	}

	d.logger.Info(fmt.Sprintf("downgraded to %s plan", planToChange.String()), zap.String("orgID", orgID), zap.String("status", status.String()))
	return sub, nil
}
