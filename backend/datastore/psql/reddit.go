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
		"sub_reddit/create_sub_reddit.sql",
		"sub_reddit/query_sub_reddit_by_url.sql",
		"sub_reddit/delete_sub_reddit_by_id.sql",
		"sub_reddit/query_sub_reddit_by_id.sql",
		"sub_reddit/query_sub_reddit_by_project.sql",
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
	stmt := r.mustGetStmt("sub_reddit/create_sub_reddit.sql")

	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"subreddit_id":         subreddit.SubRedditID,
		"url":                  subreddit.URL,
		"name":                 subreddit.Name,
		"description":          subreddit.Description,
		"project_id":           subreddit.ProjectID,
		"subreddit_created_at": subreddit.SubredditCreatedAt,
		"subscribers":          subreddit.Subscribers,
		"title":                subreddit.Title,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert subreddit: %w", err)
	}

	subreddit.ID = id
	return subreddit, nil
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

func (r *Database) GetSubRedditsByProject(ctx context.Context, projectID string) ([]*models.SubReddit, error) {
	return getMany[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_project.sql", map[string]any{
		"project_id": projectID,
	})
}

func (r *Database) GetSubRedditByUrl(ctx context.Context, url, projectID string) (*models.SubReddit, error) {
	return getOne[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_url.sql", map[string]any{
		"url":        url,
		"project_id": projectID,
	})
}

func (r *Database) GetSubRedditByID(ctx context.Context, ID string) (*models.SubReddit, error) {
	return getOne[models.SubReddit](ctx, r, "sub_reddit/query_sub_reddit_by_id.sql", map[string]any{
		"id": ID,
	})
}

func (r *Database) DeleteSubRedditByID(ctx context.Context, ID string) (*models.SubReddit, error) {
	stmt := r.mustGetStmt("sub_reddit/delete_sub_reddit_by_id.sql")

	var subreddit models.SubReddit
	err := stmt.QueryRowxContext(ctx, map[string]interface{}{
		"id": ID,
	}).StructScan(&subreddit)
	if err != nil {
		return nil, fmt.Errorf("failed to delete subreddit with ID %s: %w", ID, err)
	}

	return &subreddit, nil
}
