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
	"strings"
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

type ServerMessage struct {
	Message struct {
		Timestamp   int64  `json:"timestamp"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		EndedReason string `json:"endedReason"`
		Call        struct {
			Id string `json:"id"`
		} `json:"call"`
	} `json:"message"`
}

func (m *VAPIVoiceProvider) HandleWebhook(ctx context.Context, req []byte) (*models.CallResponse, error) {
	serverMessage := ServerMessage{}
	err := json.Unmarshal(req, &serverMessage)
	if err != nil {
		return nil, err
	}
	return m.handleCallStatusUpdate(ctx, serverMessage.Message.Call.Id)
}

func (m *VAPIVoiceProvider) handleCallStatusUpdate(ctx context.Context, callID string) (*models.CallResponse, error) {
	if callID == "" {
		m.logger.Warn("handleCallStatusUpdate, no call found")
		return nil, nil
	}
	call, err := m.client.Calls.Get(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("could not get call '%s': %w", callID, err)
	}

	return transformCallToCallResponse(call), nil
}

var assistantErrors = []string{"failed", "error", "invalid", "not-found", "shutdown", "blocked"}

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

	if call.Analysis != nil {
		callResponse.Summary = *call.Analysis.Summary
	}

	if call.Artifact != nil {
		callResponse.RecordingURL = *call.Artifact.RecordingUrl
	}

	for _, message := range call.Messages {
		if message.SystemMessage != nil {
			callResponse.CallMessages = append(callResponse.CallMessages, models.CallMessage{
				SystemMessage: &models.SystemMessage{
					Role:             message.SystemMessage.Role,
					Message:          message.SystemMessage.Message,
					Time:             message.SystemMessage.Time,
					SecondsFromStart: message.SystemMessage.SecondsFromStart,
				},
			})
		} else if message.UserMessage != nil {
			callResponse.CallMessages = append(callResponse.CallMessages, models.CallMessage{
				UserMessage: &models.UserMessage{
					Role:             message.UserMessage.Role,
					Message:          message.UserMessage.Message,
					Time:             message.UserMessage.Time,
					SecondsFromStart: message.UserMessage.SecondsFromStart,
					EndTime:          message.UserMessage.EndTime,
					Duration:         message.UserMessage.Duration,
				},
			})
		} else if message.BotMessage != nil {
			callResponse.CallMessages = append(callResponse.CallMessages, models.CallMessage{
				BotMessage: &models.BotMessage{
					Role:             message.BotMessage.Role,
					Message:          message.BotMessage.Message,
					Time:             message.BotMessage.Time,
					EndTime:          message.BotMessage.EndTime,
					SecondsFromStart: message.BotMessage.SecondsFromStart,
					Source:           message.BotMessage.Source,
					Duration:         message.BotMessage.Duration,
				},
			})
		}
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

	// if still unknown, check for error reasons
	if call.EndedReason != nil && callResponse.CallEndedReason == models.CallEndedReasonUNKNOWN {
		for _, reason := range assistantErrors {
			if strings.Contains(strings.ToLower(string(callResponse.CallEndedReason)), strings.ToLower(reason)) {
				callResponse.CallEndedReason = models.CallEndedReasonASSISTANTERROR
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
				Role:    api.OpenAiMessageRoleAssistant,
				Content: utils.Ptr(message.GetContent()),
			})
		}
	}

	toPhone, err := utils.ConvertToE164(req.ToPhone, "IN")
	if err != nil {
		return nil, fmt.Errorf("could not ToPhone convert to E164: %w", err)
	}

	registerCall := api.CreateCallDto{
		Name:          &req.ConversationID,
		PhoneNumberId: &req.FromPhone,
		Customer: &api.CreateCustomerDto{
			Number: &toPhone,
		},
		Assistant: &api.CreateAssistantDto{
			EndCallPhrases: []string{"bye", "googbye", "call again later", "have a nice day", "thanks you bye", "see you"},
			EndCallMessage: utils.Ptr("bye! have a nice day"),
			Voice: &api.CreateAssistantDtoVoice{
				ElevenLabsVoice: &api.ElevenLabsVoice{
					VoiceId: &api.ElevenLabsVoiceId{
						String: "pzxut4zZz4GImZNlqQ3H",
					},
				},
			},
			Model: &api.CreateAssistantDtoModel{
				OpenAiModel: &api.OpenAiModel{
					Messages: messages,
					Model:    model,
				},
			},
			MaxDurationSeconds: utils.Ptr(300.0),
			ServerMessages: []api.CreateAssistantDtoServerMessagesItem{
				api.CreateAssistantDtoServerMessagesItemEndOfCallReport,
				api.CreateAssistantDtoServerMessagesItemStatusUpdate,
			},
			Server: &api.Server{
				Url: fmt.Sprintf("https://592b-122-171-23-192.ngrok-free.app/webhook/vana/call_status/%s", req.ConversationID),
			},
		},
	}

	resp, err := m.client.Calls.Create(ctx, &registerCall)
	if err != nil {
		return nil, err
	}

	return transformCallToCallResponse(resp), nil
}
