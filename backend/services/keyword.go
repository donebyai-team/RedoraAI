package services

import (
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"golang.org/x/net/context"
)

type CreateKeyword struct {
	Keyword string
	OrgID   string
}

type KeywordService interface {
	CreateKeyword(ctx context.Context, session *CreateKeyword) (*models.Keyword, error)
}

type CreateKeywordImpl struct {
	db datastore.Repository
}

func NewCreateKeywordImpl(db datastore.Repository) *CreateKeywordImpl {
	return &CreateKeywordImpl{db: db}
}

func (c *CreateKeywordImpl) CreateKeyword(ctx context.Context, session *CreateKeyword) (*models.Keyword, error) {
	keyword, err := c.db.CreateKeyword(context.Background(), &models.Keyword{
		Keyword: session.Keyword,
		OrgID:   session.OrgID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	return keyword, nil
}
