package psql

import (
	"context"
	"testing"

	"github.com/google/uuid"
	models "github.com/shank318/doota/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore_UpsertIntegration(t *testing.T) {
	testDB(t, "upsert_integration", func(pgStore *Database) {
		integration := testCreateIntegration(t, pgStore, nil, nil)
		assert.NotNil(t, integration)

		foundIntegration, err := pgStore.GetIntegrationByOrgAndType(context.Background(), integration.OrganizationID, integration.Type)
		require.NoError(t, err)
		assertEqualIntegration(t, integration, foundIntegration)

		integration.EncryptedConfig = "{}"
		integration, err = pgStore.UpsertIntegration(context.Background(), integration)
		require.NoError(t, err)

		foundIntegration, err = pgStore.GetIntegrationByOrgAndType(context.Background(), integration.OrganizationID, integration.Type)
		require.NoError(t, err)
		assertEqualIntegration(t, integration, foundIntegration)
	})
}

func TestPostgresStore_GetIntegrationByOrgAndType(t *testing.T) {
	testDB(t, "get_integration_by_org_type", func(pgStore *Database) {
		orga := testCreateOrganization(t, pgStore, nil)
		orgb := testCreateOrganization(t, pgStore, nil)
		inta := testCreateIntegration(t, pgStore, orga, nil)
		testCreateIntegration(t, pgStore, orgb, nil)

		foundA, err := pgStore.GetIntegrationByOrgAndType(context.Background(), orga.ID, inta.Type)
		require.NoError(t, err)
		assertEqualIntegration(t, inta, foundA)

		_, err = pgStore.GetIntegrationByOrgAndType(context.Background(), uuid.NewString(), models.IntegrationTypeMICROSOFT)
		require.Error(t, err)
	})
}

func TestPostgresStore_GetIntegrationById(t *testing.T) {
	testDB(t, "get_integration_by_id", func(pgStore *Database) {
		inta := testCreateIntegration(t, pgStore, nil, nil)

		found, err := pgStore.GetIntegrationById(context.Background(), inta.ID)
		require.NoError(t, err)
		assertEqualIntegration(t, inta, found)

		_, err = pgStore.GetIntegrationById(context.Background(), uuid.NewString())
		require.Error(t, err)
	})
}

func testCreateIntegration(t *testing.T, db *Database, org *models.Organization, f func(msg *models.Integration) *models.Integration) *models.Integration {
	if org == nil {
		org = testCreateOrganization(t, db, nil)
	}

	integration := &models.Integration{
		OrganizationID: org.ID,
		State:          models.IntegrationStateACTIVE,
	}
	integration = models.SetIntegrationType(integration, models.IntegrationTypeMICROSOFT, models.NewMicrosoftConfig("11"))

	if f != nil {
		integration = f(integration)
	}

	integration, err := db.UpsertIntegration(context.Background(), integration)
	require.NoError(t, err)
	return integration
}

func assertEqualIntegration(t *testing.T, expected, actual *models.Integration) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.OrganizationID, actual.OrganizationID)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.State, actual.State)
	assert.Equal(t, expected.EncryptedConfig, actual.EncryptedConfig)
}
