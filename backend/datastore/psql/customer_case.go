package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"customer_case/create.sql",
		"customer_case/update.sql",
		"customer_case/query_by_filter.sql",
	})

}

func (r *Database) CreateCustomerCase(ctx context.Context, customer *models.CustomerCase) (*models.CustomerCase, error) {
	stmt := r.mustGetStmt("customer_case/create.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"customer_id": customer.CustomerID,
		"due_date":    customer.DueDate,
		"status":      customer.Status,
		"prompt_type": customer.PromptType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer session: %w", err)
	}
	customer.ID = id
	return customer, nil
}

func (r *Database) UpdateCustomerCase(ctx context.Context, customer *models.CustomerCase) error {
	stmt := r.mustGetStmt("customer_case/update.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"status":            customer.Status,
		"last_call_status":  customer.LastCallStatus,
		"next_scheduled_at": customer.NextScheduledAt,
		"summary":           customer.Summary,
		"id":                customer.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer_session %q: %w", customer.ID, err)
	}
	return nil
}

func (r *Database) GetCustomerCases(ctx context.Context, filter datastore.CustomerCaseFilter) ([]*models.AugmentedCustomerCase, error) {
	customerCases, err := getMany[models.CustomerCase](ctx, r, "customer_case/query_by_filter.sql", map[string]any{
		"status":            filter.CaseStatus,
		"last_call_status":  filter.LastCallStatus,
		"next_scheduled_at": filter.NextScheduledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get customer cases: %w", err)
	}
	var results []*models.AugmentedCustomerCase
	for _, customerCase := range customerCases {
		conversations, err := r.GetConversationsByCaseID(ctx, customerCase.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversations for customer case %q: %w", customerCase.ID, err)
		}

		customer, err := r.GetCustomerByID(ctx, customerCase.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer by id %q: %w", customerCase.CustomerID, err)
		}

		results = append(results, &models.AugmentedCustomerCase{
			CustomerCase:  customerCase,
			Customer:      customer,
			Conversations: conversations,
		})
	}

	return results, nil
}
