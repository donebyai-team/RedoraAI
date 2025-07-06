package psql

import (
	"context"
	"fmt"

	"github.com/shank318/doota/models"
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

	metadataVal, err := post.Metadata.Value()
	if err != nil {
		fmt.Println("error while marshaling metadata:", err)
		return nil, err
	}

	if *post.ReferenceID == "" {
		post.ReferenceID = nil
	}

	var id string
	err = stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id":   post.ProjectID,
		"title":        post.Title,
		"description":  post.Description,
		"source_id":    post.SourceID,
		"status":       post.Status,
		"metadata":     metadataVal,
		"reason":       post.Reason,
		"reference_id": post.ReferenceID,
		"schedule_at":  post.ScheduleAt,
	})

	if err != nil {
		fmt.Println("error while creating posts:", err)
		return nil, err
	}

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

	metadataVal, err := post.Metadata.Value()
	if err != nil {
		fmt.Println("error while marshaling metadata:", err)
		return err
	}

	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"description":  post.Description,
		"status":       post.Status,
		"reason":       post.Reason,
		"metadata":     metadataVal,
		"reference_id": post.ReferenceID,
		"schedule_at":  post.ScheduleAt,
	})
	return err
}
