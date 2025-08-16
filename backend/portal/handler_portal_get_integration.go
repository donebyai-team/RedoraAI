package portal

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

func (p *Portal) GetIntegration(ctx context.Context, c *connect.Request[pbportal.GetIntegrationRequest]) (*connect.Response[pbportal.Integrations], error) {
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
	integrations, err := p.db.GetIntegrationByOrgAndType(ctx, actor.OrganizationID, c.Msg.Type.ToModel())
	if err != nil {
		return nil, fmt.Errorf("get integration: %w", err)
	}

	var result []*pbportal.Integration

	for _, i := range integrations {
		protoInt := p.protoIntegration(i)
		if protoInt != nil {
			result = append(result, protoInt)
		}
	}

	return connect.NewResponse(&pbportal.Integrations{
		Integrations: result,
	}), nil
}
func (p *Portal) UpdateIntegration(ctx context.Context, c *connect.Request[pbportal.UpdateIntegrationRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	integration, err := p.db.GetIntegrationById(ctx, c.Msg.Id)
	if err != nil {
		return nil, err
	}

	if integration.OrganizationID != actor.OrganizationID {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("only admins can revoke integrations"))
	}

	if integration.Type == models.IntegrationTypeREDDITDMLOGIN {
		if redditDetails, ok := c.Msg.Details.(*pbportal.UpdateIntegrationRequest_Reddit); ok {
			detailsToBeUpdated := redditDetails.Reddit
			existingDetails := integration.GetRedditDMLoginConfig()
			if existingDetails.Alpha2CountryCode != detailsToBeUpdated.Alpha2CountryCode {
				existingDetails.Alpha2CountryCode = detailsToBeUpdated.Alpha2CountryCode

				integration = models.SetIntegrationType(integration, models.IntegrationTypeREDDITDMLOGIN, existingDetails)

				if _, err = p.db.UpsertIntegration(ctx, integration); err != nil {
					p.logger.Warn("failed to update integration", zap.Error(err), zap.String("integration_id", c.Msg.Id))
					return nil, err
				}
				p.logger.Info("updated integration", zap.String("integration_id", c.Msg.Id))
			}
		}
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}
func (p *Portal) RevokeIntegration(ctx context.Context, c *connect.Request[pbportal.RevokeIntegrationRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	integration, err := p.db.GetIntegrationById(ctx, c.Msg.Id)
	if err != nil {
		return nil, err
	}

	if integration.OrganizationID != actor.OrganizationID {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("only admins can revoke integrations"))
	}

	integration.State = models.IntegrationStateAUTHREVOKED
	_, err = p.db.UpsertIntegration(ctx, integration)
	if err != nil {
		return nil, err
	}

	go p.alertNotifier.SendIntegrationRevoked(context.Background(), integration.OrganizationID, *integration.ReferenceID, integration.GetIntegrationStatus(true))

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) protoIntegration(integration *models.Integration) *pbportal.Integration {
	switch integration.Type {
	case models.IntegrationTypeREDDIT:
		config := integration.GetRedditConfig()
		resp := &pbportal.Integration{
			Id:             integration.ID,
			OrganizationId: integration.OrganizationID,
			Type:           pbportal.IntegrationType_INTEGRATION_TYPE_REDDIT,
			Status:         mapIntegrationState(integration.State),
		}
		if integration.ReferenceID != nil {
			resp.Details = &pbportal.Integration_Reddit{
				Reddit: &pbportal.RedditIntegration{
					UserName: *integration.ReferenceID,
					Reason:   integration.GetIntegrationStatus(config.IsUserOldEnough(2)),
				},
			}
		}
		return resp
	case models.IntegrationTypeREDDITDMLOGIN:
		config := integration.GetRedditDMLoginConfig()
		resp := &pbportal.Integration{
			Id:             integration.ID,
			OrganizationId: integration.OrganizationID,
			Type:           pbportal.IntegrationType_INTEGRATION_TYPE_REDDIT_DM_LOGIN,
			Status:         mapIntegrationState(integration.State),
		}
		if integration.ReferenceID != nil {
			resp.Details = &pbportal.Integration_Reddit{
				Reddit: &pbportal.RedditIntegration{
					UserName:          *integration.ReferenceID,
					Alpha2CountryCode: config.Alpha2CountryCode,
					Reason:            integration.GetIntegrationStatus(true),
				},
			}
		}
		return resp
	//case models.IntegrationTypeGOOGLE:
	//	return p.resolveGoogleIntegration(ctx, integration)
	default:
		p.logger.Error("unsupported integration type", zap.String("integration", integration.Type.String()))
		return nil
	}
}

func (p *Portal) GetIntegrations(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.Integrations], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("auth context error: %w", err))
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("fetch integrations: %w", err))
	}

	var result []*pbportal.Integration

	for _, i := range integrations {
		protoInt := p.protoIntegration(i)
		if protoInt != nil {
			result = append(result, protoInt)
		}
	}

	return connect.NewResponse(&pbportal.Integrations{
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
