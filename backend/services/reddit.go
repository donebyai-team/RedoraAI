package services

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents/state"
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
	cache        state.ConversationState
	logger       *zap.Logger
}

func NewRedditService(logger *zap.Logger, db datastore.Repository, redditClient *reddit.Client, aiClient *ai.Client, cache state.ConversationState) *redditService {
	return &redditService{logger: logger, db: db, redditClient: redditClient, aiClient: aiClient, cache: cache}
}

func (r redditService) cacheSubReddit(ctx context.Context, source *models.Source) {
	r.logger.Debug("caching subreddit async", zap.String("subreddit", source.Name))
	// Avoid evaluation here as it's a sync call
	// Get evaluation
	if len(source.Metadata.Rules) > 0 {
		evaluation, usage, err := r.aiClient.GetSourceCommunityRulesEvaluation(ctx, "", source, r.logger)
		if err != nil {
			r.logger.Error("failed to get rules evaluation of source", zap.Error(err))
			return
		}

		source.Metadata.RulesEvaluation = evaluation
		source.Metadata.RulesEvaluation.ModelUsed = usage.Model

		if evaluation != nil {
			err := r.db.UpdateSource(ctx, source)
			if err != nil {
				r.logger.Error("failed to update subreddit rules", zap.Error(err))
				return
			} else {
				r.logger.Debug("subreddit rules updated", zap.String("subreddit", source.Name))
			}
		}
	}

	err := r.cache.Set(ctx, source.GetCacheKey(), source, 0)
	if err != nil {
		r.logger.Error("failed to cache subreddit", zap.Error(err))
	} else {
		r.logger.Debug("subreddit cached", zap.String("subreddit", source.Name))
	}
}

func (r redditService) CreateSubReddit(ctx context.Context, source *models.Source) error {
	// Check if the subreddit already exists in the DB
	existingSubreddit, err := r.db.GetSourceByName(ctx, source.Name, source.ProjectID)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return fmt.Errorf("get existing subreddit: %w", err)
	}

	if existingSubreddit != nil {
		return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("subreddit already exists"))
	}

	value, err := r.cache.Get(ctx, source.GetCacheKey())
	if err != nil {
		r.logger.Warn("failed to get subreddit from cache", zap.Error(err))
	}

	if value != nil {
		globalSource := models.Source{}
		err := json.Unmarshal(value, &globalSource)
		if err != nil {
			return fmt.Errorf("failed to unmarshal cached subreddit: %w", err)
		}

		source.ExternalID = globalSource.ExternalID
		source.Name = globalSource.Name
		source.Description = globalSource.Description
		source.Metadata = globalSource.Metadata
	} else {
		// Fetch the subreddit details by URL from Reddit API
		subRedditDetails, err := r.redditClient.GetSubRedditByName(ctx, source.Name)
		if err != nil {
			return err
		}

		// Fill in the fields in models.Source using fetched details
		if subRedditDetails.ID != "" {
			source.ExternalID = utils.Ptr(subRedditDetails.ID)
		}
		source.Name = subRedditDetails.DisplayName
		source.Description = subRedditDetails.Description
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
	}

	source.SourceType = models.SourceTypeSUBREDDIT

	// Insert the subreddit into the DB
	createdSource, err := r.db.AddSource(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to add subreddit to the database: %w", err)
	}

	source.ID = createdSource.ID

	// If not cached before
	if value == nil {
		go r.cacheSubReddit(context.Background(), createdSource)
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
