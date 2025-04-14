package redora

import (
	"context"
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"time"
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
	redditClient, err := s.redditOauthClient.NewRedditClient(ctx, subReddit.Project.OrganizationID)
	if err != nil {
		return fmt.Errorf("redditOauthClient.NewRedditClient: %w", err)
	}

	// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
	// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
	// After filtering, ask AI to filter again
	// Save it into the table sub_reddits_leads (models.RedditLead)
	keywords := []string{}
	for _, keyword := range subReddit.Keywords {
		keywords = append(keywords, keyword.Keyword)
	}
	posts, err := redditClient.GetPosts(ctx, subReddit.SubReddit.SubRedditID, reddit.PostFilters{
		Keywords: keywords,
		SortBy:   utils.Ptr(reddit.SortByTOP),
		TimeRage: utils.Ptr(reddit.TimeRangeWEEK),
	})
	if err != nil {
		s.logger.Error("unable to fetch posts while tracking subreddit", zap.String("subreddit", subReddit.SubReddit.URL), zap.Error(err))
		return fmt.Errorf("unable to fetch posts: %w", err)
	}
	// Hard filters
	filteredPosts, err := s.filterAndEnrichPosts(ctx, redditClient, posts)
	if err != nil {
		return fmt.Errorf("filterAndEnrichPosts: %w", err)
	}

	// Filter by AI
	for _, post := range filteredPosts {
		redditLead := &models.RedditLead{
			ProjectID:     subReddit.Project.ID,
			SubRedditID:   subReddit.SubReddit.ID,
			Author:        post.Author,
			PostID:        post.ID,
			Type:          models.RedditLeadTypePOST,
			Title:         utils.Ptr(post.Title),
			Description:   post.Selftext,
			PostCreatedAt: time.Unix(post.CreatedAt, 0),
		}

		relevanceResponse, err := s.aiClient.IsRedditPostRelevant(ctx, subReddit.Project, redditLead, s.gptModel, s.logger)
		if err != nil {
			s.logger.Error("failed to get relevance response", zap.Error(err))
			continue
		}

		redditLead.RelevancyScore = relevanceResponse.IsRelevantConfidenceScore
		redditLead.RedditLeadMetadata = models.RedditLeadMetadata{
			ChainOfThought:                   relevanceResponse.ChainOfThoughtIsRelevant,
			SuggestedComment:                 relevanceResponse.SuggestedComment,
			SuggestedDM:                      relevanceResponse.SuggestedDM,
			ChainOfThoughtSuggestedComment:   relevanceResponse.ChainOfThoughtSuggestedComment,
			ChainOfThoughtCommentSuggestedDM: relevanceResponse.ChainOfThoughtSuggestedDM,
			NoOfComments:                     post.NumComments,
			NoOfLikes:                        post.Ups,
		}

		// Save
	}

	return nil
}

func (s *SubRedditTracker) filterAndEnrichPosts(ctx context.Context, redditClient *reddit.Client, posts []*reddit.Post) ([]*reddit.Post, error) {
	filteredPosts := []*reddit.Post{}
	for _, post := range posts {
		// By Comments
		if post.NumComments < 2 {
			continue
		}

		// By Karma
		user, err := redditClient.GetUser(ctx, post.Author)
		if err != nil {
			s.logger.Error("unable to fetch user, skipped post", zap.String("post_id", post.ID), zap.String("user", post.Author), zap.Error(err))
			continue
		}
		if user.Karma < 20 {
			continue
		}

		// By Account age,ignore if < 30days
		if time.Since(time.Unix(user.CreatedAt, 0)) < 30*24*time.Hour {
			continue
		}
		post.AuthorInfo = user
		filteredPosts = append(filteredPosts, post)
	}

	return filteredPosts, nil
}
