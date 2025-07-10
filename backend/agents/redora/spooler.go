package redora

import (
	"context"
	"fmt"
	"time"

	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
)

const defaultMaxParallelCalls = 10 // Adjust this value based on observed performance

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
	if maxParallelCalls == 0 {
		maxParallelCalls = defaultMaxParallelCalls
	}

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
	s.logger.Info("starting spooler", zap.Uint64("max_parallel_calls", s.maxParallelCalls))
	if s.keywordTracker.isDev {
		s.maxParallelCalls = 1
	}

	// Start fixed number of workers to consume from queue
	for i := uint64(0); i < s.maxParallelCalls; i++ {
		go s.worker(ctx, int(i))
	}

	if !s.keywordTracker.isDev {
		go s.pollKeywordTrackers(ctx)
		go s.interactionSpooler.Start(ctx)
	}

	return nil
}

// worker is a bounded parallel consumer from the shared queue
func (s *Spooler) worker(ctx context.Context, id int) {
	s.logger.Info("worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("worker exiting: context canceled", zap.Int("worker_id", id))
			return
		case <-s.Terminating():
			s.logger.Info("worker exiting: shutting down", zap.Int("worker_id", id))
			return
		case tracker := <-s.queue:
			s.logger.Info("worker picked up tracker", zap.Int("queue_size", len(s.queue)), zap.Int("worker_id", id))
			if err := s.processKeywordsTracking(ctx, tracker); err != nil {
				s.logger.Error("failed to process tracker", zap.Error(err))
			}
			s.queued.Delete(tracker.GetID())
		}
	}
}

const maxParallelTrackerPerProject = 5

func (s *Spooler) processKeywordsTracking(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	logger := s.logger.With(
		zap.String("project_id", tracker.Project.ID),
		zap.String("source", tracker.Source.Name),
		zap.String("tracker_id", tracker.GetID()))

	logger.Debug("processing tracker", zap.Int("queue_size", len(s.queue)))

	isRunning, err := s.state.IsRunningTracker(ctx, tracker.Project.ID, tracker.GetID())
	if err != nil {
		return fmt.Errorf("failed to check if tracker is running: %w", err)
	}

	if isRunning {
		logger.Debug("tracker is already in processing state")
		return nil
	}

	// Try to acquire the lock and if fails return
	acquired, err := s.state.AcquireTracker(ctx, tracker.Project.ID, tracker.GetID(), maxParallelTrackerPerProject)
	if err != nil {
		logger.Warn("could not acquire lock for keyword tracker, skipping", zap.Error(err))
		return fmt.Errorf("failed to acquire lock for keyword tracker: %w", err)
	}

	if !acquired {
		logger.Info("could not acquire lock for keyword tracker, max tracker per project reached, skipping")
		return nil
	}

	defer func() {
		if err := s.state.ReleaseTracker(ctx, tracker.Project.ID, tracker.GetID()); err != nil {
			logger.Error("failed to release lock on keyword tracker", zap.Error(err))
		}
	}()

	keywordTracker := s.keywordTracker.GetKeywordTrackerBySource(tracker.Source.SourceType).WithLogger(logger)

	// Run TrackInSights in parallel
	//go func() {
	//	if err := keywordTracker.TrackInSights(ctx, tracker); err != nil {
	//		logger.Error("failed to track keyword insights", zap.Error(err))
	//	}
	//}()

	if err := keywordTracker.TrackKeyword(ctx, tracker); err != nil {
		logger.Error("failed to track keyword", zap.Error(err))
	}

	return nil
}
func (s *Spooler) pollKeywordTrackers(ctx context.Context) {
	interval := 0 * time.Second // Run immediately on startup

	for {
		select {
		case <-time.After(interval):
			if err := s.loadKeywordTrackersToTrack(ctx); err != nil {
				s.Shutdown(fmt.Errorf("fail to load customer sessions from db: %w", err))
			}
		case <-ctx.Done():
			return
		}

		// After first run, switch to regular interval
		if interval == 0 {
			interval = s.dbPollingInterval
		}
	}
}

func (s *Spooler) loadKeywordTrackersToTrack(ctx context.Context) error {
	t0 := time.Now()

	trackers, err := s.db.GetKeywordTrackers(ctx)
	if err != nil {
		return fmt.Errorf("processing trackers: %w", err)
	}

	for _, tracker := range trackers {
		s.pushKeywordToTrack(tracker)
	}

	s.logger.Info("found trackers to process from db",
		zap.Int("count", len(trackers)),
		zap.Int("queue_size", len(s.queue)),
		zap.Duration("elapsed", time.Since(t0)))

	return nil
}

func (s *Spooler) pushKeywordToTrack(tracker *models.AugmentedKeywordTracker) {
	if s.queued.Has(tracker.GetID()) {
		return
	}

	s.queued.Set(tracker.GetID(), true)

	select {
	case s.queue <- tracker:
		// Enqueued successfully
	case <-time.After(5 * time.Second):
		s.logger.Warn("enqueue timeout - queue may be full", zap.String("tracker_id", tracker.GetID()), zap.Int("queue_size", len(s.queue)))
	}
}
