package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"customer/create.sql",
		"customer/query_by_phone.sql",
	})

}

func (r *Database) CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	stmt := r.mustGetStmt("customer/create.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"name":            customer.Name,
		"phone":           customer.Phone,
		"organization_id": customer.OrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	customer.ID = id
	return customer, nil
}

func (r *Database) GetCustomerByPhone(ctx context.Context, phone, organizationID string) (*models.Customer, error) {
	return getOne[models.Customer](ctx, r, "customer/query_by_phone.sql", map[string]any{
		"phone":           phone,
		"organization_id": organizationID,
	})
}
