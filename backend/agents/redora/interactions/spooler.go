package interactions

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"time"
)

type Spooler struct {
	db                    datastore.Repository
	pollingInterval       time.Duration
	automatedInteractions AutomatedInteractions
	rateLimiter           RateLimiter
	logger                *zap.Logger
}

func NewSpooler(db datastore.Repository, automatedInteractions AutomatedInteractions, rateLimiter RateLimiter, pollingInterval time.Duration, logger *zap.Logger) *Spooler {
	return &Spooler{db: db, automatedInteractions: automatedInteractions, rateLimiter: rateLimiter, pollingInterval: pollingInterval, logger: logger}
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

func (s *Spooler) leadInteractionsToExecute(ctx context.Context) error {
	t0 := time.Now()
	trackers, err := s.db.GetLeadInteractionsToExecute(ctx, []models.LeadInteractionStatus{models.LeadInteractionStatusCREATED})
	if err != nil {
		return fmt.Errorf("processing trackers: %w", err)
	}

	for _, tracker := range trackers {
		err = s.automatedInteractions.SendComment(ctx, tracker)
		if err != nil {
			s.logger.Error("failed to send interaction", zap.String("interaction", tracker.ID), zap.Error(err))
		}
	}
	s.logger.Info("found interactions to process from db", zap.Int("count", len(trackers)), zap.Duration("elapsed", time.Since(t0)))

	return nil
}
