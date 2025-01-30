package milli

import (
	"context"
	api "github.com/VapiAI/server-sdk-go"
	vapiclient "github.com/VapiAI/server-sdk-go/client"
	"github.com/VapiAI/server-sdk-go/option"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/voice_providers"
	"go.uber.org/zap"
)

type VAPIVoiceProvider struct {
	client *vapiclient.Client
	logger *zap.Logger
}

func NewVAPIVoiceProvider(config *models.VAPIConfig, logger *zap.Logger) integrations.VoiceProvider {
	client := vapiclient.NewClient(option.WithToken(config.APIKey))
	return &VAPIVoiceProvider{
		client: client,
		logger: logger,
	}
}

func (m *VAPIVoiceProvider) CreateCall(ctx context.Context, req integrations.CallRequest) (*integrations.CallResponse, error) {
	model, err := api.NewOpenAiModelModelFromString(req.Prompt.Model.String())
	if err != nil {
		return nil, err
	}
	registerCall := api.CreateCallDto{
		PhoneNumberId: nil,
		PhoneNumber:   nil,
		Customer: &api.CreateCustomerDto{
			Number: &req.ToPhone,
		},
		Assistant: &api.CreateAssistantDto{
			Model: &api.CreateAssistantDtoModel{
				OpenAiModel: &api.OpenAiModel{
					Messages: []*api.OpenAiMessage{
						{
							Content: &req.Prompt.PromptTmpl,
							Role:    api.OpenAiMessageRoleSystem,
						},
						{
							Content: &req.Prompt.HumanTmpl,
							Role:    api.OpenAiMessageRoleUser,
						},
					},
					Model: model,
				},
			},
		},
	}
	resp, err := m.client.Calls.Create(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return &integrations.CallResponse{
		CallID: resp.Id,
	}, nil
}
