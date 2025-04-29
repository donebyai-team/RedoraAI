package redora

import (
	"context"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type KeywordTracker interface {
	TrackKeyword(ctx context.Context, tracker *models.AugmentedKeywordTracker) error
	WithLogger(logger *zap.Logger) KeywordTracker
}

type KeywordTrackerFactory struct {
	gptModel          ai.GPTModel
	db                datastore.Repository
	aiClient          *ai.Client
	logger            *zap.Logger
	state             state.ConversationState
	redditOauthClient *reddit.OauthClient
	isDev             bool
}

func NewKeywordTrackerFactory(
	isDev bool,
	gptModel ai.GPTModel,
	redditOauthClient *reddit.OauthClient,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState) *KeywordTrackerFactory {
	return &KeywordTrackerFactory{
		gptModel:          gptModel,
		db:                db,
		aiClient:          aiClient,
		logger:            logger,
		state:             state,
		redditOauthClient: redditOauthClient,
		isDev:             isDev,
	}
}

func (f *KeywordTrackerFactory) GetKeywordTrackerBySource(sourceType models.SourceType) KeywordTracker {
	return newRedditKeywordTracker(
		f.isDev,
		f.gptModel,
		f.redditOauthClient,
		f.db,
		f.aiClient,
		f.logger,
		f.state)
}
