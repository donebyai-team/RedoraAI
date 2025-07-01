package psql

import (
	"context"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"time"
)

func init() {
	registerFiles([]string{
		"post_insight/create_post_insight.sql",
		"post_insight/query_insight_by_post_id.sql",
		"post_insight/query_insight_by_project.sql",
	})
}

func (r *Database) CreatePostInsight(ctx context.Context, insight *models.PostInsight) (*models.PostInsight, error) {
	stmt := r.mustGetStmt("post_insight/create_post_insight.sql")
	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id":      insight.ProjectID,
		"post_id":         insight.PostID,
		"topic":           insight.Topic,
		"sentiment":       insight.Sentiment,
		"metadata":        insight.Metadata,
		"relevancy_score": insight.RelevancyScore,
		"source":          insight.Source,
		"highlights":      insight.Highlights,
	})
	insight.ID = id
	return insight, err
}

func (r *Database) GetInsightsByPostID(ctx context.Context, projectID, postID string) ([]*models.PostInsight, error) {
	return getMany[models.PostInsight](ctx, r, "post_insight/query_insight_by_post_id.sql", map[string]any{
		"project_id": projectID,
		"post_id":    postID,
	})
}

func (r *Database) GetInsights(ctx context.Context, projectID string, filter datastore.LeadsFilter) ([]*models.PostInsight, error) {
	startDateTime, endDateTime := GetDateRange(filter.DateRange, time.Now().UTC())
	return getMany[models.PostInsight](ctx, r, "post_insight/query_insight_by_project.sql", map[string]any{
		"relevancy_score": filter.RelevancyScore,
		"project_id":      projectID,
		"limit":           filter.Limit,
		"offset":          filter.Offset * filter.Limit,
		"start_datetime":  sqlNullTime(startDateTime),
		"end_datetime":    sqlNullTime(endDateTime),
	})
}
