package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"keywords/keywords.sql",
	})
}

func (r *Database) CreateKeyword(ctx context.Context, keywords *models.Keyword) (*models.Keyword, error) {
	stmt := r.mustGetStmt("keywords/keywords.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"keyword":         keywords.Keyword,
		"organization_id": "dd80c434-43c3-41aa-84a6-fc908e7b36ad",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	keywords.ID = id
	return keywords, nil
}
