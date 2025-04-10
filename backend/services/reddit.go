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
	// Fetch the subreddit details by URL
	subRedditDetails, err := r.redditClient.GetSubRedditByURL(ctx, subReddit.URL)
	if err != nil {
		return err
	}
	fmt.Println("data is", subRedditDetails)

	// Check if the subreddit already exists in the DB
	//existingSubReddit, err := r.db.GetSubRedditByID(ctx, subReddit.SubRedditID)
	//if err != nil && !errors.Is(err, datastore.NotFound) {
	//	return fmt.Errorf("failed to get subreddit by ID: %w", err)
	//}
	//
	//// If subreddit already exists in DB, return an error
	//if existingSubReddit != nil {
	//	return fmt.Errorf("subreddit already exists in the database: %s", subReddit.SubRedditID)
	//}

	// Fill the fields in models.SubReddit using fetched details
	subReddit.Name = subRedditDetails.DisplayName
	subReddit.Description = subRedditDetails.Description
	//*subReddit.Subscribers = int64(subRedditDetails.Subscribers)
	subReddit.SubredditCreatedAt = time.Unix(int64(subRedditDetails.CreatedRaw), 0)
	//subReddit.LastTrackedAt = subRedditDetails.l
	//subReddit.LastPostCreatedAt = subRedditDetails.LastPostCreatedAt

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
