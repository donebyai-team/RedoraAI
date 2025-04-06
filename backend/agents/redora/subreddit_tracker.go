package redora

import (
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"go.uber.org/zap"
)

type SubRedditTracker struct {
	gptModel ai.GPTModel
	db       datastore.Repository
	aiClient *ai.Client
	logger   *zap.Logger
	state    state.ConversationState
}

func NewSubRedditTracker(gptModel ai.GPTModel, db datastore.Repository, aiClient *ai.Client, logger *zap.Logger, state state.ConversationState) *SubRedditTracker {
	return &SubRedditTracker{gptModel: gptModel, db: db, aiClient: aiClient, logger: logger, state: state}
}
