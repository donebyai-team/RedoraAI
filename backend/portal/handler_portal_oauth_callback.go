package portal

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/portal/state"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

type IntegrationHandler func(ctx context.Context, p *Portal, code string, organizationID string, oauthState *state.State) error

var integrations = map[pbportal.IntegrationType]IntegrationHandler{
	pbportal.IntegrationType_INTEGRATION_TYPE_REDDIT: handleRedditOauth,
}

func (p *Portal) OauthCallback(ctx context.Context, c *connect.Request[pbportal.OauthCallbackRequest]) (*connect.Response[pbportal.OauthCallbackResponse], error) {
	//actor, err := p.gethAuthContext(ctx)
	//if err != nil {
	//	return nil, err
	//}
	authState, err := p.validateState(c.Msg.State)
	if err != nil {
		return nil, fmt.Errorf("error validating state: %w", err)
	}

	handler, ok := integrations[authState.IntegrationType]
	if !ok {
		return nil, fmt.Errorf("unknown %s handler: %w", authState.IntegrationType.String(), err)
	}
	if err = handler(ctx, p, c.Msg.GetExternalCode(), "e250ced8-7441-4805-b9dd-2686d9492c4f", authState); err != nil {
		return nil, fmt.Errorf("error handling %s: %w", authState.IntegrationType.String(), err)
	}

	return connect.NewResponse(&pbportal.OauthCallbackResponse{RedirectUrl: authState.RedirectUri}), nil
}

func handleRedditOauth(ctx context.Context, p *Portal, code string, organizationID string, oauthState *state.State) error {
	token, err := p.redditOauthClient.Authorize(ctx, code)
	if err != nil {
		return err
	}

	integration := &models.Integration{
		OrganizationID: organizationID,
		State:          models.IntegrationStateACTIVE,
	}

	out := &models.RedditConfig{AccessToken: token}
	integrationType := models.SetIntegrationType(integration, models.IntegrationTypeREDDIT, out)

	integration, err = p.db.UpsertIntegration(ctx, integrationType)
	if err != nil {
		return fmt.Errorf("upsert integration: %w", err)
	}

	logging.Logger(ctx, p.logger).Info("reddit integration created",
		zap.String("organization_id", integration.OrganizationID),
		zap.String("integration_id", integration.ID),
	)
	return nil
}

func (p *Portal) validateState(state string) (*state.State, error) {
	s, err := p.authStateStore.GetState(state)
	if err != nil {
		return nil, fmt.Errorf("unable to get state: %w", err)
	}

	if s.HasExpired() {
		return nil, fmt.Errorf("state expired")
	}

	return s, nil
}
