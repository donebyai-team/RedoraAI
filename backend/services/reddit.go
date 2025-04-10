package services

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
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
	existingSubreddit, err := r.db.GetSubRedditByUrl(ctx, subReddit.URL, subReddit.OrganizationID)

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
	subscribers := subRedditDetails.Subscribers
	subReddit.Subscribers = &subscribers
	title := subRedditDetails.Title
	subReddit.Title = &title

	// Insert the subreddit into the DB
	_, err = r.db.AddSubReddit(ctx, subReddit)
	if err != nil {
		return fmt.Errorf("failed to add subreddit to the database: %w", err)
	}

	return nil
}

func (r redditService) GetSubReddits(ctx context.Context, orgID string) ([]*models.SubReddit, error) {
	//TODO implement me
	panic("implement me")
	// Query the subreddits from Db and return
}

func (r redditService) RemoveSubReddit(ctx context.Context, id string) error {
	panic("implement me")
	// Query if the subreddit exists
	// if yes, delete it
}
