package interactions

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
)

func (r redditInteractions) ProcessScheduledPost(ctx context.Context, post *models.Post) (err error) {
	logger := r.logger.With(zap.String("post_id", post.ID))

	logger.Info("sending post to reddit")

	project, err := r.db.GetProject(ctx, post.ProjectID)
	if err != nil {
		return err
	}

	source, err := r.db.GetSourceByID(ctx, post.SourceID)
	if err != nil {
		return err
	}

	defer func() {
		if updateErr := r.db.UpdatePost(ctx, post); updateErr != nil {
			logger.Error("failed to update post in defer", zap.Error(updateErr))
			if err == nil {
				err = fmt.Errorf("post update failed: %w", updateErr)
			}
		}
	}()

	if !project.IsActive {
		post.Status = models.PostStatusFAILED
		post.Reason = "Project is not active"
		return nil
	}

	err = r.redditOauthClient.WithRotatingAPIClient(ctx, project.OrganizationID, func(client *reddit.Client) error {
		config := client.GetConfig()

		subredditName := utils.CleanSubredditName(source.Name)
		err = client.JoinSubreddit(ctx, subredditName)
		if err != nil && !errors.Is(err, reddit.ErrForbidden) {
			post.Status = models.PostStatusFAILED
			post.Reason = fmt.Sprintf("Reason: failed to join subreddit %v", err)
			return err
		}

		post.Metadata.Author = config.Name

		redditPost, err := client.CreatePost(ctx, subredditName, post)
		if err != nil {
			post.Status = models.PostStatusFAILED
			post.Reason = fmt.Sprintf("Failed to Post: %v", err)
			return err
		}

		post.PostID = &redditPost.ID
		post.Status = models.PostStatusSENT
		post.Reason = ""

		logger.Info("successfully posted to Reddit", zap.String("reddit_post_id", redditPost.ID))
		return nil
	}, reddit.MostQualifiedAccountStrategy(logger), logger)

	if err != nil {
		post.Status = models.PostStatusFAILED
		// if the reason is not set then set it to the error message
		if post.Reason == "" {
			post.Reason = err.Error()
		}
	}

	return err
}
