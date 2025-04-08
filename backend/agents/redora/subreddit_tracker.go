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

type SubRedditTracker struct {
	gptModel          ai.GPTModel
	db                datastore.Repository
	aiClient          *ai.Client
	logger            *zap.Logger
	state             state.ConversationState
	redditOauthClient *reddit.OauthClient
}

func NewSubRedditTracker(
	gptModel ai.GPTModel,
	redditOauthClient *reddit.OauthClient,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState) *SubRedditTracker {
	return &SubRedditTracker{
		gptModel:          gptModel,
		db:                db,
		aiClient:          aiClient,
		logger:            logger,
		state:             state,
		redditOauthClient: redditOauthClient,
	}
}

func (s *SubRedditTracker) TrackSubreddit(ctx context.Context, subReddit *models.AugmentedSubReddit) error {
	//integration, err := s.db.GetIntegrationByOrgAndType(ctx, subReddit.SubReddit.OrganizationID, models.IntegrationTypeREDDIT)
	//if err != nil {
	//	return err
	//}
	//redditClient, err := s.redditOauthClient.NewRedditClient(ctx, s.logger, s.db, integration)
	//if err != nil {
	//	return fmt.Errorf("redditOauthClient.NewRedditClient: %w", err)
	//}

	// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
	// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
	// After filtering, ask AI to filter again
	// Save it into the table sub_reddits_leads (models.RedditLead)

	panic("implement me")
}
