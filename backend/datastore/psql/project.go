package psql

import (
	"context"
	"fmt"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"project/create_project.sql",
		"project/query_project_by_org.sql",
		"project/query_project_by_id.sql",
		"project/query_project_by_name.sql",
		"project/update_project.sql",
	})
}

func (r *Database) GetProjectByName(ctx context.Context, name, orgID string) (*models.Project, error) {
	return getOne[models.Project](ctx, r, "project/query_project_by_name.sql", map[string]any{
		"name":            name,
		"organization_id": orgID,
	})
}

func (r *Database) CreateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	stmt := r.mustGetStmt("project/create_project.sql")
	var id string

	err := stmt.GetContext(ctx, &id, map[string]interface{}{
		"name":             project.Name,
		"description":      project.ProductDescription,
		"organization_id":  project.OrganizationID,
		"customer_persona": project.CustomerPersona,
		"goals":            project.EngagementGoals,
		"website":          project.WebsiteURL,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create keyword for organization: %w", err)
	}

	project.ID = id
	return project, nil
}

func (r *Database) UpdateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	stmt := r.mustGetStmt("project/update_project.sql")

	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"name":             project.Name,
		"description":      project.ProductDescription,
		"organization_id":  project.OrganizationID,
		"customer_persona": project.CustomerPersona,
		"id":               project.ID,
		"website":          project.WebsiteURL,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update project organization: %w", err)
	}

	return project, nil
}

func (r *Database) GetProjects(ctx context.Context, orgID string) ([]*models.Project, error) {
	return getMany[models.Project](ctx, r, "project/query_project_by_org.sql", map[string]any{
		"organization_id": orgID,
	})
}

func (r *Database) GetProject(ctx context.Context, id string) (*models.Project, error) {
	return getOne[models.Project](ctx, r, "project/query_project_by_id.sql", map[string]any{
		"id": id,
	})
}
