package psql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
)

func init() {
	registerFiles([]string{
		"post/create_post.sql",
		"post/query_post_by_id.sql",
		"post/update_post.sql",
	})
}

func (r *Database) CreatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	stmt := r.mustGetStmt("post/create_post.sql")

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(post.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post metadata: %w", err)
	}

	var id string
	err = stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id":   post.ProjectID,
		"title":        post.Title,
		"description":  post.Description,
		"source_id":    post.SourceID,
		"status":       post.Status,
		"metadata":     metadataJSON,
		"reason":       post.Reason,
		"reference_id": utils.NullableUUID(post.ReferenceID), // optional
		"schedule_at":  post.ScheduleAt,                      // nullable
	})
	if err != nil {
		fmt.Println("error while creating post", err)
		return nil, err
	}

	r.zlogger.Info("Post created successfully", zap.String("post_id", id))

	// Set the returned ID and return the populated object
	post.ID = id
	return post, nil
}

func (r *Database) GetPostByID(ctx context.Context, ID string) (*models.Post, error) {
	return getOne[models.Post](ctx, r, "post/query_post_by_id.sql", map[string]any{
		"id": ID,
	})
}

func (r *Database) UpdatePost(ctx context.Context, post *models.Post) error {
	stmt := r.mustGetStmt("post/update_post.sql")

	metadataJSON, err := json.Marshal(post.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal post metadata: %w", err)
	}

	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"description":  post.Description,
		"status":       post.Status,
		"reason":       post.Reason,
		"metadata":     metadataJSON,
		"reference_id": post.ReferenceID,
		"schedule_at":  post.ScheduleAt,
	})
	return err
}
