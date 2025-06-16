package portal

import (
	"connectrpc.com/connect"
	"context"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *Portal) UpgradeSubscription(ctx context.Context, c *connect.Request[pbportal.UpgradeSubscriptionRequest]) (*connect.Response[pbcore.Subscription], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	subscription, err := p.subscriptionService.UpgradePlan(ctx, c.Msg.Plan.ToModel(), actor.OrganizationID)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	return connect.NewResponse(new(pbcore.Subscription).FromModel(subscription)), nil
}

func (p *Portal) VerifySubscription(ctx context.Context, c *connect.Request[pbportal.VerifySubscriptionRequest]) (*connect.Response[pbcore.Subscription], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	subscription, err := p.subscriptionService.Verify(ctx, actor.OrganizationID)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	return connect.NewResponse(new(pbcore.Subscription).FromModel(subscription)), nil
}

func (p *Portal) InitiateSubscription(ctx context.Context, c *connect.Request[pbportal.InitiateSubscriptionRequest]) (*connect.Response[pbportal.InitiateSubscriptionResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := p.subscriptionService.CreatePlan(ctx, c.Msg.Plan.ToModel(), actor.OrganizationID, c.Msg.RedirectUrl)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	return connect.NewResponse(&pbportal.InitiateSubscriptionResponse{
		PaymentLink: resp.PaymentLink,
	}), nil
}
