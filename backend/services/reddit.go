package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"strings"
	"time"
)

type RedditService interface {
	CreateSubReddit(ctx context.Context, subReddit *models.Source) error
	GetSubReddits(ctx context.Context, orgID string) ([]*models.Source, error)
	RemoveSubReddit(ctx context.Context, id string) error
}

type redditService struct {
	aiClient     *ai.Client
	db           datastore.Repository
	redditClient *reddit.Client
	logger       *zap.Logger
}

func NewRedditService(logger *zap.Logger, db datastore.Repository, redditClient *reddit.Client, aiClient *ai.Client) *redditService {
	return &redditService{logger: logger, db: db, redditClient: redditClient, aiClient: aiClient}
}

func (r redditService) CreateSubReddit(ctx context.Context, source *models.Source) error {
	// Check if the subreddit already exists in the DB
	existingSubreddit, err := r.db.GetSourceByName(ctx, source.Name, source.ProjectID)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return fmt.Errorf("get existing subreddit: %w", err)
	}

	if existingSubreddit != nil {
		return fmt.Errorf("subreddit already exists: %s", source.Name)
	}

	// Fetch the subreddit details by URL from Reddit API
	subRedditDetails, err := r.redditClient.GetSubRedditByName(ctx, source.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch subreddit details from Reddit: %w", err)
	}

	// Fill in the fields in models.Source using fetched details
	if subRedditDetails.ID != "" {
		source.ExternalID = utils.Ptr(subRedditDetails.ID)
	}
	source.Name = subRedditDetails.DisplayName
	source.Description = subRedditDetails.Description
	source.SourceType = models.SourceTypeSUBREDDIT
	metadata := models.SubRedditMetadata{
		Title:     utils.Ptr(subRedditDetails.Title),
		CreatedAt: time.Unix(int64(subRedditDetails.CreatedAt), 0),
	}

	for _, rule := range subRedditDetails.Rules {
		if strings.TrimSpace(rule.Description) != "" {
			metadata.Rules = append(metadata.Rules, rule.Description)
		}
	}

	source.Metadata = metadata

	// Get evaluation
	evaluation, usage, err := r.aiClient.GetSourceCommunityRulesEvaluation(ctx, "", source, r.logger)
	if err != nil {
		return fmt.Errorf("failed to evaluate community guidelines: %w", err)
	}

	source.Metadata.RulesEvaluation = evaluation
	source.Metadata.RulesEvaluation.ModelUsed = usage.Model

	// Insert the subreddit into the DB
	createdSource, err := r.db.AddSource(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to add subreddit to the database: %w", err)
	}

	// Create trackers for each keyword
	keywords, err := r.db.GetKeywords(ctx, createdSource.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to fetch sources by project: %w", err)
	}

	for _, keyword := range keywords {
		_, err := r.db.CreateKeywordTracker(ctx, &models.KeywordTracker{
			SourceID:  createdSource.ID,
			KeywordID: keyword.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to add keyword tracker for source [%s]: %w", createdSource.Name, err)
		}
	}

	return nil
}

func (r redditService) GetSubReddits(ctx context.Context, projectID string) ([]*models.Source, error) {
	sources, err := r.db.GetSourcesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		if source.SourceType == models.SourceTypeSUBREDDIT {
			source.Name = fmt.Sprintf("r/%s", source.Name)
		}
	}
	return sources, nil
}

func (r redditService) RemoveSubReddit(ctx context.Context, id string) error {
	// Step 1: Try to get the subreddit to check if it exists
	subreddit, err := r.db.GetSourceByID(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to fetch subreddit with ID %s: %w", id, err)
	}
	if subreddit == nil {
		return fmt.Errorf("no subreddit found under the project")
	}

	// Step 2: Delete the subreddit
	err = r.db.DeleteSourceByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete subreddit with ID %s: %w", id, err)
	}

	return nil
}
