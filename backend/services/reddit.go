package services

import (
	"context"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
)

type RedditService interface {
	CreateSubReddit(ctx context.Context, subReddit *models.SubReddit) error
	GetSubReddits(ctx context.Context) ([]*models.SubReddit, error)
}

type redditService struct {
	db           datastore.Repository
	redditClient reddit.Client
}

func (r redditService) CreateSubReddit(ctx context.Context, subReddit *models.SubReddit) error {
	//TODO implement me
	panic("implement me")
	// Check if the requested subreddit exists
	// if yes, check it should not exists in our DB
	// if yes, fetch the details, fill fields in  models.SubReddit and store it in DB
}

func (r redditService) GetSubReddits(ctx context.Context) ([]*models.SubReddit, error) {
	//TODO implement me
	panic("implement me")
}
