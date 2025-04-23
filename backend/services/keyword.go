package services

import (
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"golang.org/x/net/context"
)

type CreateKeyword struct {
	Keyword   string
	ProjectID string
}

type KeywordService interface {
	CreateKeyword(ctx context.Context, session *CreateKeyword) (*models.Keyword, error)
}

type KeywordServiceImpl struct {
	db datastore.Repository
}

func NewKeywordServiceImpl(db datastore.Repository) *KeywordServiceImpl {
	return &KeywordServiceImpl{db: db}
}

func (c *KeywordServiceImpl) CreateKeyword(ctx context.Context, session *CreateKeyword) (*models.Keyword, error) {
	sanitizeKeyword := utils.SanitizeKeyword(session.Keyword)
	if sanitizeKeyword == "" {
		return nil, fmt.Errorf("invalid keyword [%s]", session.Keyword)
	}

	keyword, err := c.db.CreateKeyword(ctx, &models.Keyword{
		Keyword:   session.Keyword,
		ProjectID: session.ProjectID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	return keyword, nil
}
