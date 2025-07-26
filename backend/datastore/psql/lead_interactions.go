package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"time"
)

func init() {
	registerFiles([]string{
		"lead_interactions/create_lead_interaction.sql",
		"lead_interactions/update_lead_interaction.sql",
		"lead_interactions/query_interaction_by_project.sql",
		"lead_interactions/query_interaction_to_execute.sql",
		"lead_interactions/set_interaction_status_processing.sql",
		"lead_interactions/query_interaction_by_to_from.sql",
		"lead_interactions/query_interactions.sql",
		"lead_interactions/query_interaction_by_lead_id.sql",
		"lead_interactions/query_interaction_by_id.sql",
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

func (r *Database) GetLeadInteractions(ctx context.Context, projectID string, status models.LeadInteractionStatus, dateRange pbportal.DateRangeFilter) ([]*models.LeadInteraction, error) {
	startDateTime, endDateTime := GetDateRange(dateRange, time.Now().UTC())
	return getMany[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_project.sql", map[string]any{
		"project_id":     projectID,
		"status":         status,
		"start_datetime": sqlNullTime(startDateTime),
		"end_datetime":   sqlNullTime(endDateTime),
	})
}

func (r *Database) GetAugmentedLeadInteractions(ctx context.Context, projectID string, dateRange pbportal.DateRangeFilter) ([]*models.AugmentedLeadInteraction, error) {
	startDateTime, endDateTime := GetDateRange(dateRange, time.Now().UTC())
	return getMany[models.AugmentedLeadInteraction](ctx, r, "lead_interactions/query_interactions.sql", map[string]any{
		"project_id":     projectID,
		"start_datetime": sqlNullTime(startDateTime),
		"end_datetime":   sqlNullTime(endDateTime),
	})
}

func (r *Database) IsInteractionExists(ctx context.Context, interaction *models.LeadInteraction) (bool, error) {
	one, err := getOne[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_to_from.sql", map[string]any{
		"project_id": interaction.ProjectID,
		"status":     models.LeadInteractionStatusSENT,
		"type":       interaction.Type,
		"from_user":  interaction.From,
		"to_user":    interaction.To,
	})
	if err != nil {
		if errors.Is(err, datastore.NotFound) {
			return false, nil
		}
		return false, err
	}

	return one != nil, nil
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
			interaction.Organization = organization
			orgCache[organization.ID] = organization
			projectToOrg[projectID] = organization.ID
		}
	}

	return interactions, nil
}

func (r *Database) GetLeadInteractionByID(ctx context.Context, id string) (*models.LeadInteraction, error) {
	return getOne[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_id.sql", map[string]any{
		"id": id,
	})
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

func (r *Database) GetLeadInteractionByLeadID(ctx context.Context, leadID string) ([]*models.LeadInteraction, error) {
	return getMany[models.LeadInteraction](ctx, r, "lead_interactions/query_interaction_by_lead_id.sql", map[string]any{
		"lead_id": leadID,
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
		"from_user":  reddit.From,
	})
	return err
}
