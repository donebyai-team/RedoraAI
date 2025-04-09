package portal

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	pbreddit "github.com/shank318/doota/pb/doota/reddit/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

func (p *Portal) GetIntegration(ctx context.Context, c *connect.Request[pbportal.GetIntegrationRequest]) (*connect.Response[pbportal.Integration], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	logging.Logger(ctx, p.logger).Info("get integration",
		zap.Stringer("integration_type", c.Msg.Type),
	)
	integration, err := p.db.GetIntegrationByOrgAndType(ctx, actor.OrganizationID, c.Msg.Type.ToModel())
	if err != nil {
		return nil, fmt.Errorf("get integration: %w", err)
	}
	return p.protoIntegration(ctx, integration)
}

func (p *Portal) protoIntegration(ctx context.Context, integration *models.Integration) (*connect.Response[pbportal.Integration], error) {
	switch integration.Type {
	//case models.IntegrationTypeMICROSOFT:
	//	return p.resolveMicrosoftIntegration(ctx, integration)
	//case models.IntegrationTypeGOOGLE:
	//	return p.resolveGoogleIntegration(ctx, integration)
	default:
		panic(fmt.Errorf("unsupported integration type: %s", integration.Type))
	}
}

func (p *Portal) GetIntegrations(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.IntegrationsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth context error: %w", err)
	}

	logging.Logger(ctx, p.logger).Info("fetching integrations for organization",
		zap.String("org_id", actor.OrganizationID),
	)

	integrations, err := p.db.GetIntegrationsByOrgID(ctx, actor.OrganizationID)

	if err != nil {
		logging.Logger(ctx, p.logger).Error("failed to fetch integrations",
			zap.String("org_id", actor.OrganizationID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("fetch integrations: %w", err)
	}

	var result []*pbportal.Integration

	for _, i := range integrations {
		redditCfg := i.GetRedditConfig()
		result = append(result, &pbportal.Integration{
			Id:             i.ID,
			OrganizationId: i.OrganizationID,
			Type:           pbportal.IntegrationType_INTEGRATION_TYPE_REDDIT,
			Status:         mapIntegrationState(i.State),
			Details: &pbportal.Integration_Reddit{
				Reddit: &pbreddit.Integration{
					UserName: redditCfg.UserName,
				},
			},
		})
	}

	return connect.NewResponse(&pbportal.IntegrationsResponse{
		Integrations: result,
	}), nil
}

func mapIntegrationState(state models.IntegrationState) pbportal.IntegrationState {
	enumKey := "INTEGRATION_STATE_" + strings.ToUpper(string(state))
	if val, ok := pbportal.IntegrationState_value[enumKey]; ok {
		return pbportal.IntegrationState(val)
	}
	return pbportal.IntegrationState_INTEGRATION_STATE_UNSPECIFIED
}
