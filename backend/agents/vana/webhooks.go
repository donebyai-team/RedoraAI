package vana

import (
	"fmt"
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

	// MARK a case alive
	err = s.markCallAlive(ctx, callResponse, augConversation.CustomerCase.OrgID, augConversation.Customer.Phone)
	if err != nil {
		return fmt.Errorf("error while mark call alive: %w", err)
	}

	// Case Decision (using llm)
	// Check if I should call again

	return nil
}

func shouldRetry(err error) bool {

}
