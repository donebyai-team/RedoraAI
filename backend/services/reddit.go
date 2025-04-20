package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"time"
)

type RedditService interface {
	CreateSubReddit(ctx context.Context, subReddit *models.SubReddit) error
	GetSubReddits(ctx context.Context, orgID string) ([]*models.SubReddit, error)
	RemoveSubReddit(ctx context.Context, id string) error
}

type redditService struct {
	db           datastore.Repository
	redditClient *reddit.Client
	logger       *zap.Logger
}

func NewRedditService(logger *zap.Logger, db datastore.Repository, redditClient *reddit.Client) *redditService {
	return &redditService{logger: logger, db: db, redditClient: redditClient}
}

func (r redditService) CreateSubReddit(ctx context.Context, subReddit *models.SubReddit) error {
	// Check if the subreddit already exists in the DB
	existingSubreddit, err := r.db.GetSubRedditByUrl(ctx, subReddit.URL, subReddit.ProjectID)
	if !errors.Is(err, datastore.NotFound) {
		return fmt.Errorf("get existing subreddit: %w", err)
	}

	if existingSubreddit != nil {
		return fmt.Errorf("subreddit already exists: %s", subReddit.URL)
	}

	// Fetch the subreddit details by URL from Reddit API
	subRedditDetails, err := r.redditClient.GetSubRedditByURL(ctx, subReddit.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch subreddit details from Reddit: %w", err)
	}

	if subRedditDetails.ID == "" {
		return fmt.Errorf("subreddit ID is missing or invalid subreddit URL: %s", subReddit.URL)
	}

	// Fill in the fields in models.SubReddit using fetched details
	subReddit.SubRedditID = subRedditDetails.ID
	subReddit.URL = subRedditDetails.URL
	subReddit.Name = subRedditDetails.DisplayName
	subReddit.Description = subRedditDetails.Description
	subReddit.SubredditCreatedAt = time.Unix(int64(subRedditDetails.CreatedAt), 0)
	subReddit.Subscribers = utils.Ptr(subRedditDetails.Subscribers)
	subReddit.Title = utils.Ptr(subRedditDetails.Title)

	// Insert the subreddit into the DB
	_, err = r.db.AddSubReddit(ctx, subReddit)
	if err != nil {
		return fmt.Errorf("failed to add subreddit to the database: %w", err)
	}

	return nil
}

func (r redditService) GetSubReddits(ctx context.Context, projectID string) ([]*models.SubReddit, error) {
	return r.db.GetSubRedditsByProject(ctx, projectID)
}

func (r redditService) RemoveSubReddit(ctx context.Context, id string) error {
	// Step 1: Try to get the subreddit to check if it exists
	subreddit, err := r.db.GetSubRedditByID(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to fetch subreddit with ID %s: %w", id, err)
	}
	if subreddit == nil {
		return fmt.Errorf("no subreddit found under the project")
	}

	// Step 2: Delete the subreddit
	_, err = r.db.DeleteSubRedditByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete subreddit with ID %s: %w", id, err)
	}

	return nil
}
