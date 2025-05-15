package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	"strings"
)

func init() {
	registerFiles([]string{
		"source/create_source.sql",
		"source/query_source_by_name.sql",
		"source/delete_source_by_id.sql",
		"source/query_source_by_id.sql",
		"source/query_source_by_project.sql",
		"source/update_source.sql",
		"keyword/delete_keyword_tracker_by_source_id.sql",
		"keyword/create_keyword_tracker.sql",
	})
}

func (r *Database) UpdateSource(ctx context.Context, subreddit *models.Source) error {
	stmt := r.mustGetStmt("source/update_source.sql")

	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"id":       subreddit.ID,
		"metadata": subreddit.Metadata,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Database) AddSource(ctx context.Context, subreddit *models.Source) (*models.Source, error) {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	stmt := r.mustGetTxStmt(ctx, "source/create_source.sql", tx)

	var id string
	err = stmt.GetContext(ctx, &id, map[string]interface{}{
		"external_id": subreddit.ExternalID,
		"name":        strings.ToLower(subreddit.Name),
		"description": subreddit.Description,
		"project_id":  subreddit.ProjectID,
		"source_type": subreddit.SourceType,
		"metadata":    subreddit.Metadata,
	})
	if err != nil {
		return nil, err
	}

	subreddit.ID = id

	// Add trackers
	// Create trackers for each keyword
	keywords, err := r.GetKeywords(ctx, subreddit.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keywords by project: %w", err)
	}

	stmtKeywordTracker := r.mustGetTxStmt(ctx, "keyword/create_keyword_tracker.sql", tx)

	for _, keyword := range keywords {
		_, err = stmtKeywordTracker.ExecContext(ctx, map[string]interface{}{
			"keyword_id": keyword.ID,
			"source_id":  subreddit.ID,
			"project_id": subreddit.ProjectID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add keyword tracker for keyword [%s]: %w", keyword.Keyword, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return subreddit, nil
}

func (r *Database) GetSourcesByProject(ctx context.Context, projectID string) ([]*models.Source, error) {
	return getMany[models.Source](ctx, r, "source/query_source_by_project.sql", map[string]any{
		"project_id": projectID,
	})
}

func (r *Database) GetSourceByName(ctx context.Context, name, projectID string) (*models.Source, error) {
	return getOne[models.Source](ctx, r, "source/query_source_by_name.sql", map[string]any{
		"name":       name,
		"project_id": projectID,
	})
}

func (r *Database) GetSourceByID(ctx context.Context, ID string) (*models.Source, error) {
	return getOne[models.Source](ctx, r, "source/query_source_by_id.sql", map[string]any{
		"id": ID,
	})
}

func (r *Database) DeleteSourceByID(ctx context.Context, id string) error {
	tx, err := r.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err = executePotentialRollback(tx, err)
	}()

	// Delete source
	stmt := r.mustGetTxStmt(ctx, "source/delete_source_by_id.sql", tx)
	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"id": id,
	})

	// Delete tracker
	stmt = r.mustGetTxStmt(ctx, "keyword/delete_keyword_tracker_by_source_id.sql", tx)
	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"source_id": id,
	})

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return err
}
