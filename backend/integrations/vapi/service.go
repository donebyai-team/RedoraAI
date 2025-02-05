package vapi

import (
	"context"
	"encoding/json"
	"fmt"
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

func (m *VAPIVoiceProvider) HandleWebhook(ctx context.Context, req []byte) (*models.CallResponse, error) {
	serverMessage := api.ServerMessage{}
	err := json.Unmarshal(req, &serverMessage)
	if err != nil {
		return nil, err
	}

	if serverMessage.Message.ServerMessageEndOfCallReport != nil {
		return m.handleEndOfCallReport(ctx, serverMessage.Message.ServerMessageEndOfCallReport)
	} else if serverMessage.Message.ServerMessageStatusUpdate != nil {
		return m.handleCallStatusUpdate(ctx, serverMessage.Message.ServerMessageStatusUpdate)
	}
	return nil, err
}

func (m *VAPIVoiceProvider) handleCallStatusUpdate(ctx context.Context, report *api.ServerMessageStatusUpdate) (*models.CallResponse, error) {
	if report.Call == nil {
		m.logger.Warn("handleCallStatusUpdate, no call found")
		return nil, nil
	}
	call, err := m.client.Calls.Get(ctx, report.Call.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get call '%s': %w", report.Call.Id, err)
	}

	return transformCallToCallResponse(call), nil
}

func (m *VAPIVoiceProvider) handleEndOfCallReport(ctx context.Context, report *api.ServerMessageEndOfCallReport) (*models.CallResponse, error) {
	if report.Call == nil {
		m.logger.Warn("handleEndOfCallReport, no call found")
		return nil, nil
	}
	call, err := m.client.Calls.Get(ctx, report.Call.Id)
	if err != nil {
		return nil, fmt.Errorf("could not get call '%s': %w", report.Call.Id, err)
	}

	return transformCallToCallResponse(call), nil
}

func transformCallToCallResponse(call *api.Call) *models.CallResponse {
	callResponse := &models.CallResponse{
		CallID:          call.Id,
		CallEndedReason: models.CallEndedReasonUNKNOWN,
		CallStatus:      models.CallStatusUNKNOWN,
		RawResponse:     call.String(),
	}

	if call.EndedReason != nil {
		for key, reasons := range endReasonMapping {
			for _, reason := range reasons {
				if reason == *call.EndedReason {
					callResponse.CallEndedReason = key
				}
			}
		}
	}

	for key, statuses := range callStatusMapping {
		for _, status := range statuses {
			if status == *call.Status {
				callResponse.CallStatus = key
			}
		}
	}

	return callResponse
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
			ServerMessages: []api.CreateAssistantDtoServerMessagesItem{
				api.CreateAssistantDtoServerMessagesItemEndOfCallReport,
				api.CreateAssistantDtoServerMessagesItemStatusUpdate,
			},
			Server: &api.Server{
				Url: "",
				Headers: map[string]interface{}{
					"Content-Type":    "application/json",
					"conversation_id": req.ConversationID,
				},
			},
		},
	}

	resp, err := m.client.Calls.Create(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return transformCallToCallResponse(resp), nil
}
