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
	"time"
)

func (s *redditKeywordTracker) TrackInSights(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	if !s.shouldTrack(tracker) {
		go s.disableProject(ctx, tracker.Organization)
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

	redditQuery := reddit.PostFilters{
		Keywords: []string{keyword.Keyword},
		SortBy:   utils.Ptr(reddit.SortByTOP),
		TimeRage: utils.Ptr(reddit.TimeRangeWEEK),
		Limit:    100,
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
		lead, err := s.db.GetLeadByPostID(ctx, project.ID, post.ID)
		if err != nil && !errors.Is(err, datastore.NotFound) {
			// Unexpected error, log and skip
			s.logger.Error("error while checking if lead exists by post id", zap.Error(err))
			continue
		}
		if err == nil && lead != nil {
			s.logger.Debug("post already exists", zap.String("post_id", post.ID))
			continue
		}
		// Post doesn't exist, keep it
		newPosts = append(newPosts, post)
	}

	countPostsWithHighRelevancy := 0
	countSkippedPosts := 0
	aiErrorsCount := 0

	s.logger.Info("posts to be evaluated on relevancy via ai", zap.Int("total_posts", len(newPosts)))
	// Filter by AI
	for _, post := range newPosts {
		if aiErrorsCount >= defaultLLMFailedCount {
			return fmt.Errorf("more than %d llm called failed, skipped processing", defaultLLMFailedCount)
		}

		insight := &models.PostInsight{
			ProjectID:      project.ID,
			PostID:         post.ID,
			Source:         models.SourceTypeSUBREDDIT,
			RelevancyScore: 0,
			Metadata: models.PostInsightMetadata{
				Title: post.Title,
			},
		}

		isValid, reason := s.isValidPost(post)
		if isValid {
			// call ai
			postWithAllComments, err := redditClient.GetPostWithAllComments(ctx, post.ID, 10, false)
			if err != nil {
				return err
			}

			postInsight, m, err := s.aiClient.ExtractPostInsight(ctx, s.aiClient.GetAdvanceModel(), ai.PostInsightInput{
				Project: nil,
				Post:    nil,
				Source:  nil,
			}, s.logger)
			if err != nil {
				return err
			}

		} else {
			countSkippedPosts++
			s.logger.Info("ignoring reddit post for ai relevancy check",
				zap.String("post_id", post.ID),
				zap.String("author", post.Author),
				zap.String("reason", reason),
			)

			insight.RelevancyScore = 0
			insight.Metadata.ChainOfThought = reason
		}
	}

}
