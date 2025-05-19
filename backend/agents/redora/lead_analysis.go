package redora

import (
	"context"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"go.uber.org/zap"
	"time"
)

type LeadAnalysis struct {
	db                    datastore.Repository
	automatedInteractions interactions.AutomatedInteractions
	logger                *zap.Logger
}

func NewLeadAnalysis(db datastore.Repository, logger *zap.Logger) *LeadAnalysis {
	return &LeadAnalysis{db: db,
		automatedInteractions: interactions.NewRedditInteractions(db, nil, logger),
		logger:                logger,
	}
}

func (a LeadAnalysis) GenerateLeadAnalysis(ctx context.Context, projectID string) (*pbportal.LeadAnalysis, error) {
	analysis := pbportal.LeadAnalysis{}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	// Relevant leads
	leadsData, err := a.db.CountLeadByCreatedAt(ctx, projectID, dailyPostsRelevancyScore, today, tomorrow)
	if err != nil {
		return nil, err
	}
	analysis.RelevantPostsFound = leadsData.Count

	// Total leads tracked
	leadsData, err = a.db.CountLeadByCreatedAt(ctx, projectID, 0, today, tomorrow)
	if err != nil {
		return nil, err
	}
	analysis.PostsTracked = leadsData.Count

	// total comment scheduled
	interactionsScheduled, err := a.automatedInteractions.GetInteractionsPerDay(ctx, projectID, models.LeadInteractionStatusCREATED)
	if err != nil {
		return nil, err
	}
	analysis.CommentScheduled = uint32(len(interactionsScheduled))

	// total comment sent
	interactionsSent, err := a.automatedInteractions.GetInteractionsPerDay(ctx, projectID, models.LeadInteractionStatusSENT)
	if err != nil {
		return nil, err
	}
	analysis.CommentSent = uint32(len(interactionsSent))
	return &analysis, nil
}
