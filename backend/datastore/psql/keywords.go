package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"strings"
	"time"
)

func init() {
	registerFiles([]string{
		"keyword/query_keyword_by_id.sql",
		"keyword/create_keyword.sql",
		"keyword/query_keyword_by_project.sql",
		"keyword/create_keyword_tracker.sql",
		"keyword/delete_keyword_by_id.sql",
		"keyword/delete_keyword_tracker_by_keyword.sql",
		"keyword/query_keyword_tracker_by_filter.sql",
		"keyword/update_keyword_tracker_last_tracked_at.sql",
		"keyword/query_keyword_tracker_by_project.sql",
	})
}

func (r *Database) RemoveKeyword(ctx context.Context, projectID, keywordID string) error {
	stmt := r.mustGetStmt("keyword/delete_keyword_by_id.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"project_id": projectID,
		"id":         keywordID,
	})
	return err
}

func diffKeywords(existing []*models.Keyword, newKeywords []string) (toAdd []string, toDelete []*models.Keyword) {
	existingMap := make(map[string]*models.Keyword)
	newMap := make(map[string]bool)

	// Normalize and store existing keywords
	for _, kw := range existing {
		if kw == nil {
			continue
		}
		lowered := strings.ToLower(strings.TrimSpace(kw.Keyword))
		existingMap[lowered] = kw
	}

	// Normalize and map new keywords
	for _, kw := range newKeywords {
		lowered := strings.ToLower(strings.TrimSpace(kw))
		newMap[lowered] = true

		// If new keyword not in existing, add to toAdd
		if _, exists := existingMap[lowered]; !exists {
			toAdd = append(toAdd, kw)
		}
	}

	// Anything in existing not in new -> delete
	for lowered, kw := range existingMap {
		if _, exists := newMap[lowered]; !exists {
			toDelete = append(toDelete, kw)
		}
	}

	return
}

func (r *Database) CreateKeywords(ctx context.Context, projectID string, keywords []string) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	existingKeywords, err := r.GetKeywords(ctx, projectID)
	if err != nil {
		return err
	}

	toAdd, toDelete := diffKeywords(existingKeywords, keywords)

	if len(toAdd) == 0 && len(toDelete) == 0 {
		return nil
	}

	// Add
	stmt := r.mustGetTxStmt(ctx, "keyword/create_keyword.sql", tx)
	stmtTracker := r.mustGetTxStmt(ctx, "keyword/create_keyword_tracker.sql", tx)

	// Add tracker for each source
	sources, err := r.GetSourcesByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch sources by project: %w", err)
	}

	for _, kw := range toAdd {
		var id string
		// Add keyword
		err := stmt.GetContext(ctx, &id, map[string]interface{}{
			"keyword":    kw,
			"project_id": projectID,
		})
		if err != nil {
			return fmt.Errorf("failed to create keyword[%s] for project: %w", kw, err)
		}

		// Add keywords for the sources
		for _, source := range sources {
			_, err := stmtTracker.ExecContext(ctx, map[string]interface{}{
				"keyword_id": id,
				"source_id":  source.ID,
				"project_id": projectID,
			})
			if err != nil {
				return fmt.Errorf("failed to add keyword tracker for keyword [%s]: %w", kw, err)
			}
		}
	}

	// Delete
	stmtDelete := r.mustGetTxStmt(ctx, "keyword/delete_keyword_by_id.sql", tx)
	stmtDeleteTracker := r.mustGetTxStmt(ctx, "keyword/delete_keyword_tracker_by_keyword.sql", tx)
	for _, kw := range toDelete {
		// Delete keyword
		_, err := stmtDelete.ExecContext(ctx, map[string]interface{}{
			"project_id": projectID,
			"id":         kw.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to remove keyword[%s] for project: %w", kw, err)
		}

		// Delete tracker
		_, err = stmtDeleteTracker.ExecContext(ctx, map[string]interface{}{
			"keyword_id": kw.ID,
		})

		if err != nil {
			return fmt.Errorf("failed to remove keyword[%s] for project: %w", kw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return err
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

		org, err := r.GetOrganizationById(ctx, project.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get organization %q: %w", keyword.ProjectID, err)
		}

		results = append(results, &models.AugmentedKeywordTracker{
			Tracker:      tracker,
			Source:       source,
			Keyword:      keyword,
			Project:      project,
			Organization: org,
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
		"project_id": tracker.ProjectID,
	})

	if err != nil {
		return nil, err
	}

	tracker.ID = id
	return tracker, nil
}
