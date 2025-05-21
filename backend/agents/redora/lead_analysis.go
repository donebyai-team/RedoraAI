package redora

import (
	"context"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"go.uber.org/zap"
)

type LeadAnalysis struct {
	db                    datastore.Repository
	automatedInteractions interactions.AutomatedInteractions
	logger                *zap.Logger
}

func NewLeadAnalysis(db datastore.Repository, logger *zap.Logger) *LeadAnalysis {
	return &LeadAnalysis{db: db,
		automatedInteractions: interactions.NewSimpleRedditInteractions(db, logger),
		logger:                logger,
	}
}

func (a LeadAnalysis) GenerateLeadAnalysis(ctx context.Context, projectID string, dateRange pbportal.DateRangeFilter) (*pbportal.LeadAnalysis, error) {
	analysis := pbportal.LeadAnalysis{}
	// Relevant leads
	leadsData, err := a.db.CountLeadByCreatedAt(ctx, projectID, dailyPostsRelevancyScore, dateRange)
	if err != nil {
		return nil, err
	}
	analysis.RelevantPostsFound = leadsData.Count

	// Total leads tracked
	leadsData, err = a.db.CountLeadByCreatedAt(ctx, projectID, 0, dateRange)
	if err != nil {
		return nil, err
	}
	analysis.PostsTracked = leadsData.Count

	// total comment scheduled
	interactionsScheduled, err := a.automatedInteractions.GetInteractions(ctx, projectID, models.LeadInteractionStatusCREATED, dateRange)
	if err != nil {
		return nil, err
	}
	analysis.CommentScheduled = uint32(len(interactionsScheduled))

	// total comment sent
	interactionsSent, err := a.automatedInteractions.GetInteractions(ctx, projectID, models.LeadInteractionStatusSENT, dateRange)
	if err != nil {
		return nil, err
	}
	analysis.CommentSent = uint32(len(interactionsSent))
	return &analysis, nil
}
