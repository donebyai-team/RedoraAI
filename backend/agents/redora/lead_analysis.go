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
	// Relevant leads check
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

	for _, interaction := range interactionsScheduled {
		if interaction.Type == models.LeadInteractionTypeCOMMENT {
			analysis.CommentScheduled++
		} else if interaction.Type == models.LeadInteractionTypeDM {
			analysis.DmScheduled++
		}
	}

	// total comment sent
	interactionsSent, err := a.automatedInteractions.GetInteractions(ctx, projectID, models.LeadInteractionStatusSENT, dateRange)
	if err != nil {
		return nil, err
	}
	for _, interaction := range interactionsSent {
		if interaction.Type == models.LeadInteractionTypeCOMMENT {
			analysis.CommentSent++
		} else if interaction.Type == models.LeadInteractionTypeDM {
			analysis.DmSent++
		}
	}
	return &analysis, nil
}
