package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"keyword/create_keyword.sql",
		"keyword/query_keyword_by_org.sql",
		"sub_reddit/query_sub_reddit_by_filter.sql",
	})
}

func (r *Database) CreateKeyword(ctx context.Context, keywords *models.Keyword) (*models.Keyword, error) {
	stmt := r.mustGetStmt("keyword/create_keyword.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"keyword":    keywords.Keyword,
		"project_id": keywords.ProjectID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	keywords.ID = id
	return keywords, nil
}

func (r *Database) GetKeywords(ctx context.Context, orgID string) ([]*models.Keyword, error) {
	return getMany[models.Keyword](ctx, r, "keyword/query_keyword_by_org.sql", map[string]any{
		"project_id": orgID,
	})
}

func (r *Database) AddSubReddit(ctx context.Context, subreddit *models.SubReddit) (*models.SubReddit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *Database) GetSubReddits(ctx context.Context) ([]*models.AugmentedSubReddit, error) {
	subReddits, err := getMany[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_filter.sql", map[string]any{})

	if err != nil {
		return nil, fmt.Errorf("failed to get customer cases: %w", err)
	}
	var results []*models.AugmentedSubReddit
	for _, subreddit := range subReddits {
		keywords, err := r.GetKeywords(ctx, subreddit.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get keywords for project %q: %w", subreddit.ProjectID, err)
		}

		project, err := r.GetProject(ctx, subreddit.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %q: %w", subreddit.ProjectID, err)
		}

		results = append(results, &models.AugmentedSubReddit{
			SubReddit: subreddit,
			Keywords:  keywords,
			Project:   project,
		})
	}

	return results, nil
}
