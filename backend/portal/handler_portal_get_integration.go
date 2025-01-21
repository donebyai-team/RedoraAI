package portal

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func (p *Portal) GetIntegration(ctx context.Context, c *connect.Request[pbportal.GetIntegrationRequest]) (*connect.Response[pbportal.Integration], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	if !actor.IsAdmin() {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("only admins can get integrations"))
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
