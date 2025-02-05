package vana

import (
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *Spooler) UpdateCallStatus(ctx context.Context, conversationID string, req []byte) error {
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

	return s.updateConversationFromCall(ctx, augConversation, callResponse)
}

// TODO: Make it idempotent, CallStatus should not change once ended
func (s *Spooler) EndConversation(ctx context.Context, conversationID string, req []byte) error {
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

	return s.updateConversationFromCall(ctx, augConversation, callResponse)
}

func (s *Spooler) updateCaseDecision(ctx context.Context, augConversation *models.AugmentedConversation, callResponse *models.CallResponse) {
	conversation := augConversation.Conversation

	if augConversation.CustomerCase.Status == models.CustomerCaseStatusCLOSED {
		s.logger.Warn("cannot updateCaseDecision because it is already closed",
			zap.String("conversationID", conversation.ID))
		return
	}

	if callResponse.CallStatus != models.CallStatusENDED {
		s.logger.Debug("end conversation call status is not ENDED",
			zap.String("call id", callResponse.CallID),
			zap.String("phone", augConversation.Customer.Phone),
			zap.String("conversationID", conversation.ID),
		)
		return
	}

	// === END CONVERSATION AND UPDATE CASE ===

	// Release call
	err := s.state.Release(ctx, augConversation.CustomerCase.OrgID, augConversation.Customer.Phone)
	if err != nil {
		if err != nil {
			s.logger.Error("failed to release call",
				zap.Error(err),
				zap.String("conversationID", conversation.ID),
				zap.String("phone", augConversation.Customer.Phone),
				zap.String("call id", callResponse.CallID))
		}
	}

	// Mark case failed if reached max tries
	// Case Decision (using llm)
	// Call again if customer not picked the call
	pastConversations, err := s.db.GetConversationsByCaseID(ctx, augConversation.CustomerCase.ID)
	callsToday, totalCalls := getCustomerCallStats(pastConversations)
	if totalCalls >= maxTotalAllowedCalls {
		conversation.CustomerCaseStatus = models.CustomerCaseStatusCLOSED
		conversation.CustomerCaseReason = models.CustomerCaseReasonMAXCALLTRIESREACHED
	} else if !hasCustomerPickedCall(callResponse.CallEndedReason) {
		conversation.NextScheduledAt = getNextCallTime(callsToday, conversation.CreatedAt)
	} else if shouldAskAI(augConversation) {
		decision, err := s.aiClient.CustomerCaseDecision(ctx, conversation, nil, s.logger)
		if err != nil {
			s.logger.Error("failed to ask customer case decision",
				zap.Error(err),
				zap.String("conversationID", conversation.ID),
				zap.String("phone", augConversation.Customer.Phone),
				zap.String("call id", callResponse.CallID))
		} else {
			conversation.CustomerCaseStatus = caseDecisionToStatus(decision.CaseStatusReason, conversation.CustomerCaseStatus)
			conversation.CustomerCaseReason = decision.CaseStatusReason
			if decision.NextCallScheduledAtTime != nil {
				conversation.NextScheduledAt = decision.NextCallScheduledAtTime
			}
		}
	}

	err = s.db.UpdateConversation(ctx, conversation)
	if err != nil {
		s.logger.Error("failed to update conversation",
			zap.Error(err),
			zap.String("conversationID", conversation.ID),
			zap.String("phone", augConversation.Customer.Phone),
			zap.String("call id", callResponse.CallID))
	}
}

func (s *Spooler) updateConversationFromCall(ctx context.Context, augConversation *models.AugmentedConversation, callResponse *models.CallResponse) error {
	conversation := augConversation.Conversation
	conversation.CallStatus = callResponse.CallStatus
	conversation.ExternalID = callResponse.CallID
	conversation.CallEndedReason = callResponse.CallEndedReason

	// Update conversation
	err := s.db.UpdateConversation(ctx, conversation)
	if err != nil {
		s.logger.Error("failed to update conversation",
			zap.Error(err),
			zap.String("conversationID", conversation.ID),
			zap.String("phone", augConversation.Customer.Phone),
			zap.String("call id", callResponse.CallID))
		return err
	}

	// Mark a call running
	if callResponse.CallID != "" && agents.IsCallRunning(callResponse.CallStatus) {
		err := s.state.KeepAlive(ctx, augConversation.CustomerCase.OrgID, augConversation.Customer.Phone)
		if err != nil {
			return fmt.Errorf("failed to keep alive for %s, phone %s: %w", augConversation.CustomerCase.ID, augConversation.Customer.Phone, err)
		}
	}

	go s.updateCaseDecision(ctx, augConversation, callResponse)

	return nil
}

func caseDecisionToStatus(decision models.CustomerCaseReason, existingStatus models.CustomerCaseStatus) models.CustomerCaseStatus {
	if utils.Some(caseClosedReasons, func(reason models.CustomerCaseReason) bool {
		return reason == decision
	}) {
		return models.CustomerCaseStatusCLOSED
	}

	return existingStatus
}

var caseClosedReasons = []models.CustomerCaseReason{
	models.CustomerCaseReasonPAID,
	models.CustomerCaseReasonPARTIALLYPAID,
	models.CustomerCaseReasonMAXCALLTRIESREACHED,
}

var callNotPickedReasons = []models.CallEndedReason{
	models.CallEndedReasonASSISTANTERROR,
	models.CallEndedReasonCUSTOMERBUSY,
}

func hasCustomerPickedCall(callEndedReason models.CallEndedReason) bool {
	return !utils.Some(callNotPickedReasons, func(matchType models.CallEndedReason) bool {
		return matchType == callEndedReason
	})
}

func shouldAskAI(augConversation *models.AugmentedConversation) bool {
	if len(augConversation.Conversation.CallMessages) == 0 {
		return false
	}
	hasUserResponded := false
	for _, msg := range augConversation.Conversation.CallMessages {
		if msg.UserMessage != nil {
			hasUserResponded = true
		}
	}

	return hasUserResponded
}
