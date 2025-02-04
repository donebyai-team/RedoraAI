package models

import (
	"github.com/tmc/langchaingo/llms"
)

type CallRequest struct {
	ConversationID string
	FromPhone      string
	ToPhone        string
	ChatMessages   []llms.ChatMessage
	GPTModel       string
}

type CallResponse struct {
	CallID          string
	SessionID       string
	CallStatus      CallStatus
	CallEndedReason CallEndedReason
	RawResponse     string
}
