package vapi

import (
	"context"
	api "github.com/VapiAI/server-sdk-go"
	vapiclient "github.com/VapiAI/server-sdk-go/client"
	"github.com/VapiAI/server-sdk-go/option"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

type VAPIVoiceProvider struct {
	client *vapiclient.Client
	logger *zap.Logger
}

func (m *VAPIVoiceProvider) Name() models.IntegrationType {
	return models.IntegrationTypeVOICEVAPI
}

func NewVAPIVoiceProvider(config *models.VAPIConfig, logger *zap.Logger) *VAPIVoiceProvider {
	client := vapiclient.NewClient(option.WithToken(config.APIKey))
	return &VAPIVoiceProvider{
		client: client,
		logger: logger,
	}
}

func (m *VAPIVoiceProvider) CreateCall(ctx context.Context, req models.CallRequest) (*models.CallResponse, error) {
	model, err := api.NewOpenAiModelModelFromString(req.GPTModel)
	if err != nil {
		return nil, err
	}

	messages := []*api.OpenAiMessage{}
	for _, message := range req.ChatMessages {
		if message.GetType() == llms.ChatMessageTypeSystem {
			messages = append(messages, &api.OpenAiMessage{
				Role:    api.OpenAiMessageRoleSystem,
				Content: utils.Ptr(message.GetContent()),
			})
		}

		if message.GetType() == llms.ChatMessageTypeHuman {
			messages = append(messages, &api.OpenAiMessage{
				Role:    api.OpenAiMessageRoleUser,
				Content: utils.Ptr(message.GetContent()),
			})
		}
	}

	registerCall := api.CreateCallDto{
		Name:          &req.ConversationID,
		PhoneNumberId: nil,
		PhoneNumber:   nil,
		Customer: &api.CreateCustomerDto{
			Number: &req.ToPhone,
		},
		Assistant: &api.CreateAssistantDto{
			Model: &api.CreateAssistantDtoModel{
				OpenAiModel: &api.OpenAiModel{
					Messages: messages,
					Model:    model,
				},
			},
		},
	}

	resp, err := m.client.Calls.Create(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return &models.CallResponse{
		CallID: resp.Id,
	}, nil
}
