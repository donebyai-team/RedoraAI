package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"time"
)

func init() {
	registerFiles([]string{
		"keyword/query_keyword_by_id.sql",
		"keyword/create_keyword.sql",
		"keyword/query_keyword_by_project.sql",
		"keyword/create_keyword_tracker.sql",
		"keyword/query_keyword_tracker_by_filter.sql",
		"keyword/update_keyword_tracker_last_tracked_at.sql",
		"keyword/query_keyword_tracker_by_project.sql",
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

func (r *Database) GetKeywords(ctx context.Context, projectID string) ([]*models.Keyword, error) {
	return getMany[models.Keyword](ctx, r, "keyword/query_keyword_by_project.sql", map[string]any{
		"project_id": projectID,
	})
}

func (r *Database) GetKeywordByID(ctx context.Context, id string) (*models.Keyword, error) {
	return getOne[models.Keyword](ctx, r, "keyword/query_keyword_by_id.sql", map[string]any{
		"id": id,
	})
}

func (r *Database) GetKeywordTrackerByProjectID(ctx context.Context, projectID string) ([]*models.KeywordTracker, error) {
	return getMany[models.KeywordTracker](ctx, r, "keyword/query_keyword_tracker_by_project.sql", map[string]any{
		"project_id": projectID,
	})
}

func (r *Database) UpdatKeywordTrackerLastTrackedAt(ctx context.Context, id string) error {
	stmt := r.mustGetStmt("keyword/update_keyword_tracker_last_tracked_at.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"id":              id,
		"last_tracked_at": time.Now(),
	})
	return err
}

func (r *Database) GetKeywordTrackers(ctx context.Context) ([]*models.AugmentedKeywordTracker, error) {
	trackers, err := getMany[models.KeywordTracker](ctx, r, "keyword/query_keyword_tracker_by_filter.sql", map[string]any{})

	if err != nil {
		return nil, fmt.Errorf("failed to get keyword trackers to track: %w", err)
	}
	var results []*models.AugmentedKeywordTracker
	for _, tracker := range trackers {
		keyword, err := r.GetKeywordByID(ctx, tracker.KeywordID)
		if err != nil && errors.Is(err, datastore.NotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get keyword for project %q: %w", tracker.KeywordID, err)
		}

		source, err := r.GetSourceByID(ctx, tracker.SourceID)
		if err != nil && errors.Is(err, datastore.NotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get source for project %q: %w", tracker.KeywordID, err)
		}

		project, err := r.GetProject(ctx, tracker.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %q: %w", keyword.ProjectID, err)
		}

		results = append(results, &models.AugmentedKeywordTracker{
			Tracker: tracker,
			Source:  source,
			Keyword: keyword,
			Project: project,
		})
	}

	return results, nil
}

func (r *Database) CreateKeywordTracker(ctx context.Context, tracker *models.KeywordTracker) (*models.KeywordTracker, error) {
	stmt := r.mustGetStmt("keyword/create_keyword_tracker.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"keyword_id": tracker.KeywordID,
		"source_id":  tracker.SourceID,
	})

	if err != nil {
		return nil, err
	}

	tracker.ID = id
	return tracker, nil
}
