package portal

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	pbreddit "github.com/shank318/doota/pb/doota/reddit/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (p *Portal) GetIntegrationByOrgId(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.IntegrationByOrgIdResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth context error: %w", err)
	}

	logging.Logger(ctx, p.logger).Info("get integrations by org ID", zap.String("org_id", actor.OrganizationID))

	integrations, err := p.db.GetIntegrationsByOrgID(ctx, actor.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("fetch integrations: %w", err)
	}

	var redditIntegrations []*pbreddit.Integration

	for _, i := range integrations {
		// Parse plain text config to extract username
		var cfg models.RedditConfig
		if err := json.Unmarshal([]byte(i.PlainTextConfig), &cfg); err != nil {
			return nil, fmt.Errorf("parse reddit config: %w", err)
		}

		redditIntegrations = append(redditIntegrations, &pbreddit.Integration{
			UserName: cfg.UserName,
		})

	}
	return connect.NewResponse(&pbportal.IntegrationByOrgIdResponse{
		Reddit: redditIntegrations,
	}), nil
}
