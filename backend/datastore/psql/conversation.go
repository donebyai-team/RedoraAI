package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"conversation/create.sql",
		"conversation/update.sql",
		"customer_sessions/update.sql",
	})

}
func (r *Database) CreateConversation(ctx context.Context, obj *models.Conversation) (*models.Conversation, error) {
	stmt := r.mustGetStmt("conversation/create.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"customer_session_id": obj.CustomerSessionID,
		"from_phone":          obj.FromPhone,
		"status":              obj.Status,
		"provider":            obj.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer comversation: %w", err)
	}
	obj.ID = id
	return obj, nil
}

func (r *Database) UpdateConversation(ctx context.Context, externalSessionID string, obj *models.Conversation) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	stmt := r.mustGetTxStmt(ctx, "conversation/update.sql", tx)
	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"summary":       obj.Summary,
		"recording_url": obj.RecordingURL,
		"call_duration": obj.CallDuration,
		"status":        obj.Status,
		"id":            obj.ID,
		"external_id":   obj.ExternalID,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer conversation %q: %w", obj.ID, err)
	}

	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"external_id": externalSessionID,
		"status":      obj.Status,
		"id":          obj.CustomerSessionID,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer_session %q: %w", obj.CustomerSessionID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction, update conversation: %w", err)
	}

	return nil
}
