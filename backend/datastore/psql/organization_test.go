package psql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	models "github.com/shank318/doota/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore_CreateOrganization(t *testing.T) {
	testDB(t, "create_organization", func(pgStore *Database) {
		org := testCreateOrganization(t, pgStore, nil)
		assert.NotNil(t, org)
	})
}

func TestPostgresStore_GetOrganizationById(t *testing.T) {
	testDB(t, "get_organization_by_id", func(pgStore *Database) {
		org := testCreateOrganization(t, pgStore, nil)

		foundOrg, err := pgStore.GetOrganizationById(context.Background(), org.ID)
		require.NoError(t, err)
		assert.NotNil(t, foundOrg)
		assertEqualOrganization(t, org, foundOrg)

		_, err = pgStore.GetOrganizationById(context.Background(), uuid.NewString())
		require.Error(t, err)
	})
}
func TestPostgresStore_GetOrganizations(t *testing.T) {
	testDB(t, "get_organizations", func(pgStore *Database) {
		orgA := testCreateOrganization(t, pgStore, func(org *models.Organization) *models.Organization {
			org.Name = "Corp A"
			return org
		})
		orgB := testCreateOrganization(t, pgStore, func(org *models.Organization) *models.Organization {
			org.Name = "Corp B"
			return org
		})

		foundOrgs, err := pgStore.GetOrganizations(context.Background())
		require.NoError(t, err)
		require.Equal(t, 2, len(foundOrgs))
		assertEqualOrganization(t, orgA, foundOrgs[0])
		assertEqualOrganization(t, orgB, foundOrgs[1])
	})
}
func testCreateOrganization(t *testing.T, db *Database, f func(org *models.Organization) *models.Organization) *models.Organization {
	org := &models.Organization{
		Name:         randomStr(10),
		FeatureFlags: models.OrganizationFeatureFlags{},
		CreatedAt:    time.Time{},
		UpdatedAt:    nil,
	}

	if f != nil {
		org = f(org)
	}

	ba, err := db.CreateOrganization(context.Background(), org)
	require.NoError(t, err)
	return ba
}

func assertEqualOrganization(t *testing.T, expected, actual *models.Organization) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.FeatureFlags, actual.FeatureFlags)
}
