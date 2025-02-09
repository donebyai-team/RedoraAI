package vana

import (
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type vanaWebhookHandler struct {
	db                 datastore.Repository
	integrationFactory *integrations.Factory
	state              state.ConversationState
	caseInvestigator   *CaseInvestigator
	logger             *zap.Logger
}

func NewVanaWebhookHandler(
	db datastore.Repository,
	state state.ConversationState,
	caseInvestigator *CaseInvestigator,
	integrationFactory *integrations.Factory,
	logger *zap.Logger,
) *vanaWebhookHandler {
	return &vanaWebhookHandler{
		db:                 db,
		state:              state,
		caseInvestigator:   caseInvestigator,
		integrationFactory: integrationFactory,
		logger:             logger,
	}
}

func (s *vanaWebhookHandler) UpdateCallStatus(ctx context.Context, conversationID string, req []byte) error {
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

	return s.caseInvestigator.UpdateCustomerCase(ctx, augConversation, callResponse)
}

func (s *vanaWebhookHandler) EndConversation(ctx context.Context, conversationID string, req []byte) error {
	augConversation, err := s.db.GetConversationByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("error while lookup conversation: %w", err)
	}

	if augConversation.CustomerCase.Status == models.CustomerCaseStatusCLOSED {
		s.logger.Warn("cannot end conversation because it is already closed",
			zap.String("conversationID", conversationID))
		return nil
	}

	client, err := s.integrationFactory.NewVoiceClient(ctx, augConversation.CustomerCase.OrgID)
	if err != nil {
		return fmt.Errorf("error while lookup voice client: %w", err)
	}

	callResponse, err := client.HandleWebhook(ctx, req)
	if err != nil {
		return fmt.Errorf("error while handling webhook: %w", err)
	}

	if callResponse.CallStatus != models.CallStatusENDED {
		s.logger.Debug("end conversation call status is not ENDED",
			zap.String("call id", callResponse.CallID),
			zap.String("phone", augConversation.Customer.Phone),
			zap.String("conversationID", conversationID),
		)
		return nil
	}

	return s.caseInvestigator.UpdateCustomerCase(ctx, augConversation, callResponse)
}
