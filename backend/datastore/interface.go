package datastore

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"time"
)

var NotFound = errors.New("not found")
var ErrMessageSourceAlreadyExists = errors.New("message source already exists")
var IntegrationNotFoundOrActive = errors.New("integration not found or active")

func IsUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}
	return false
}

type Repository interface {
	OrganizationRepository
	IntegrationRepository
	UserRepository
	PromptTypeRepository
	ConversationRepository
	CustomerRepository
	CustomerCaseRepository
	SourceRepository
	LeadRepository
	KeywordRepository
	ProjectRepository
	LeadInteractionRepository
	SubscriptionRepository
	PostInsightRepository
	PostRepository
}

type OrganizationRepository interface {
	CreateOrganization(context.Context, *models.Organization) (*models.Organization, error)
	UpdateOrganization(context.Context, *models.Organization) error
	GetOrganizations(context.Context) ([]*models.Organization, error)
	GetOrganizationById(context.Context, string) (*models.Organization, error)
	GetOrganizationByName(context.Context, string) (*models.Organization, error)
	UpdateOrganizationFeatureFlags(ctx context.Context, orgID string, updates map[string]any) error
}

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub *models.Subscription, tx *sqlx.Tx) (*models.Subscription, error)
	GetSubscriptionByOrgID(ctx context.Context, orgID string) (*models.Subscription, error)
	GetSubscriptionByIDAndOrg(ctx context.Context, id, orgID string) (*models.Subscription, error)
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
	GetUsersByOrgID(ctx context.Context, orgID string) ([]*models.User, error)
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
	GetProjectByName(ctx context.Context, name, orgID string) (*models.Project, error)
	UpdateProject(ctx context.Context, project *models.Project) (*models.Project, error)
	UpdateProjectIsActive(ctx context.Context, orgID string, isActive bool) error
}

type SourceRepository interface {
	AddSource(ctx context.Context, subreddit *models.Source) (*models.Source, error)
	UpdateSource(ctx context.Context, subreddit *models.Source) error
	GetSourceByName(ctx context.Context, url, orgID string) (*models.Source, error)
	DeleteSourceByID(ctx context.Context, id string) error
	GetSourceByID(ctx context.Context, ID string) (*models.Source, error)
	GetSourcesByProject(ctx context.Context, projectID string) ([]*models.Source, error)
}

type LeadInteractionRepository interface {
	CreateLeadInteraction(ctx context.Context, reddit *models.LeadInteraction) (*models.LeadInteraction, error)
	UpdateLeadInteraction(ctx context.Context, reddit *models.LeadInteraction) error
	GetLeadInteractionByLeadID(ctx context.Context, leadID string) ([]*models.LeadInteraction, error)
	GetLeadInteractionByID(ctx context.Context, id string) (*models.LeadInteraction, error)
	GetLeadInteractions(ctx context.Context, projectID string, status models.LeadInteractionStatus, dateRange pbportal.DateRangeFilter) ([]*models.LeadInteraction, error)
	GetLeadInteractionsToExecute(ctx context.Context, statuses []models.LeadInteractionStatus) ([]*models.LeadInteraction, error)
	SetLeadInteractionStatusProcessing(ctx context.Context, id string) error
	IsInteractionExists(ctx context.Context, interaction *models.LeadInteraction) (bool, error)
	GetAugmentedLeadInteractions(ctx context.Context, projectID string, dateRange pbportal.DateRangeFilter) ([]*models.AugmentedLeadInteraction, error)
}

type LeadsFilter struct {
	RelevancyScore float32
	Sources        []string
	Statuses       []string
	Limit          int
	Offset         int
	DateRange      pbportal.DateRangeFilter
}

type LeadRepository interface {
	GetLeadsByStatus(ctx context.Context, projectID string, filter LeadsFilter) ([]*models.AugmentedLead, error)
	GetLeadsByRelevancy(ctx context.Context, projectID string, filter LeadsFilter) ([]*models.AugmentedLead, error)
	GetLeadByPostID(ctx context.Context, projectID, postID string) (*models.Lead, error)
	GetLeadByCommentID(ctx context.Context, projectID, commentID string) (*models.Lead, error)
	CreateLead(ctx context.Context, reddit *models.Lead) (*models.Lead, error)
	UpdateLeadStatus(ctx context.Context, lead *models.Lead) error
	GetLeadByID(ctx context.Context, projectID, id string) (*models.Lead, error)
	CountLeadByCreatedAt(ctx context.Context, projectID string, relevancyScore int, dateRange pbportal.DateRangeFilter) (*models.LeadsData, error)
}

type KeywordRepository interface {
	GetKeywords(ctx context.Context, projectID string) ([]*models.Keyword, error)
	CreateKeywords(ctx context.Context, projectID string, keywords []string) error
	RemoveKeyword(ctx context.Context, projectID, keywordID string) error
	GetKeywordTrackers(ctx context.Context) ([]*models.AugmentedKeywordTracker, error)
	UpdatKeywordTrackerLastTrackedAt(ctx context.Context, id string) error
	CreateKeywordTracker(ctx context.Context, tracker *models.KeywordTracker) (*models.KeywordTracker, error)
	GetKeywordTrackerByProjectID(ctx context.Context, projectID string) ([]*models.KeywordTracker, error)
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

type PostInsightRepository interface {
	CreatePostInsight(ctx context.Context, insight *models.PostInsight) (*models.PostInsight, error)
	GetInsightsByPostID(ctx context.Context, projectID, postID string) ([]*models.PostInsight, error)
	GetInsights(ctx context.Context, projectID string, filter LeadsFilter) ([]*models.AugmentedPostInsight, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	GetPostByID(ctx context.Context, ID string) (*models.Post, error)
	UpdatePost(ctx context.Context, post *models.Post) error
	GetPostsByProjectID(ctx context.Context, projectID string) ([]*models.Post, error)
}

type CustomerCaseFilter struct {
	LastCallStatus []models.CallStatus
	CaseStatus     []models.CustomerCaseStatus
	CurrentTime    time.Time
}
