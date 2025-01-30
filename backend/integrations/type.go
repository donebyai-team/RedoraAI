package integrations

import (
	"github.com/shank318/doota/ai"
	"golang.org/x/net/context"
)

type CallRequest struct {
	FromPhone string
	ToPhone   string
	Prompt    ai.Prompt
}

type CallResponse struct {
	CallID    string
	SessionID string
}

type VoiceProvider interface {
	CreateCall(ctx context.Context, req CallRequest) (*CallResponse, error)
}
