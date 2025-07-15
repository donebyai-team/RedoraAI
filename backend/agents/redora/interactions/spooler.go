package interactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/alerts"
	"go.uber.org/zap"
)

type Spooler struct {
	db                    datastore.Repository
	state                 state.ConversationState
	pollingInterval       time.Duration
	automatedInteractions AutomatedInteractions
	rateLimiter           RateLimiter
	notifier              alerts.AlertNotifier
	logger                *zap.Logger
}

func NewSpooler(db datastore.Repository, notifier alerts.AlertNotifier, state state.ConversationState, automatedInteractions AutomatedInteractions, rateLimiter RateLimiter, pollingInterval time.Duration, logger *zap.Logger) *Spooler {
	return &Spooler{db: db, state: state, notifier: notifier, automatedInteractions: automatedInteractions, rateLimiter: rateLimiter, pollingInterval: pollingInterval, logger: logger}
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
			if err := s.postsToExecute(ctx); err != nil {
				s.logger.Error("failed to poll scheduled posts", zap.Error(err))
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

	// we want only one type of interaction to be executed at once
	uniqueID := fmt.Sprintf("interactions:%s:type:%s", tracker.ProjectID, tracker.Type.String())

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
		if err := s.state.Acquire(ctx, tracker.Organization.ID, uniqueID); err != nil {
			s.logger.Warn("could not acquire lock for keyword tracker, skipping", zap.Error(err))
			return
		}
		defer func() {
			if err := s.state.Release(ctx, uniqueID); err != nil {
				s.logger.Error("failed to release lock on keyword tracker", zap.Error(err))
			}
		}()

		const maxRetries = 2
		const retryDelay = 10 * time.Second
		nonRetryableErrors := []string{
			"Unable to invite the selected invitee", // banned
			"Unable to show the room",               // should not happen, because of redirect
			"suspended",
			"banned",
		}

		var lastErr error
		for i := 0; i < maxRetries; i++ {
			if tracker.Type == models.LeadInteractionTypeCOMMENT {
				lastErr = s.automatedInteractions.SendComment(ctx, tracker)
			} else if tracker.Type == models.LeadInteractionTypeDM {
				lastErr = s.automatedInteractions.SendDM(ctx, tracker)
			} else {
				lastErr = fmt.Errorf("unsupported interaction type: %v", tracker.Type)
			}

			if lastErr == nil {
				return // success
			}

			errMsg := lastErr.Error()
			for _, nonRetry := range nonRetryableErrors {
				if strings.Contains(errMsg, nonRetry) {
					s.logger.Warn("non-retryable error occurred, skipping retries",
						zap.String("interaction_type", tracker.Type.String()),
						zap.String("reason", nonRetry),
						zap.Error(lastErr),
					)

					if s.notifier != nil && !strings.Contains(errMsg, "suspended") {
						s.notifier.SendInteractionError(ctx, tracker.ID, lastErr)
					}
					return
				}
			}

			s.logger.Warn("interaction attempt failed, will retry in 10 seconds:",
				zap.Int("attempt", i+1),
				zap.String("interaction_type", tracker.Type.String()),
				zap.Error(lastErr),
			)

			time.Sleep(retryDelay)
		}

		// Final failure after retries
		s.logger.Error("failed to send interaction after retries",
			zap.String("interaction_type", tracker.Type.String()),
			zap.Error(lastErr),
		)

		if s.notifier != nil {
			s.notifier.SendInteractionError(ctx, tracker.ID, lastErr)
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

func (s *Spooler) processPost(ctx context.Context, post *models.Post) error {
	logger := s.logger.With(
		zap.String("post_id", post.ID),
		zap.String("project_id", post.ProjectID),
	)

	uniqueID := fmt.Sprintf("post:%s", post.ID)

	isRunning, err := s.state.IsRunning(ctx, uniqueID)
	if err != nil {
		return fmt.Errorf("failed to check if post is running: %w", err)
	}

	if isRunning {
		logger.Debug("post is already in processing state")
		return nil
	}

	go func() {
		if err := s.state.Acquire(ctx, post.ProjectID, uniqueID); err != nil {
			s.logger.Warn("could not acquire lock for post, skipping", zap.Error(err))
			return
		}
		defer func() {
			if err := s.state.Release(ctx, uniqueID); err != nil {
				s.logger.Error("failed to release lock on post", zap.Error(err))
			}
		}()

		err := s.automatedInteractions.ProcessScheduledPost(ctx, post)
		if err != nil {
			s.logger.Error("failed to send post",
				zap.String("post_id", post.ID),
				zap.Error(err),
			)

			// TODO: Notify about the error
			// if s.notifier != nil {
			//     s.notifier.SendPostError(ctx, post.ID, err)
			// }

			return
		}

		logger.Info("successfully sent post")
	}()

	return nil
}

func (s *Spooler) postsToExecute(ctx context.Context) error {
	t0 := time.Now()

	posts, err := s.db.GetPostsToExecute(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch scheduled posts: %w", err)
	}

	for _, post := range posts {
		s.processPost(ctx, post)
	}

	s.logger.Info("found posts to process from db", zap.Int("count", len(posts)), zap.Duration("elapsed", time.Since(t0)))
	return nil
}
