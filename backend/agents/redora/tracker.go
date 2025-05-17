package redora

import (
	"context"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/alerts"
	"go.uber.org/zap"
)

type KeywordTracker interface {
	TrackKeyword(ctx context.Context, tracker *models.AugmentedKeywordTracker) error
	WithLogger(logger *zap.Logger) KeywordTracker
}

type KeywordTrackerFactory struct {
	db                datastore.Repository
	aiClient          *ai.Client
	logger            *zap.Logger
	state             state.ConversationState
	redditOauthClient *reddit.OauthClient
	isDev             bool
	slackNotifier     alerts.AlertNotifier
	emailNotifier     alerts.AlertNotifier
}

func NewKeywordTrackerFactory(
	isDev bool,
	redditOauthClient *reddit.OauthClient,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState,
	slackNotifier alerts.AlertNotifier,
	emailNotifier alerts.AlertNotifier) *KeywordTrackerFactory {
	return &KeywordTrackerFactory{
		db:                db,
		aiClient:          aiClient,
		logger:            logger,
		state:             state,
		redditOauthClient: redditOauthClient,
		isDev:             isDev,
		slackNotifier:     slackNotifier,
		emailNotifier:     emailNotifier,
	}
}

func (f *KeywordTrackerFactory) GetKeywordTrackerBySource(sourceType models.SourceType) KeywordTracker {
	return newRedditKeywordTracker(
		f.isDev,
		f.redditOauthClient,
		interactions.NewRedditInteractions(f.redditOauthClient, f.db, f.logger),
		f.db,
		f.aiClient,
		f.logger,
		f.state, f.slackNotifier, f.emailNotifier)
}
