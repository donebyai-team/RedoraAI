package psql

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/shank318/doota/models"
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

func (r *Database) CountLeadByCreatedAt(ctx context.Context, projectID string, relevancyScore int, start, end time.Time) (*models.LeadsData, error) {
	return getOne[models.LeadsData](ctx, r, "leads/count_lead_by_created_at.sql", map[string]any{
		"relevancy_score": relevancyScore,
		"project_id":      projectID,
		"start_date":      start.Format(time.DateOnly),
		"end_date":        end.Format(time.DateOnly),
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

func (r *Database) GetLeadsByStatus(ctx context.Context, projectID string, status models.LeadStatus) ([]*models.AugmentedLead, error) {
	return getMany[models.AugmentedLead](ctx, r, "leads/query_lead_by_status.sql", map[string]any{
		"status":     status,
		"project_id": projectID,
	})
}

func (r *Database) GetLeadsByRelevancy(ctx context.Context, projectID string, relevancy float32, sources []string) ([]*models.AugmentedLead, error) {
	return getMany[models.AugmentedLead](ctx, r, "leads/query_lead_by_filter.sql", map[string]any{
		"source_ids":      pq.Array(sources),
		"relevancy_score": relevancy,
		"status":          models.LeadStatusNEW,
		"project_id":      projectID,
	})
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
