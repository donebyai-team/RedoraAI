package redora

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"time"
)

type Spooler struct {
	*shutter.Shutter
	dbPollingInterval  time.Duration
	gptModel           ai.GPTModel
	db                 datastore.Repository
	aiClient           *ai.Client
	queue              chan *models.AugmentedSubReddit
	queued             *agents.QueuedMap[string, bool]
	integrationFactory *integrations.Factory
	state              state.ConversationState
	appIsReady         func() bool
	maxParallelCalls   uint64
	subRedditTracker   *SubRedditTracker
	logger             *zap.Logger
}

func New(
	db datastore.Repository,
	aiClient *ai.Client,
	gptModel ai.GPTModel,
	state state.ConversationState,
	integrationFactory *integrations.Factory,
	bufferSize int,
	maxParallelCalls uint64,
	dbPollingInterval time.Duration,
	isShuttingDown func() bool,
	subRedditTracker *SubRedditTracker,
	logger *zap.Logger,
) *Spooler {
	return &Spooler{
		Shutter:            shutter.New(),
		db:                 db,
		gptModel:           gptModel,
		state:              state,
		maxParallelCalls:   maxParallelCalls,
		aiClient:           aiClient,
		integrationFactory: integrationFactory,
		dbPollingInterval:  dbPollingInterval,
		appIsReady:         isShuttingDown,
		queue:              make(chan *models.AugmentedSubReddit, bufferSize),
		queued:             agents.NewQueuedMap[string, bool](bufferSize),
		logger:             logger,
		subRedditTracker:   subRedditTracker,
	}
}

func (s *Spooler) Run(ctx context.Context) error {
	go s.runLoop(ctx)
	go s.pollSubReddits(ctx)
	return nil
}

func (s *Spooler) runLoop(ctx context.Context) {
	s.logger.Info("running spooler loop")
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("ctx done, run loop ending")
			return
		case <-s.Terminating():
			s.logger.Info("spooler terminating, run loop ending")
			return
		case subReddit := <-s.queue:
			// Remove the case from the queued map, we are processing it
			s.queued.Delete(subReddit.SubReddit.ID)

			// FIXME: We need to deal with errors differently here. We need to separated
			// internal spooler error that are irecoverable from the ones that are
			// coming from the investigator or are recoverable.
			//
			// Indeed, we don't want to stop the spooler if the investigator is broken or
			// something.
			if err := s.processKeywordsTracking(ctx, subReddit); err != nil {
				s.Shutdown(fmt.Errorf("process investigation: %w", err))
				return
			}
		}
	}
}

func (s *Spooler) processKeywordsTracking(ctx context.Context, subReddit *models.AugmentedSubReddit) error {
	logger := s.logger.With(
		zap.String("subreddit_id", subReddit.SubReddit.SubRedditID),
		zap.String("organization_id", subReddit.SubReddit.OrganizationID),
		zap.String("creator", "redora"),
	)
	logger.Debug("processing customer cases", zap.Int("queue_size", len(s.queue)))

	// Check if a call is already running across organizations
	isRunning, err := s.state.IsRunning(ctx, subReddit.SubReddit.ID)
	if err != nil {
		return fmt.Errorf("failed to check if subreddit is running: %w", err)
	}

	if isRunning {
		logger.Debug("subreddit is already in processting state")
		return nil
	}

	// Call Reddit APIs to fetch the list of posts

	return nil
}

func (s *Spooler) pollSubReddits(ctx context.Context) {
	// 0 so the first time we poll, we do it right away
	interval := 0 * time.Second
	for {
		select {
		case <-time.After(interval):
			if err := s.loadSubRedditsToTrack(ctx); err != nil {
				s.Shutdown(fmt.Errorf("fail to load customer sessions from db: %w", err))
			}
		case <-ctx.Done():
		}

		// If we have 0 it means we just started, move to the real interval now
		if interval == 0 {
			interval = s.dbPollingInterval
		}
	}
}

func (s *Spooler) loadSubRedditsToTrack(ctx context.Context) error {
	t0 := time.Now()
	// Query all subreddits per org based on lastTrackedAt, should be > 24hours
	// For each subreddit start the process
	subReddits, err := s.db.GetSubReddits(ctx)
	if err != nil {
		return fmt.Errorf("processing subreddits: %w", err)
	}

	for _, reddit := range subReddits {
		s.pushSubRedditToTack(reddit)
	}
	s.logger.Info("found subreddit to process from db", zap.Int("count", len(subReddits)), zap.Duration("elapsed", time.Since(t0)))

	return nil
}

func (s *Spooler) pushSubRedditToTack(subReddit *models.AugmentedSubReddit) {
	if s.queued.Has(subReddit.SubReddit.ID) {
		return
	}

	// TODO should we check size vs buffer?
	s.queue <- subReddit
	s.queued.Set(subReddit.SubReddit.ID, true)
}
