package psql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"time"
)

func init() {
	registerFiles([]string{
		"leads/create_lead.sql",
		"leads/query_lead_by_filter.sql",
		"leads/query_lead_by_post_id.sql",
		"leads/query_lead_by_status.sql",
		"leads/update_lead_status.sql",
		"leads/query_lead_by_id.sql",
		"leads/count_lead_by_created_at.sql",
	})
}

func (r *Database) CountLeadByCreatedAt(ctx context.Context, projectID string, relevancyScore int, dateRange pbportal.DateRangeFilter) (*models.LeadsData, error) {
	startDateTime, endDateTime := GetDateRange(dateRange, time.Now().UTC())
	return getOne[models.LeadsData](ctx, r, "leads/count_lead_by_created_at.sql", map[string]any{
		"relevancy_score": relevancyScore,
		"project_id":      projectID,
		"start_datetime":  sqlNullTime(startDateTime),
		"end_datetime":    sqlNullTime(endDateTime),
	})
}

func (r *Database) GetLeadByPostID(ctx context.Context, projectID, postID string) (*models.Lead, error) {
	return getOne[models.Lead](ctx, r, "leads/query_lead_by_post_id.sql", map[string]any{
		"post_id":    postID,
		"project_id": projectID,
	})
}

func (r *Database) GetLeadByID(ctx context.Context, projectID, id string) (*models.Lead, error) {
	return getOne[models.Lead](ctx, r, "leads/query_lead_by_id.sql", map[string]any{
		"id":         id,
		"project_id": projectID,
	})
}

func (r *Database) UpdateLeadStatus(ctx context.Context, lead *models.Lead) error {
	if lead.Status == "" {
		return fmt.Errorf("status cannot be empty")
	}

	stmt := r.mustGetStmt("leads/update_lead_status.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"status":     lead.Status,
		"project_id": lead.ProjectID,
		"metadata":   lead.LeadMetadata,
		"id":         lead.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update lead status %q: %w", lead.ID, err)
	}
	return nil
}

func (r *Database) GetLeadsByStatus(ctx context.Context, projectID string, filter datastore.LeadsFilter) ([]*models.AugmentedLead, error) {
	startDateTime, endDateTime := GetDateRange(filter.DateRange, time.Now().UTC())
	return getMany[models.AugmentedLead](ctx, r, "leads/query_lead_by_status.sql", map[string]any{
		"status":         filter.Status,
		"project_id":     projectID,
		"limit":          filter.Limit,
		"offset":         filter.Offset * filter.Limit,
		"start_datetime": sqlNullTime(startDateTime),
		"end_datetime":   sqlNullTime(endDateTime),
	})
}

func (r *Database) GetLeadsByRelevancy(ctx context.Context, projectID string, filter datastore.LeadsFilter) ([]*models.AugmentedLead, error) {
	startDateTime, endDateTime := GetDateRange(filter.DateRange, time.Now().UTC())
	return getMany[models.AugmentedLead](ctx, r, "leads/query_lead_by_filter.sql", map[string]any{
		"source_ids":      pq.Array(filter.Sources),
		"relevancy_score": filter.RelevancyScore,
		"status":          models.LeadStatusNEW,
		"project_id":      projectID,
		"limit":           filter.Limit,
		"offset":          filter.Offset * filter.Limit,
		"start_datetime":  sqlNullTime(startDateTime),
		"end_datetime":    sqlNullTime(endDateTime),
	})
}

func sqlNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func (r *Database) GetLeadByCommentID(ctx context.Context, projectID, commentID string) (*models.Lead, error) {
	panic("implement me")
}

func (r *Database) CreateLead(ctx context.Context, reddit *models.Lead) (*models.Lead, error) {
	stmt := r.mustGetStmt("leads/create_lead.sql")
	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id":      reddit.ProjectID,
		"source_id":       reddit.SourceID,
		"author":          reddit.Author,
		"post_id":         reddit.PostID,
		"keyword_id":      reddit.KeywordID,
		"intents":         reddit.Intents,
		"type":            reddit.Type,
		"relevancy_score": reddit.RelevancyScore,
		"post_created_at": reddit.PostCreatedAt,
		"metadata":        reddit.LeadMetadata,
		"description":     reddit.Description,
		"title":           reddit.Title,
	})
	reddit.ID = id
	return reddit, err
}

// GetDateRange returns the start and end time range for filters like "today", "yesterday", "7_days_ago".
// If the filter is unknown, it returns (nil, nil).
func GetDateRange(filter pbportal.DateRangeFilter, now time.Time) (*time.Time, *time.Time) {
	loc := now.Location()
	switch filter {
	case pbportal.DateRangeFilter_DATE_RANGE_TODAY:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		return &start, &end
	case pbportal.DateRangeFilter_DATE_RANGE_YESTERDAY:
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		start := end.AddDate(0, 0, -1)
		return &start, &end
	case pbportal.DateRangeFilter_DATE_RANGE_7_DAYS:
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -7)
		return &start, &end
	default:
		return nil, nil
	}
}
