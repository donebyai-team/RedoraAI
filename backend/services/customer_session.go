package services

import (
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"golang.org/x/net/context"
	"time"
)

type CreateCustomerSession struct {
	FirstName  string
	LastName   string
	Phone      string
	OrgID      string
	PromptType string
	DueDate    time.Time
}

type CustomerSessionService interface {
	Create(ctx context.Context, session *CreateCustomerSession) error
}

type CustomerSessionServiceImpl struct {
	db datastore.Repository
}

func (c *CustomerSessionServiceImpl) Create(ctx context.Context, session *CreateCustomerSession) error {
	customer, err := c.db.GetCustomerByPhone(ctx, session.Phone, session.OrgID)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return fmt.Errorf("get customer by phone: %w", err)
	}

	if customer == nil {
		customer, err = c.db.CreateCustomer(context.Background(), &models.Customer{
			FirstName: session.FirstName,
			LastName:  session.LastName,
			Phone:     session.Phone,
			OrgID:     session.OrgID,
		})
		if err != nil {
			return err
		}
	}

	// Validate if the given prompt type is synced
	promptType, err := c.db.GetPromptTypeByName(ctx, session.PromptType, session.OrgID)
	if err != nil {
		if errors.Is(err, datastore.NotFound) {
			return fmt.Errorf("prompt type not configured: %s", session.PromptType)
		}
		return fmt.Errorf("get prompt type: %w", err)
	}

	_, err = c.db.CreateCustomerSession(ctx, &models.CustomerSession{
		PromptType: promptType.Name,
		OrgID:      customer.OrgID,
		CustomerID: customer.ID,
		DueDate:    session.DueDate,
		Status:     models.ConversationStatusQUEUED,
	})
	return err
}
