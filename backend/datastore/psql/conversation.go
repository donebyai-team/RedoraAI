package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"conversation/create_conversation.sql",
		"conversation/update_conversation.sql",
		"conversation/query_conversation_by_id.sql",
		"conversation/query_conversation_by_caseid.sql",
		"customer_case/update_customer_case.sql",
	})
}
func (r *Database) GetConversationByID(ctx context.Context, id string) (*models.AugmentedConversation, error) {
	conversation, err := getOne[models.Conversation](ctx, r, "conversation/query_conversation_by_id.sql", map[string]any{
		"id": id,
	})

	if err != nil {
		return nil, fmt.Errorf("get conversation by id: %w", err)
	}

	customerCase, err := r.GetCustomerCaseByID(ctx, conversation.CustomerCaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations for customer case %q: %w", customerCase.ID, err)
	}

	customer, err := r.GetCustomerByID(ctx, customerCase.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by id %q: %w", customerCase.CustomerID, err)
	}

	return &models.AugmentedConversation{
		CustomerCase: customerCase,
		Customer:     customer,
		Conversation: conversation,
	}, nil
}

func (r *Database) CreateConversation(ctx context.Context, obj *models.Conversation) (*models.Conversation, error) {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	stmt := r.mustGetTxStmt(ctx, "conversation/create_conversation.sql", tx)
	var id string

	err = stmt.GetContext(ctx, &id, map[string]interface{}{
		"customer_case_id": obj.CustomerCaseID,
		"from_phone":       obj.FromPhone,
		"provider":         obj.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer comversation: %w", err)
	}
	obj.ID = id

	//stmt = r.mustGetTxStmt(ctx, "customer_case/update.sql", tx)
	//_, err = stmt.ExecContext(ctx, map[string]interface{}{
	//	"id":                obj.CustomerCaseID,
	//	"last_call_status":  models.ConversationStatusCREATED,
	//	"status":            models.CustomerCaseStatusCREATED,
	//	"next_scheduled_at": obj.NextScheduledAt,
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("failed to update customer_case %q: %w", obj.CustomerCaseID, err)
	//}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction, update conversation: %w", err)
	}

	return obj, nil
}

func (r *Database) UpdateConversationAndCase(ctx context.Context, augConv *models.AugmentedConversation) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	obj := augConv.Conversation

	stmt := r.mustGetTxStmt(ctx, "conversation/update_conversation.sql", tx)
	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"summary":            obj.Summary,
		"recording_url":      obj.RecordingURL,
		"call_duration":      obj.CallDuration,
		"end_of_call_reason": obj.CallEndedReason,
		"call_status":        obj.CallStatus,
		"next_scheduled_at":  obj.NextScheduledAt,
		"id":                 obj.ID,
		"external_id":        obj.ExternalID,
		"call_messages":      obj.CallMessages,
		"ai_decision":        obj.AIDecision,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer conversation %q: %w", obj.ID, err)
	}

	caseObj := augConv.CustomerCase

	stmt = r.mustGetTxStmt(ctx, "customer_case/update_customer_case.sql", tx)
	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"id":                obj.CustomerCaseID,
		"last_call_status":  obj.CallStatus,
		"next_scheduled_at": obj.NextScheduledAt,

		"summary":     caseObj.Summary,
		"case_reason": caseObj.CaseReason,
		"status":      caseObj.Status,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer_case %q: %w", obj.CustomerCaseID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction, update conversation: %w", err)
	}

	return nil
}

func (r *Database) GetConversationsByCaseID(ctx context.Context, customerCaseID string) ([]*models.Conversation, error) {
	return getMany[models.Conversation](ctx, r, "conversation/query_conversation_by_caseid.sql", map[string]any{
		"customer_case_id": customerCaseID,
	})
}
