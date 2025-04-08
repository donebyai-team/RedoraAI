package services

import (
	"context"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
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
	//TODO implement me
	panic("implement me")
	// Check if the requested subreddit exists
	// if yes, check it should not exists in our DB
	// if yes, fetch the details, fill fields in  models.SubReddit and store it in DB
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
