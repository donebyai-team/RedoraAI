package agents

import "context"

type WebhookHandler interface {
	UpdateCallStatus(ctx context.Context, conversationID string, req []byte) error
	EndConversation(ctx context.Context, conversationID string, req []byte) error
}
