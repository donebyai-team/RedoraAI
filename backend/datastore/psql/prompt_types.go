package psql

import (
	"context"
	"fmt"

	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"prompt_type/create_prompt_type.sql",
		"prompt_type/update_prompt_type.sql",
		"prompt_type/query_prompt_type_by_name.sql",
	})
}

func (r *Database) CreatePromptType(ctx context.Context, PromptType *models.PromptType) (*models.PromptType, error) {
	stmt := r.mustGetStmt("prompt_type/create_prompt_type.sql")

	out := &models.PromptType{}
	err := stmt.GetContext(ctx, out, map[string]interface{}{
		"name":            PromptType.Name,
		"description":     PromptType.Description,
		"organization_id": PromptType.OrganizationId,
		"config":          PromptType.Config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message type: %w", err)
	}
	return out, nil
}

func (r *Database) UpdatePromptType(ctx context.Context, PromptType *models.PromptType) error {
	stmt := r.mustGetStmt("prompt_type/update_prompt_type.sql")
	_, err := stmt.ExecContext(ctx, map[string]interface{}{
		"description": PromptType.Description,
		"name":        PromptType.Name,
		"config":      PromptType.Config,
	})
	if err != nil {
		return fmt.Errorf("failed to update message type %q: %w", PromptType.Name, err)
	}
	return nil
}

func (r *Database) GetPromptTypeByName(ctx context.Context, name string) (*models.PromptType, error) {
	return getOne[models.PromptType](ctx, r, "prompt_type/query_prompt_type_by_name.sql", map[string]any{
		"name": name,
	})
}
