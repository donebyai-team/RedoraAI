package models

import (
	"github.com/openai/openai-go"
)

type CallRequest struct {
	ConversationID string
	FromPhone      string
	ToPhone        string
	ChatMessages   []openai.ChatCompletionMessageParamUnion
	GPTModel       string
}

type CallResponse struct {
	CallID          string
	SessionID       string
	CallStatus      CallStatus
	CallEndedReason CallEndedReason
	RawResponse     string
	Summary         string
	CallMessages    []CallMessage
	RecordingURL    string
}
