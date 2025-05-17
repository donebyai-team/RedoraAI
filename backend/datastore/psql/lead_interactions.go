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
		"lead_interactions/create_lead_interaction.sql",
		"lead_interactions/update_lead_interaction.sql",
		"lead_interactions/query_interaction_by_project.sql",
		"lead_interactions/query_interaction_to_execute.sql",
		"lead_interactions/set_interaction_status_processing.sql",
	})
}

func (r *Database) CreateLeadInteraction(ctx context.Context, reddit *models.LeadInteraction) (*models.LeadInteraction, error) {
	stmt := r.mustGetStmt("lead_interactions/create_lead_interaction.sql")
	var id string
	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"project_id":  reddit.ProjectID,
		"lead_id":     reddit.LeadID,
		"schedule_at": reddit.ScheduledAt,
		"type":        reddit.Type,
		"from_user":   reddit.From,
		"to_user":     reddit.To,
		"metadata":    reddit.Metadata,
		"reason":      reddit.Reason,
	})
	reddit.ID = id
	return reddit, err
}

func (r *Database) GetLeadInteractions(ctx context.Context, projectID string, status models.LeadInteractionStatus, start, end time.Time) ([]*models.LeadInteraction, error) {
	return getMany[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_project.sql", map[string]any{
		"project_id": projectID,
		"status":     status,
		"start_date": start.Format(time.DateOnly),
		"end_date":   end.Format(time.DateOnly),
	})
}

func (r *Database) GetLeadInteractionsToExecute(ctx context.Context, statuses []models.LeadInteractionStatus) ([]*models.LeadInteraction, error) {
	interactions, err := getMany[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_to_execute.sql", map[string]any{
		"statuses": pq.Array(statuses),
	})

	if err != nil {
		return nil, err
	}

	orgCache := map[string]*models.Organization{} // org_id -> Org
	projectToOrg := map[string]string{}           // project_id -> org_id

	for _, interaction := range interactions {
		projectID := interaction.ProjectID

		// Get org_id from project metadata (maybe in a separate call or map)
		orgID := projectToOrg[projectID]

		if org, ok := orgCache[orgID]; ok {
			// Already loaded
			interaction.Organization = org
		} else {
			// Fetch and cache
			project, err := r.GetProject(ctx, projectID)
			if err != nil {
				return nil, err
			}

			organization, err := r.GetOrganizationById(ctx, project.OrganizationID)
			if err != nil {
				return nil, err
			}

			orgCache[orgID] = organization
			interaction.Organization = organization
		}
	}

	return interactions, nil
}

func (r *Database) SetLeadInteractionStatusProcessing(ctx context.Context, id string) error {
	stmt := r.mustGetStmt("lead_interactions/set_interaction_status_processing.sql")
	res, err := stmt.ExecContext(ctx, map[string]interface{}{
		"id": id,
	})

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("status not updated: current status is not CREATED")
	}

	return nil
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
