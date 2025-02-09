package datastore

import (
	"context"
	"errors"
	"github.com/shank318/doota/models"
	"time"
)

var NotFound = errors.New("not found")
var ErrMessageSourceAlreadyExists = errors.New("message source already exists")
var ErrMessageAlreadyExists = errors.New("message already exists")

type Repository interface {
	OrganizationRepository
	IntegrationRepository
	UserRepository
	PromptTypeRepository
	ConversationRepository
	CustomerRepository
	CustomerCaseRepository
}

type OrganizationRepository interface {
	CreateOrganization(context.Context, *models.Organization) (*models.Organization, error)
	UpdateOrganization(context.Context, *models.Organization) error
	GetOrganizations(context.Context) ([]*models.Organization, error)
	GetOrganizationById(context.Context, string) (*models.Organization, error)
	GetOrganizationByName(context.Context, string) (*models.Organization, error)
}

type IntegrationRepository interface {
	UpsertIntegration(ctx context.Context, integration *models.Integration) (*models.Integration, error)
	GetIntegrationByOrgAndType(ctx context.Context, organizationId string, integrationType models.IntegrationType) (*models.Integration, error)
	GetIntegrationsByOrgID(ctx context.Context, orgID string) ([]*models.Integration, error)
	GetIntegrationById(ctx context.Context, id string) (*models.Integration, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	GetUserById(ctx context.Context, userID string) (*models.User, error)
	GetUserByAuth0Id(ctx context.Context, auth0ID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type CustomerRepository interface {
	CreateCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	GetCustomerByPhone(ctx context.Context, phone, organizationID string) (*models.Customer, error)
}

type CustomerCaseRepository interface {
	CreateCustomerCase(ctx context.Context, customer *models.CustomerCase) (*models.CustomerCase, error)
	UpdateCustomerCase(ctx context.Context, customer *models.CustomerCase) error
	GetCustomerCases(ctx context.Context, filter CustomerCaseFilter) ([]*models.AugmentedCustomerCase, error)
	GetCustomerCaseByID(ctx context.Context, id string) (*models.CustomerCase, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, obj *models.Conversation) (*models.Conversation, error)
	UpdateConversationAndCase(ctx context.Context, obj *models.AugmentedConversation) error
	GetConversationsByCaseID(ctx context.Context, customerCaseID string) ([]*models.Conversation, error)
	GetConversationByID(ctx context.Context, id string) (*models.AugmentedConversation, error)
}

type PromptTypeRepository interface {
	CreatePromptType(ctx context.Context, PromptType *models.PromptType) (*models.PromptType, error)
	UpdatePromptType(ctx context.Context, PromptType *models.PromptType) error
	GetPromptTypeByName(ctx context.Context, name string) (*models.PromptType, error)
}

type CustomerCaseFilter struct {
	LastCallStatus  []models.CallStatus
	CaseStatus      []models.CustomerCaseStatus
	NextScheduledAt time.Time
}
