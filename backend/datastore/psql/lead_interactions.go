package psql

import (
	"context"
	"github.com/shank318/doota/models"
	"time"
)

func init() {
	registerFiles([]string{
		"lead_interactions/create_lead_interaction.sql",
		"lead_interactions/update_lead_interaction.sql",
		"lead_interactions/query_interaction_by_project.sql",
	})
}

func (r *Database) CreateLeadInteraction(ctx context.Context, reddit *models.LeadInteraction) (*models.LeadInteraction, error) {
	stmt := r.mustGetStmt("lead_interactions/create_lead_interaction.sql")
	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id": reddit.ProjectID,
		"lead_id":    reddit.LeadID,
		"type":       reddit.Type,
		"from_user":  reddit.From,
		"to_user":    reddit.To,
		"metadata":   reddit.Metadata,
		"reason":     reddit.Reason,
	})
	reddit.ID = id
	return reddit, err
}

func (r *Database) GetLeadInteractions(ctx context.Context, projectID string, start, end time.Time) ([]*models.LeadInteraction, error) {
	return getMany[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_project.sql", map[string]any{
		"project_id": projectID,
		"start_date": start.Format(time.DateOnly),
		"end_date":   end.Format(time.DateOnly),
	})
}

func (r *Database) UpdateLeadInteraction(ctx context.Context, reddit *models.LeadInteraction) error {
	stmt := r.mustGetStmt("lead_interactions/update_lead_interaction.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"project_id": reddit.ProjectID,
		"id":         reddit.ID,
		"status":     reddit.Status,
		"reason":     reddit.Reason,
		"metadata":   reddit.Metadata,
	})
	return err
}
