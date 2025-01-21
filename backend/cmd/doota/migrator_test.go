package main

import (
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
)

func Test_retrieveLastMigration(t *testing.T) {
	assert.GreaterOrEqual(t, retrieveLastMigration(), uint64(15))
}
