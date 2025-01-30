package milli

import (
	"context"
	"github.com/shank318/doota/voice_providers"
	"github.com/shank318/doota/voice_providers/milli/api"
)

type MilliVoiceProvider struct {
	client *api.Client
}

func (m *MilliVoiceProvider) CreateCall(ctx context.Context, req integrations.CallRequest) (*integrations.CallResponse, error) {
	registerCall := api.RegisterCallRequest{
		FromPhone:               req.FromPhone,
		ToPhone:                 req.ToPhone,
		IncludeMetadataInPrompt: req.IncludeMetadataInPrompt,
		Metadata:                req.Metadata,
	}
	resp, err := m.client.CreateCall(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return &integrations.CallResponse{
		CallID:    resp.CallID,
		SessionID: resp.SessionID,
	}, nil
}
