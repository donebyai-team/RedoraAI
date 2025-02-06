package vana

import (
	"context"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type CaseInvestigator struct {
	gptModel ai.GPTModel
	db       datastore.Repository
	aiClient *ai.Client
	logger   *zap.Logger
	state    state.ConversationState
}

func NewCaseInvestigator(gptModel ai.GPTModel, db datastore.Repository, aiClient *ai.Client, logger *zap.Logger, state state.ConversationState) *CaseInvestigator {
	return &CaseInvestigator{gptModel: gptModel, db: db, aiClient: aiClient, logger: logger, state: state}
}

// TODO: Make it called only once
// Only downside is that it might end up calling AI everytime this is hit in case duplicate webhooks
// and the outcome of AI can change
func (s *CaseInvestigator) UpdateCaseDecision(ctx context.Context, augConversation *models.AugmentedConversation, callResponse *models.CallResponse) error {
	conversation := augConversation.Conversation

	if augConversation.CustomerCase.Status == models.CustomerCaseStatusCLOSED {
		s.logger.Warn("cannot updateCaseDecision because it is already closed, skipped..",
			zap.String("conversationID", conversation.ID))
		return nil
	}

	// if either of these two are true, we have already taken the decision
	if conversation.NextScheduledAt != nil || conversation.Summary != "" {
		s.logger.Warn("next conversation is already scheduled, skipped..",
			zap.String("conversationID", conversation.ID))
		return nil
	}

	if callResponse.CallStatus != models.CallStatusENDED {
		s.logger.Debug("end conversation call status is not ENDED, skipped..",
			zap.String("call id", callResponse.CallID),
			zap.String("phone", augConversation.Customer.Phone),
			zap.String("conversationID", conversation.ID),
		)
		return nil
	}

	// === END CONVERSATION AND UPDATE CASE ===

	// Release call
	err := s.state.Release(ctx, augConversation.Customer.Phone)
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
	// Note: Keep the priority
	pastConversations, err := s.db.GetConversationsByCaseID(ctx, augConversation.CustomerCase.ID)
	callsToday, totalCalls := getCustomerCallStats(pastConversations)
	if totalCalls >= maxTotalAllowedCalls {
		conversation.CustomerCaseStatus = models.CustomerCaseStatusCLOSED
		conversation.CustomerCaseReason = models.CustomerCaseReasonMAXCALLTRIESREACHED
	} else if !hasCustomerPickedCall(callResponse.CallEndedReason) {
		conversation.NextScheduledAt = getNextCallTime(callsToday, conversation.CreatedAt)
	} else if shouldAskAI(augConversation) {
		decision, err := s.aiClient.CustomerCaseDecision(ctx, conversation, s.gptModel, s.logger)
		if err != nil {
			s.logger.Error("failed to ask customer case decision",
				zap.Error(err),
				zap.String("conversationID", conversation.ID),
				zap.String("phone", augConversation.Customer.Phone),
				zap.String("call_id", callResponse.CallID))
		} else {
			conversation.CustomerCaseStatus = caseDecisionToStatus(decision.CaseStatusReason, conversation.CustomerCaseStatus)
			conversation.CustomerCaseReason = decision.CaseStatusReason
			conversation.Summary = decision.Summary
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

	return err
}
