package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"customer_sessions/create.sql",
		"customer_sessions/update.sql",
	})

}

func (r *Database) CreateCustomerSession(ctx context.Context, customer *models.CustomerSession) (*models.CustomerSession, error) {
	stmt := r.mustGetStmt("customer_sessions/create.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"customer_id": customer.CustomerID,
		"due_date":    customer.DueDate,
		"status":      customer.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer session: %w", err)
	}
	customer.ID = id
	return customer, nil
}

func (r *Database) UpdateCustomerSession(ctx context.Context, customer *models.CustomerSession) error {
	stmt := r.mustGetStmt("customer_sessions/update.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"external_id": customer.ExternalID,
		"status":      customer.Status,
		"id":          customer.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer_session %q: %w", customer.ID, err)
	}
	return nil
}
