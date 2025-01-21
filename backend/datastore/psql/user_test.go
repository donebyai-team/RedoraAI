package psql

import (
	"context"
	"testing"

	"github.com/shank318/doota/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore_CreateUser(t *testing.T) {
	testDB(t, "create_user", func(pgStore *Database) {
		user := testCreateUser(t, pgStore, nil, nil)
		assert.NotNil(t, user)
	})
}

func TestPostgresStore_UpdateUser(t *testing.T) {
	testDB(t, "update_user", func(pgStore *Database) {
		user := testCreateUser(t, pgStore, nil, func(user *models.User) *models.User {
			user.EmailVerified = false
			return user
		})
		assert.NotNil(t, user)
		assert.NotNil(t, user.OrganizationID)

		orgb := testCreateOrganization(t, pgStore, nil)

		user.OrganizationID = orgb.ID
		user.EmailVerified = true
		require.NoError(t, pgStore.UpdateUser(context.Background(), user))

		foundUser, err := pgStore.GetUserById(context.Background(), user.ID)
		require.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, true, foundUser.EmailVerified)
		assert.NotNil(t, foundUser.OrganizationID)
		assert.Equal(t, orgb.ID, foundUser.OrganizationID)
	})
}

func TestPostgresStore_GetUserById(t *testing.T) {
	testDB(t, "get_user_by_id", func(pgStore *Database) {
		user := testCreateUser(t, pgStore, nil, nil)

		foudUser, err := pgStore.GetUserById(context.Background(), user.ID)
		require.NoError(t, err)
		assert.NotNil(t, foudUser)
		assert.Equal(t, foudUser.ID, user.ID)
	})
}
func TestPostgresStore_GetUserByAuth0Id(t *testing.T) {
	testDB(t, "get_user_by_auth_id", func(pgStore *Database) {
		user := testCreateUser(t, pgStore, nil, nil)

		foundUser, err := pgStore.GetUserByAuth0Id(context.Background(), user.Auth0ID)
		require.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.ID, foundUser.ID)

		_, err = pgStore.GetUserByAuth0Id(context.Background(), "notauth0id")
		require.Error(t, err)
	})
}

func testCreateUser(t *testing.T, db *Database, org *models.Organization, f func(user *models.User) *models.User) *models.User {
	if org == nil {
		org = testCreateOrganization(t, db, nil)
	}
	prefix := randomStr(5)
	user := &models.User{
		Auth0ID:        "auth0" + prefix,
		Email:          prefix + "@gmail.com",
		EmailVerified:  true,
		OrganizationID: org.ID,
	}

	if f != nil {
		user = f(user)
	}

	user, err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	return user
}
