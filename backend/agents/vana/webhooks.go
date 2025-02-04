package vana

import (
	"fmt"
	"github.com/shank318/doota/agents"
	"golang.org/x/net/context"
)

func (s *Spooler) UpdateConversation(ctx context.Context, conversationID string, req []byte) error {
	augConversation, err := s.db.GetConversationByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("error while lookup conversation: %w", err)
	}

	client, err := s.integrationFactory.NewVoiceClient(ctx, augConversation.CustomerCase.OrgID)
	if err != nil {
		return fmt.Errorf("error while lookup voice client: %w", err)
	}

	callResponse, err := client.HandleWebhook(ctx, req)
	if err != nil {
		return fmt.Errorf("error while handling webhook: %w", err)
	}

	conversation := augConversation.Conversation
	conversation.CallStatus = callResponse.CallStatus
	conversation.CallEndedReason = callResponse.CallEndedReason

	// Update conversation
	err = s.db.UpdateConversation(ctx, conversation)
	if err != nil {
		return fmt.Errorf("error while updating conversation: %w", err)
	}

	// MARK a case alive if running
	if agents.IsCallRunning(callResponse.CallStatus) {
		err := s.state.KeepAlive(ctx, augConversation.CustomerCase.OrgID, augConversation.Customer.Phone)
		if err != nil {
			return fmt.Errorf("failed to keep alive for %s, phone %s: %w", augConversation.CustomerCase.ID, augConversation.Customer.Phone, err)
		}
		return nil
	}

	// Case Decision (using llm)
	// Check if I should call again

	return nil
}

func shouldRetry(err error) bool {

}
