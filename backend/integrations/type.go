package integrations

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/milli"
	"github.com/shank318/doota/integrations/vapi"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type VoiceProvider interface {
	Name() models.IntegrationType
	CreateCall(ctx context.Context, req models.CallRequest) (*models.CallResponse, error)
	HandleWebhook(ctx context.Context, req []byte) (*models.CallResponse, error)
}

type Factory struct {
	db     datastore.Repository
	logger *zap.Logger
}

func NewFactory(db datastore.Repository, logger *zap.Logger) *Factory {
	return &Factory{
		db:     db,
		logger: logger,
	}
}

func (c *Factory) NewVoiceClient(ctx context.Context, orgID string) (VoiceProvider, error) {
	integrations, err := c.db.GetIntegrationsByOrgID(ctx, orgID)
	if err != nil {
		return nil, errors.Wrap(err, "service: failed to get integration by org")
	}

	for _, integration := range integrations {
		if integration.Type == models.IntegrationTypeVOICEVAPI {
			return vapi.NewVAPIVoiceProvider(integration.GetVAPIConfig(), c.logger), nil
		}

		if integration.Type == models.IntegrationTypeVOICEMILLIS {
			return milli.NewMilliVoiceProvider(), nil
		}
	}
	return nil, fmt.Errorf("no voice integration found for orgID: %s", orgID)
}
