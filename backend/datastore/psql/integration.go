package psql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/shank318/doota/models"
)

func init() {
	registerFiles([]string{
		"integration/upsert_integration.sql",
		"integration/query_integration_by_org_and_type.sql",
		"integration/query_integration_by_id.sql",
		"integration/query_integration_by_org.sql",
		"integration/query_integration_by_reference.sql",
	})
}

func (r *Database) UpsertIntegration(ctx context.Context, integration *models.Integration) (*models.Integration, error) {
	stmt := r.mustGetStmt("integration/upsert_integration.sql")
	out := models.Integration{}

	if integration.ID == "" {
		integration.ID = uuid.NewString()
	}

	if integration.EncryptedConfig != "" {
		encryptedString, err := r.encryptMessage(integration.EncryptedConfig)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to encrypt [%s] config", integration.Type))
		}
		integration.EncryptedConfig = encryptedString
	}

	err := stmt.GetContext(ctx, &out, map[string]interface{}{
		"id":                integration.ID,
		"organization_id":   integration.OrganizationID,
		"reference_id":      integration.ReferenceID,
		"type":              integration.Type,
		"encrypted_config":  integration.EncryptedConfig,
		"plain_text_config": integration.PlainTextConfig,
		"state":             integration.State,
		"metadata":          integration.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}
	return &out, nil
}

func (r *Database) GetIntegrationByOrgAndType(ctx context.Context, organizationId string, integrationType models.IntegrationType) ([]*models.Integration, error) {
	integrations, err := getMany[models.Integration](ctx, r, "integration/query_integration_by_org_and_type.sql", map[string]any{
		"organization_id": organizationId,
		"type":            integrationType,
	})

	if err != nil {
		return nil, err
	}

	for _, integration := range integrations {
		integration.EncryptedConfig, err = r.decryptMessage(integration.EncryptedConfig)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decrypt [%s] config", integration.Type))
		}

	}

	return integrations, nil
}

func (r *Database) GetIntegrationsByOrgID(ctx context.Context, orgID string) ([]*models.Integration, error) {
	integrations, err := getMany[models.Integration](ctx, r, "integration/query_integration_by_org.sql", map[string]any{
		"organization_id": orgID,
	})

	if err != nil {
		return nil, err
	}

	for _, integration := range integrations {
		integration.EncryptedConfig, err = r.decryptMessage(integration.EncryptedConfig)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decrypt [%s] config", integration.Type))
		}

	}

	return integrations, nil
}

func (r *Database) GetIntegrationById(ctx context.Context, id string) (*models.Integration, error) {
	integration, err := getOne[models.Integration](ctx, r, "integration/query_integration_by_id.sql", map[string]any{
		"id": id,
	})

	if err != nil {
		return nil, err
	}

	integration.EncryptedConfig, err = r.decryptMessage(integration.EncryptedConfig)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decrypt [%s] config", integration.Type))
	}

	return integration, nil
}

func (r *Database) GetIntegrationsByReferenceId(ctx context.Context, referenceId string) ([]*models.Integration, error) {
	integrations, err := getMany[models.Integration](ctx, r, "integration/query_integration_by_reference.sql", map[string]any{
		"reference_id": referenceId,
	})

	if err != nil {
		return nil, err
	}

	for _, integration := range integrations {
		integration.EncryptedConfig, err = r.decryptMessage(integration.EncryptedConfig)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to decrypt [%s] config", integration.Type))
		}

	}
	return integrations, nil
}
