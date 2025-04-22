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
	RedditRepository
	ProjectRepository
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

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *models.Project) (*models.Project, error)
	GetProjects(ctx context.Context, orgID string) ([]*models.Project, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)
}

type RedditRepository interface {
	CreateKeyword(ctx context.Context, keyword *models.Keyword) (*models.Keyword, error)
	GetSubReddits(ctx context.Context) ([]*models.AugmentedSubReddit, error)

	// Subreddit
	AddSubReddit(ctx context.Context, subreddit *models.SubReddit) (*models.SubReddit, error)
	GetSubRedditByName(ctx context.Context, url, orgID string) (*models.SubReddit, error)
	DeleteSubRedditByID(ctx context.Context, id string) error
	GetSubRedditByID(ctx context.Context, ID string) (*models.SubReddit, error)
	GetSubRedditsByProject(ctx context.Context, projectID string) ([]*models.SubReddit, error)
	UpdateSubRedditLastTrackedAt(ctx context.Context, id string) error

	// Subreddit leads
	GetRedditLeadsByStatus(ctx context.Context, projectID string, status models.LeadStatus) ([]*models.RedditLead, error)
	GetRedditLeadsByRelevancy(ctx context.Context, projectID string, relevancy float32, subReddits []string) ([]*models.RedditLead, error)
	GetRedditLeadByPostID(ctx context.Context, projectID, postID string) (*models.RedditLead, error)
	GetRedditLeadByCommentID(ctx context.Context, projectID, commentID string) (*models.RedditLead, error)
	CreateRedditLead(ctx context.Context, reddit *models.RedditLead) error
	UpdateRedditLeadStatus(ctx context.Context, lead *models.RedditLead) error
	GetRedditLeadByID(ctx context.Context, projectID, id string) (*models.RedditLead, error)

	// Subreddit tracker
	GetOrCreateSubRedditTracker(ctx context.Context, subredditID, keywordID string) (*models.SubRedditTracker, error)
	UpdateSubRedditTracker(ctx context.Context, subreddit *models.SubRedditTracker) (*models.SubRedditTracker, error)
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
	LastCallStatus []models.CallStatus
	CaseStatus     []models.CustomerCaseStatus
	CurrentTime    time.Time
}
