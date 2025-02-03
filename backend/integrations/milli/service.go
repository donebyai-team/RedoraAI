package milli

import (
	"context"
	"github.com/shank318/doota/integrations/milli/api"
	"github.com/shank318/doota/models"
)

type MilliVoiceProvider struct {
	client *api.Client
}

func NewMilliVoiceProvider() *MilliVoiceProvider {
	return &MilliVoiceProvider{}
}

func (m *MilliVoiceProvider) Name() models.IntegrationType {
	return models.IntegrationTypeVOICEMILLIS
}

func (m *MilliVoiceProvider) CreateCall(ctx context.Context, req models.CallRequest) (*models.CallResponse, error) {
	registerCall := api.RegisterCallRequest{
		FromPhone: req.FromPhone,
		ToPhone:   req.ToPhone,
	}
	resp, err := m.client.CreateCall(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return &models.CallResponse{
		CallID:    resp.CallID,
		SessionID: resp.SessionID,
	}, nil
}
