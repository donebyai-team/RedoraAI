package interactions

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
)

type SendCommentInfo struct {
	LeadID        string
	ProjectID     string
	SubredditName string
	Comment       string
	UserName      string
	ThingID       string
}

type AutomatedInteractions interface {
	SendComment(ctx context.Context, leadInteraction *SendCommentInfo) (*models.LeadInteraction, error)
}

type redditInteractions struct {
	redditClient *reddit.Client
	db           datastore.Repository
	logger       *zap.Logger
}

func NewRedditInteractions(redditClient *reddit.Client, db datastore.Repository, logger *zap.Logger) AutomatedInteractions {
	return &redditInteractions{redditClient: redditClient, db: db, logger: logger}
}

func (r redditInteractions) SendComment(ctx context.Context, info *SendCommentInfo) (intr *models.LeadInteraction, err error) {
	r.logger.Info("creating interaction",
		zap.String("type", models.LeadInteractionTypeCOMMENT.String()),
		zap.String("thing_id", info.ThingID),
	)

	info.SubredditName = utils.CleanSubredditName(info.SubredditName)

	intr = &models.LeadInteraction{
		ProjectID: info.ProjectID,
		LeadID:    info.LeadID,
		Type:      models.LeadInteractionTypeCOMMENT,
		From:      info.UserName,
		To:        info.ThingID,
		Metadata:  models.LeadInteractionsMetadata{},
	}

	interaction, err := r.db.CreateLeadInteraction(ctx, intr)
	if err != nil {
		return intr, fmt.Errorf("failed to create lead interaction: %w", err)
	}

	defer func() {
		// Always update interaction at the end
		updateErr := r.db.UpdateLeadInteraction(ctx, interaction)
		if updateErr != nil && err == nil {
			err = fmt.Errorf("failed to update interaction: %w", updateErr)
		}
	}()

	if err = r.redditClient.JoinSubreddit(ctx, info.SubredditName); err != nil {
		interaction.Reason = fmt.Sprintf("failed to join subreddit: %v", err)
		interaction.Status = models.LeadInteractionStatusFAILED
		return intr, err
	}

	var comment *reddit.Comment
	if comment, err = r.redditClient.PostComment(ctx, fmt.Sprintf("t3_%s", info.ThingID), info.Comment); err != nil {
		interaction.Reason = fmt.Sprintf("failed to post comment: %v", err)
		interaction.Status = models.LeadInteractionStatusFAILED
		return intr, err
	}

	if comment == nil {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "comment is nil"
	} else {
		interaction.Status = models.LeadInteractionStatusSENT
		interaction.Reason = ""
		interaction.Metadata.ReferenceID = comment.ID
		interaction.Metadata.Permalink = fmt.Sprintf("r/%s/comments/%s/comment/%s", info.SubredditName, info.ThingID, comment.ID)
	}

	return intr, nil
}
