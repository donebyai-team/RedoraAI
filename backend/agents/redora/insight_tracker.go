package redora

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/errorx"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
)

func (s *redditKeywordTracker) TrackInSights(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	if !s.shouldTrack(tracker) || tracker.Organization.FeatureFlags.GetSubscriptionPlan() == models.SubscriptionPlanTypeFREE {
		return nil
	}

	redditClient, err := s.redditOauthClient.GetOrCreate(ctx, tracker.Project.OrganizationID, false)
	if err != nil {
		if errors.Is(err, datastore.IntegrationNotFoundOrActive) {
			return nil
		}
		var refreshTokenErr *errorx.RefreshTokenError
		if errors.As(err, &refreshTokenErr) {
			// TODO: Mark project as inactive
			s.logger.Warn("failed to refresh token, skipping tracking", zap.Error(refreshTokenErr))
			return nil
		}
		return err
	}

	project := tracker.Project
	keyword := tracker.Keyword
	source := tracker.Source
	// to make it available downstream
	source.OrgID = project.OrganizationID

	redditQuery := reddit.QueryFilters{
		Keywords:    []string{keyword.Keyword},
		SortBy:      utils.Ptr(reddit.SortByTOP),
		TimeRage:    utils.Ptr(reddit.TimeRangeWEEK),
		Limit:       100,
		MaxComments: 20,
		IncludeMore: false,
	}

	s.logger.Info("started tracking insights",
		zap.String("keyword", keyword.Keyword),
		zap.Any("query", redditQuery))

	posts, err := redditClient.GetPosts(ctx, source.Name, redditQuery)
	if err != nil {
		return err
	}

	newPosts := []*reddit.Post{}
	for _, post := range posts {
		insights, err := s.db.GetInsightsByPostID(ctx, project.ID, post.ID)
		if err != nil {
			// Unexpected error, log and skip
			s.logger.Error("error while checking if insight exists by post id", zap.Error(err))
			continue
		}

		if len(insights) > 0 {
			s.logger.Debug("post insight already exists", zap.String("post_id", post.ID))
			continue
		}
		// insight doesn't exist, keep it
		newPosts = append(newPosts, post)
	}

	countSkippedPosts := 0
	aiErrorsCount := 0
	countPostsWithHighRelevancy := 0

	s.logger.Info("posts insights to be evaluated on relevancy via ai", zap.Int("total_posts", len(newPosts)))
	// Filter by AI
	for _, post := range newPosts {
		if aiErrorsCount >= defaultLLMFailedCount {
			return fmt.Errorf("more than %d llm called failed, skipped processing", defaultLLMFailedCount)
		}

		postInsights := []models.PostInsight{
			{
				ProjectID:      project.ID,
				PostID:         post.ID,
				Source:         models.SourceTypeSUBREDDIT,
				RelevancyScore: 0,
				Metadata: models.PostInsightMetadata{
					Title: post.Title,
				},
			},
		}

		isValid, reason := s.isValidPost(post)
		if isValid {
			redditQueryComments := reddit.QueryFilters{
				SortBy:      utils.Ptr(reddit.SortByTOP),
				TimeRage:    utils.Ptr(reddit.TimeRangeWEEK),
				Limit:       100,
				MaxComments: 20,
				IncludeMore: false,
			}
			postWithAllComments, err := redditClient.GetPostWithAllComments(ctx, post.ID, redditQueryComments)
			if err != nil {
				return fmt.Errorf("failed to get post with all comments: %w", err)
			}

			postInsightAIResponse, _, err := s.aiClient.ExtractPostInsight(ctx, s.aiClient.GetAdvanceModel(), ai.PostInsightInput{
				Project: tracker.Project,
				Post:    postWithAllComments,
			}, s.logger)

			if err != nil {
				s.logger.Error("failed to get insights", zap.Error(err), zap.String("post_id", post.ID))
				aiErrorsCount++
				continue
			}

			if postInsightAIResponse.IsRelevantConfidenceScore < defaultRelevancyScoreInsights {
				s.logger.Info("post is not relevant",
					zap.String("post_id", post.ID),
				)

				postInsights[0].Metadata.ChainOfThought = reason
				postInsights[0].RelevancyScore = postInsightAIResponse.IsRelevantConfidenceScore
			} else {
				countPostsWithHighRelevancy++
				for _, item := range postInsightAIResponse.Insights {
					postInsights = append(postInsights, models.PostInsight{
						RelevancyScore: item.RelevancyScore,
						ProjectID:      project.ID,
						PostID:         post.ID,
						Source:         models.SourceTypeSUBREDDIT,
						Topic:          item.Topic,
						Sentiment:      item.Sentiment,
						Highlights:     item.Highlights,
						Metadata: models.PostInsightMetadata{
							ChainOfThought:      item.ChainOfThought,
							HighlightedComments: item.HighLightedComments,
							Title:               post.Title,
						},
					})
				}
			}
		} else {
			countSkippedPosts++
			s.logger.Info("ignoring reddit post insight for ai relevancy check",
				zap.String("post_id", post.ID),
				zap.String("reason", reason),
			)
			postInsights[0].Metadata.ChainOfThought = reason
		}

		// save the default one
		for _, insight := range postInsights {
			_, err = s.db.CreatePostInsight(ctx, &insight)
			if err != nil {
				s.logger.Error("failed to create post insight", zap.Error(err), zap.String("post_id", post.ID))
				return fmt.Errorf("failed to create post insight: %w", err)
			}
		}
	}

	s.logger.Info("post_insight_summary",
		zap.Int("total_posts_queried", len(posts)),
		zap.Int("total_new_posts", len(newPosts)),
		zap.Int("total_invalid_posts", countSkippedPosts),
		zap.Int("high_relevancy_posts", countPostsWithHighRelevancy))

	return nil
}
