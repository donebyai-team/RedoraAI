package redora

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"time"
)

type Spooler struct {
	*shutter.Shutter
	interactionSpooler *interactions.Spooler
	dbPollingInterval  time.Duration
	db                 datastore.Repository
	aiClient           *ai.Client
	queue              chan *models.AugmentedKeywordTracker
	queued             *agents.QueuedMap[string, bool]
	state              state.ConversationState
	appIsReady         func() bool
	maxParallelCalls   uint64
	keywordTracker     *KeywordTrackerFactory
	logger             *zap.Logger
}

func New(
	db datastore.Repository,
	interactionSpooler *interactions.Spooler,
	aiClient *ai.Client,
	state state.ConversationState,
	bufferSize int,
	maxParallelCalls uint64,
	dbPollingInterval time.Duration,
	isShuttingDown func() bool,
	keywordTracker *KeywordTrackerFactory,
	logger *zap.Logger,
) *Spooler {
	return &Spooler{
		Shutter:            shutter.New(),
		interactionSpooler: interactionSpooler,
		db:                 db,
		state:              state,
		maxParallelCalls:   maxParallelCalls,
		aiClient:           aiClient,
		dbPollingInterval:  dbPollingInterval,
		appIsReady:         isShuttingDown,
		queue:              make(chan *models.AugmentedKeywordTracker, bufferSize),
		queued:             agents.NewQueuedMap[string, bool](bufferSize),
		logger:             logger,
		keywordTracker:     keywordTracker,
	}
}

func (s *Spooler) Run(ctx context.Context) error {
	go s.runLoop(ctx)
	if !s.keywordTracker.isDev {
		go s.pollKeywordTrackers(ctx)
		go s.interactionSpooler.Start(ctx)
	}
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
			s.queued.Delete(subReddit.GetID())

			// FIXME: We need to deal with errors differently here. We need to separated
			// internal spooler error that are irecoverable from the ones that are
			// coming from the investigator or are recoverable.
			//
			// Indeed, we don't want to stop the spooler if the investigator is broken or
			// something.
			if err := s.processKeywordsTracking(ctx, subReddit); err != nil {
				s.Shutdown(fmt.Errorf("process subreddits: %w", err))
				return
			}
		}
	}
}

func (s *Spooler) processKeywordsTracking(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	logger := s.logger.With(
		zap.String("project_id", tracker.Project.ID),
		zap.String("source", tracker.Source.Name),
		zap.String("tracker_id", tracker.GetID()))

	logger.Debug("processing tracker", zap.Int("queue_size", len(s.queue)))

	// Check if a call is already running across organizations
	isRunning, err := s.state.IsRunning(ctx, tracker.GetID())
	if err != nil {
		return fmt.Errorf("failed to check if tracker is running: %w", err)
	}

	if isRunning {
		logger.Debug("tracker is already in processing state")
		return nil
	}

	// Async call to TrackKeyword
	go func() {
		// Try to acquire the lock and if fails return
		if err := s.state.Acquire(ctx, tracker.Project.OrganizationID, tracker.GetID()); err != nil {
			s.logger.Warn("could not acquire lock for keyword tracker, skipping", zap.Error(err))
			return
		}

		defer func() {
			if err := s.state.Release(ctx, tracker.GetID()); err != nil {
				s.logger.Error("failed to release lock on keyword tracker", zap.Error(err))
			}
		}()

		keywordTracker := s.keywordTracker.GetKeywordTrackerBySource(tracker.Source.SourceType).WithLogger(logger)

		go func() {
			if err := keywordTracker.TrackInSights(ctx, tracker); err != nil {
				logger.Error("failed to track keyword insights", zap.Error(err))
			}
		}()

		if err := keywordTracker.TrackKeyword(ctx, tracker); err != nil {
			logger.Error("failed to track keyword", zap.Error(err))
		}
	}()

	return nil
}

func (s *Spooler) pollKeywordTrackers(ctx context.Context) {
	// 0 so the first time we poll, we do it right away
	interval := 0 * time.Second
	for {
		select {
		case <-time.After(interval):
			if err := s.loadKeywordTrackersToTrack(ctx); err != nil {
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

func (s *Spooler) loadKeywordTrackersToTrack(ctx context.Context) error {
	t0 := time.Now()
	// Query all subreddits per org based on lastTrackedAt, should be > 24hours
	// For each subreddit start the process
	trackers, err := s.db.GetKeywordTrackers(ctx)
	if err != nil {
		return fmt.Errorf("processing trackers: %w", err)
	}

	for _, tracker := range trackers {
		s.pushKeywordToTack(tracker)
	}
	s.logger.Info("found trackers to process from db", zap.Int("count", len(trackers)), zap.Duration("elapsed", time.Since(t0)))

	return nil
}

func (s *Spooler) pushKeywordToTack(tracker *models.AugmentedKeywordTracker) {
	if s.queued.Has(tracker.GetID()) {
		return
	}

	// TODO should we check size vs buffer?
	s.queue <- tracker
	s.queued.Set(tracker.GetID(), true)
}
