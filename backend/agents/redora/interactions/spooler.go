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
		zap.String("interaction_type", tracker.Type.String()),
		zap.String("interaction_id", tracker.ID),
	)

	// Only one interaction of the same type should run at a time per project
	uniqueID := fmt.Sprintf("interactions:%s:type:%s", tracker.ProjectID, tracker.Type.String())

	isRunning, err := s.state.IsRunning(ctx, uniqueID)
	if err != nil {
		return fmt.Errorf("failed to check if interaction is running: %w", err)
	}
	if isRunning {
		logger.Debug("interaction is already in processing state")
		return nil
	}

	// Process asynchronously
	go s.processInteractionAsync(ctx, tracker, uniqueID, logger)

	return nil
}

func (s *Spooler) processInteractionAsync(ctx context.Context, tracker *models.LeadInteraction, uniqueID string, logger *zap.Logger) {
	// Acquire lock
	if err := s.state.Acquire(ctx, tracker.Organization.ID, uniqueID); err != nil {
		logger.Warn("could not acquire lock for keyword tracker, skipping", zap.Error(err))
		return
	}
	defer func() {
		if err := s.state.Release(ctx, uniqueID); err != nil {
			logger.Error("failed to release lock on keyword tracker", zap.Error(err))
		}
	}()

	const maxRetries = 2
	const retryDelay = 10 * time.Second
	nonRetryableErrors := []string{
		"Unable to invite the selected invitee", // banned
		"Unable to show the room",               // should not happen, because of redirect
		"suspended",
		"banned",
		"Direct messages may be disabled by the user",
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lastErr = s.sendInteraction(ctx, tracker)

		if lastErr == nil {
			return // Success
		}

		errMsg := lastErr.Error()
		if reason := s.findNonRetryableReason(errMsg, nonRetryableErrors); reason != "" {
			logger.Warn("non-retryable error occurred, skipping retries",
				zap.String("reason", reason),
				zap.Error(lastErr),
			)
			if s.notifier != nil && !strings.Contains(errMsg, "suspended") {
				s.notifier.SendInteractionError(ctx, tracker.ID, lastErr)
			}
			return
		}

		logger.Warn("interaction attempt failed, will retry",
			zap.Int("attempt", i+1),
			zap.Duration("retry_delay", retryDelay),
			zap.Error(lastErr),
		)
		time.Sleep(retryDelay)
	}

	// Final failure after retries
	logger.Error("failed to send interaction after retries", zap.Error(lastErr))
	if s.notifier != nil {
		s.notifier.SendInteractionError(ctx, tracker.ID, fmt.Errorf("failed to send interaction[%s]: %w", tracker.Type.String(), lastErr))
	}
}

// sendInteraction chooses the right method based on tracker type
func (s *Spooler) sendInteraction(ctx context.Context, tracker *models.LeadInteraction) error {
	switch tracker.Type {
	case models.LeadInteractionTypeCOMMENT:
		return s.automatedInteractions.SendComment(ctx, tracker)
	case models.LeadInteractionTypeDM:
		return s.automatedInteractions.SendDM(ctx, tracker)
	default:
		return fmt.Errorf("unsupported interaction type: %v", tracker.Type)
	}
}

// findNonRetryableReason checks if an error message matches a known non-retryable reason
func (s *Spooler) findNonRetryableReason(errMsg string, nonRetryableErrors []string) string {
	for _, reason := range nonRetryableErrors {
		if strings.Contains(errMsg, reason) {
			return reason
		}
	}
	return ""
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

	// Check if already running
	isRunning, err := s.state.IsRunning(ctx, uniqueID)
	if err != nil {
		return fmt.Errorf("failed to check if post is running: %w", err)
	}
	if isRunning {
		logger.Debug("post is already in processing state")
		return nil
	}

	// Process asynchronously
	go s.processPostAsync(ctx, post, uniqueID, logger)

	return nil
}

func (s *Spooler) processPostAsync(ctx context.Context, post *models.Post, uniqueID string, logger *zap.Logger) {
	// Try to acquire lock
	if err := s.state.Acquire(ctx, post.ProjectID, uniqueID); err != nil {
		logger.Warn("could not acquire lock for post, skipping", zap.Error(err))
		return
	}
	defer func() {
		if err := s.state.Release(ctx, uniqueID); err != nil {
			logger.Error("failed to release lock on post", zap.Error(err))
		}
	}()

	// Process post
	if err := s.automatedInteractions.ProcessScheduledPost(ctx, post); err != nil {
		logger.Error("failed to send post", zap.Error(err))
		if s.notifier != nil {
			s.notifier.SendInteractionError(ctx, post.ID, fmt.Errorf("failed to send post: %w", err))
		}
		return
	}

	logger.Info("successfully sent post")
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
