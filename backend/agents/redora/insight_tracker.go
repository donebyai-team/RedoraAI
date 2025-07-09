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
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"time"
)

const (
	keyMaxInsightsPerWeek         = "insights_per_week"
	keyMaxPostsForInsightsPerWeek = "posts_tracked_for_insights_per_week"

	// insights
	defaultRelevancyScoreInsights     = 90
	maxInsightsPerWeek                = 5
	maxPostsToTrackForInsightsPerWeek = 100
)

func (s *redditKeywordTracker) TrackInSights(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	if !s.shouldTrack(tracker) || tracker.Organization.FeatureFlags.GetSubscriptionPlan() == models.SubscriptionPlanTypeFREE {
		return nil
	}

	// We will try to keep searching until we reach the max relevant posts per day >= defaultRelevancyScore
	if ok, err := s.isMaxPostInsightsReached(ctx, tracker.Project.ID); err != nil || ok {
		return err
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
				SourceID:       source.ID,
				KeywordID:      keyword.ID,
				RelevancyScore: 0,
				Metadata: models.PostInsightMetadata{
					Title:         post.Title,
					PostCreatedAt: time.Unix(int64(post.CreatedAt), 0),
					NoOfComments:  post.NumComments,
					Upvotes:       post.Ups,
					Score:         post.Score,
				},
			},
		}

		isValid, reason := s.isValidPostForInsight(post)
		if isValid {
			redditQueryComments := reddit.QueryFilters{
				SortBy:      utils.Ptr(reddit.SortByCONFIDENCE),
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

			if len(postInsightAIResponse.Insights) == 0 {
				s.logger.Info("post has not enough insights, skipping",
					zap.String("post_id", post.ID),
				)

				postInsights[0].Metadata.ChainOfThought = reason
			} else {
				var generatedPostInsights []models.PostInsight
				for _, item := range postInsightAIResponse.Insights {
					if item.RelevancyScore >= defaultRelevancyScoreInsights {
						countPostsWithHighRelevancy++
					}
					generatedPostInsights = append(generatedPostInsights, models.PostInsight{
						RelevancyScore: item.RelevancyScore,
						ProjectID:      project.ID,
						PostID:         post.ID,
						SourceID:       source.ID,
						KeywordID:      keyword.ID,
						Topic:          item.Topic,
						Sentiment:      item.Sentiment,
						Highlights:     item.Highlights,
						Metadata: models.PostInsightMetadata{
							ChainOfThought:      item.ChainOfThought,
							HighlightedComments: item.HighLightedComments,
							Title:               post.Title,
							PostCreatedAt:       time.Unix(int64(post.CreatedAt), 0),
							NoOfComments:        post.NumComments,
							Upvotes:             post.Ups,
							Score:               post.Score,
						},
					})
				}

				if len(generatedPostInsights) > 0 {
					postInsights = generatedPostInsights
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

		shouldFinish := false
		// we will let extra insights to get saved
		// if max is achieved, will break after saving all
		for _, insight := range postInsights {
			_, err = s.db.CreatePostInsight(ctx, &insight)
			if err != nil {
				s.logger.Error("failed to create post insight", zap.Error(err), zap.String("post_id", post.ID))
				return fmt.Errorf("failed to create post insight: %w", err)
			}

			// check if we reached max
			if insight.RelevancyScore >= defaultRelevancyScoreInsights {
				isAllowed, err := s.state.CheckIfUnderLimitAndIncrement(ctx, weeklyCounterKey(tracker.Organization.ID), keyMaxInsightsPerWeek, maxInsightsPerWeek, 8*24*time.Hour)
				if err != nil {
					return fmt.Errorf("failed to check if under limit and increment: %w", err)
				}
				if !isAllowed {
					shouldFinish = true
				}
			}
		}

		if shouldFinish {
			s.logger.Info("max insight limit reached, skipping comment", zap.String("post_id", post.ID))
			break
		}

		// We will try to keep searching until we reach the max relevant posts per day >= defaultRelevancyScore
		ok, err := s.isMaxPostInsightsReached(ctx, tracker.Project.ID)
		if err != nil || ok {
			if err != nil {
				return err
			}
			break
		}
	}

	s.logger.Info("post_insight_summary",
		zap.Int("total_posts_queried", len(posts)),
		zap.Int("total_new_posts", len(newPosts)),
		zap.Int("total_invalid_posts", countSkippedPosts),
		zap.Int("high_relevancy_insights", countPostsWithHighRelevancy))

	return nil
}

func (s *redditKeywordTracker) isValidPostForInsight(post *reddit.Post) (bool, string) {
	if post.NumComments < 10 {
		return false, "less than 10 comments"
	}

	if post.Ups < 10 {
		return false, "less than 10 upvotes"
	}

	if post.Score == 0 {
		return false, "score is 0"
	}

	return s.isValidPost(post)
}

func (s *redditKeywordTracker) isMaxPostInsightsReached(ctx context.Context, projectID string) (bool, error) {
	insights, err := s.db.GetInsights(ctx, projectID, datastore.LeadsFilter{
		RelevancyScore: defaultRelevancyScoreInsights,
		Limit:          100,
		Offset:         0,
		DateRange:      pbportal.DateRangeFilter_DATE_RANGE_7_DAYS,
	})
	if err != nil {
		return false, err
	}

	if len(insights) >= maxInsightsPerWeek {
		s.logger.Info("reached max insights per week",
			zap.Int("count", len(insights)))
		return true, nil
	}

	insights2, err := s.db.GetInsights(ctx, projectID, datastore.LeadsFilter{
		RelevancyScore: 1,
		Limit:          100,
		Offset:         0,
		DateRange:      pbportal.DateRangeFilter_DATE_RANGE_7_DAYS,
	})
	if err != nil {
		return false, err
	}

	if len(insights2) >= maxPostsToTrackForInsightsPerWeek {
		s.logger.Info("reached max posts to extact insights per week",
			zap.Int("count", len(insights)))
		return true, nil
	}

	return false, nil
}

func weeklyCounterKey(orgID string) string {
	year, week := time.Now().UTC().ISOWeek()
	return fmt.Sprintf("org:%s:counters:%d-W%02d", orgID, year, week)
}
