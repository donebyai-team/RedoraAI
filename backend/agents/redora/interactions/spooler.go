package interactions

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"time"
)

type Spooler struct {
	db                    datastore.Repository
	state                 state.ConversationState
	pollingInterval       time.Duration
	automatedInteractions AutomatedInteractions
	rateLimiter           RateLimiter
	logger                *zap.Logger
}

func NewSpooler(db datastore.Repository, state state.ConversationState, automatedInteractions AutomatedInteractions, rateLimiter RateLimiter, pollingInterval time.Duration, logger *zap.Logger) *Spooler {
	return &Spooler{db: db, state: state, automatedInteractions: automatedInteractions, rateLimiter: rateLimiter, pollingInterval: pollingInterval, logger: logger}
}

func (s *Spooler) Start(ctx context.Context) {
	// 0 so the first time we poll, we do it right away
	interval := 0 * time.Second
	for {
		select {
		case <-time.After(interval):
			if err := s.leadInteractionsToExecute(ctx); err != nil {
				s.logger.Error("failed to poll interactions", zap.Error(err))
			}
		case <-ctx.Done():
		}
		// If we have 0 it means we just started, move to the real interval now
		if interval == 0 {
			interval = s.pollingInterval
		}
	}
}

func (s *Spooler) processInteraction(ctx context.Context, tracker *models.LeadInteraction) error {
	logger := s.logger.With(
		zap.String("interaction_id", tracker.ID))

	uniqueID := fmt.Sprintf("interactions:%s", tracker.ID)

	// Check if a call is already running across organizations
	isRunning, err := s.state.IsRunning(ctx, uniqueID)
	if err != nil {
		return fmt.Errorf("failed to check if interaction is running: %w", err)
	}

	if isRunning {
		logger.Debug("interaction is already in processing state")
		return nil
	}

	// Async call to TrackKeyword
	go func() {
		// Try to acquire the lock and if fails return
		if err := s.state.Acquire(ctx, tracker.Organization.ID, uniqueID); err != nil {
			s.logger.Warn("could not acquire lock for keyword tracker, skipping", zap.Error(err))
			return
		}

		defer func() {
			if err := s.state.Release(ctx, uniqueID); err != nil {
				s.logger.Error("failed to release lock on keyword tracker", zap.Error(err))
			}
		}()
		err = s.automatedInteractions.SendComment(ctx, tracker)
		if err != nil {
			s.logger.Error("failed to send interaction", zap.String("interaction", tracker.ID), zap.Error(err))
		}
	}()

	return nil
}

func (s *Spooler) leadInteractionsToExecute(ctx context.Context) error {
	t0 := time.Now()
	trackers, err := s.db.GetLeadInteractionsToExecute(ctx, []models.LeadInteractionStatus{models.LeadInteractionStatusCREATED})
	if err != nil {
		return fmt.Errorf("processing trackers: %w", err)
	}

	for _, tracker := range trackers {
		s.processInteraction(ctx, tracker)
	}
	s.logger.Info("found interactions to process from db", zap.Int("count", len(trackers)), zap.Duration("elapsed", time.Since(t0)))

	return nil
}
