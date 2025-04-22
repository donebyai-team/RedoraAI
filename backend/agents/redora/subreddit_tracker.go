package redora

import (
	"context"
	"errors"
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

type checkIfLeadExists func(ctx context.Context, projectID, ID string) (*models.RedditLead, error)

func (s *SubRedditTracker) TrackSubreddit(ctx context.Context, subReddit *models.AugmentedSubReddit) error {
	redditClient, err := s.redditOauthClient.NewRedditClient(ctx, subReddit.Project.OrganizationID)
	if err != nil {
		return fmt.Errorf("redditOauthClient.NewRedditClient: %w", err)
	}
	for _, keyword := range subReddit.Keywords {
		err = s.searchLeadsFromPosts(ctx, keyword, subReddit.Project, subReddit.SubReddit, redditClient, s.db.GetRedditLeadByPostID)
		if err != nil {
			return fmt.Errorf("searchLeadsFromPosts: %w", err)
		}
	}
	return nil
}

//func (s *SubRedditTracker) searchLeadsFromComments(ctx context.Context, lead *models.RedditLead, redditClient *reddit.Client, leadExists checkIfLeadExists) ([]*models.RedditLead, error) {
//	post, err := redditClient.GetPostWithAllComments(ctx, lead.PostID)
//	if err != nil {
//		return nil, fmt.Errorf("redditClient.GetPostWithAllComments: %w", err)
//	}
//
//	for _, comment := range post.Comments {
//		lead, err := leadExists(ctx, lead.SubRedditID, post.ID)
//		if err == nil && lead != nil {
//			continue
//		}
//		if err != nil && !errors.As(err, &datastore.NotFound) {
//			// Unexpected error, log and skip
//			s.logger.Error("error while checking if lead exists by post id", zap.Error(err))
//			continue
//		}
//
//	}
//
//}

// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
// After filtering, ask AI to filter again
// Save it into the table sub_reddits_leads (models.RedditLead)
func (s *SubRedditTracker) searchLeadsFromPosts(
	ctx context.Context,
	keyword *models.Keyword,
	project *models.Project,
	subReddit *models.SubReddit,
	redditClient *reddit.Client,
	leadExists checkIfLeadExists) error {

	tracker, err := s.db.GetOrCreateSubRedditTracker(ctx, subReddit.ID, keyword.ID)
	if !errors.Is(err, datastore.NotFound) {
		return err
	}

	posts, err := redditClient.GetPosts(ctx, subReddit.SubRedditID, reddit.PostFilters{
		Keywords: []string{keyword.Keyword},
		SortBy:   utils.Ptr(reddit.SortByNEW),
		After:    tracker.NewestTrackedPost,
	})
	if err != nil {
		s.logger.Error("unable to fetch posts while tracking subreddit", zap.String("subreddit", subReddit.SubReddit.Name), zap.Error(err))
		return fmt.Errorf("unable to fetch posts: %w", err)
	}

	s.logger.Info("got posts from reddit sorted by TOP and time range WEEK", zap.Int("total_posts", len(posts)))

	newPosts := []*reddit.Post{}
	for _, post := range posts {
		lead, err := leadExists(ctx, subReddit.SubRedditID, post.ID)
		if err == nil && lead != nil {
			continue
		}
		if err != nil && !errors.Is(err, datastore.NotFound) {
			// Unexpected error, log and skip
			s.logger.Error("error while checking if lead exists by post id", zap.Error(err))
			continue
		}
		// Post doesn't exist, keep it
		newPosts = append(newPosts, post)
	}

	s.logger.Info("posts after check if already exists", zap.Int("total_posts", len(newPosts)))

	// Hard filters
	filteredPosts, err := s.filterAndEnrichPosts(ctx, redditClient, newPosts)
	if err != nil {
		return fmt.Errorf("filterAndEnrichPosts: %w", err)
	}

	s.logger.Info("posts after hard filters",
		zap.Int("filtered_posts", len(filteredPosts)),
		zap.Int("total_posts", len(newPosts)))

	countPostsWithHighRelevancy := 0
	// Filter by AI
	for _, post := range filteredPosts {
		redditLead := &models.RedditLead{
			ProjectID:     project.ID,
			SubRedditID:   subReddit.ID,
			Author:        post.Author,
			PostID:        post.ID,
			Type:          models.LeadTypePOST,
			Title:         utils.Ptr(post.Title),
			Description:   post.Selftext,
			PostCreatedAt: time.Unix(int64(post.CreatedAt), 0),
		}

		relevanceResponse, err := s.aiClient.IsRedditPostRelevant(ctx, project, redditLead, s.gptModel, s.logger)
		if err != nil {
			s.logger.Error("failed to get relevance response", zap.Error(err))
			continue
		}

		redditLead.RelevancyScore = relevanceResponse.IsRelevantConfidenceScore
		if redditLead.RelevancyScore >= 90 {
			countPostsWithHighRelevancy++
		}

		redditLead.LeadMetadata = models.LeadMetadata{
			ChainOfThought:                   relevanceResponse.ChainOfThoughtIsRelevant,
			SuggestedComment:                 relevanceResponse.SuggestedComment,
			SuggestedDM:                      relevanceResponse.SuggestedDM,
			ChainOfThoughtSuggestedComment:   relevanceResponse.ChainOfThoughtSuggestedComment,
			ChainOfThoughtCommentSuggestedDM: relevanceResponse.ChainOfThoughtSuggestedDM,
			PostURL:                          post.URL,
			AuthorInfo:                       post.AuthorInfo,
		}
		err = s.db.CreateRedditLead(ctx, redditLead)
		if err != nil {
			s.logger.Error("unable to create reddit lead", zap.Error(err))
		}
	}

	s.logger.Info("ai suggested posts",
		zap.Int("high_relevancy_posts", countPostsWithHighRelevancy),
		zap.Int("total_filtered_posts", len(filteredPosts)))

	return nil
}

func (s *SubRedditTracker) filterAndEnrichPosts(ctx context.Context, redditClient *reddit.Client, posts []*reddit.Post) ([]*reddit.Post, error) {
	filteredPosts := []*reddit.Post{}
	for _, post := range posts {
		// By Comments
		//if post.NumComments < 2 {
		//	continue
		//}

		if post.Author == "[deleted]" || post.Author == "AutoModerator" {
			s.logger.Info("ignoring reddit post as author is deleted", zap.String("post_id", post.ID), zap.String("author", post.Author))
			continue
		}

		// By Karma
		//user, err := redditClient.GetUser(ctx, post.Author)
		//if err != nil {
		//	s.logger.Error("unable to fetch user, skipped post", zap.String("post_id", post.ID), zap.String("user", post.Author), zap.Error(err))
		//	continue
		//}
		//if user.Karma < 20 {
		//	continue
		//}

		// By Account age,ignore if < 30days
		//if time.Since(time.Unix(int64(user.CreatedAt), 0)) < 30*24*time.Hour {
		//	continue
		//}
		filteredPosts = append(filteredPosts, post)
	}

	return filteredPosts, nil
}
