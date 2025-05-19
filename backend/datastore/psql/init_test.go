package psql

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hashicorp/go-multierror"
	"github.com/lib/pq"
	gobackend "github.com/shank318/doota"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/rand"
)

var (
	currentTime = time.Now().UTC()
)

var str = utils.Ptr[string]

const DB_FORCE_CLEAN = true

var currentPGDb *Database
var pgDriver database.Driver
var setupOnce sync.Once

var testLogger, testTracer = logging.PackageLogger("test", "github.com/getfreightstream/freightstream/datastore/psql_test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zapcore.WarnLevel))
}

func testDB(t *testing.T, testName string, testFunc func(pgStore *Database)) {
	dsn := os.Getenv("TEST_PGDSN")
	if dsn == "" {
		t.Skipf("skipping test: %s, unable to run DB test without a DSN set TEST_PGDSN", testName)
	}

	migrationPath := fmt.Sprintf("file://%s/backend/datastore/psql/migrations", gobackend.RootDir())
	pgStore, dbDriver := getTestDB(t, dsn)
	m, err := migrate.NewWithDatabaseInstance(migrationPath, "freightstream_test", dbDriver)
	require.NoError(t, err)

	version, dirty, err := m.Version()
	// TODO: we know we have 2 migration to run, ideally this number is retrieved from the file
	// this is a bit hacky clean up later
	shouldClean := (err != nil) || version != uint(2) || dirty || DB_FORCE_CLEAN

	if shouldClean {
		err = customDrop(t, dsn)
		require.NoError(t, err, "unable to custom drop database")

		m, err = migrate.New(migrationPath, dsn)
		require.NoError(t, err, "unable to create new migration")

		err = m.Up()
		require.NoError(t, err, "unable to run migration")
	}

	require.NoError(t, pgStore.Setup())

	testFunc(pgStore)
}

func customDrop(t *testing.T, dsnStr string) error {
	t.Helper()

	db, _, err := GetConnectionFromDSN(context.Background(), dsnStr)
	require.NoError(t, err, "unable to get connection for drop")

	// select all tables in current schema
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema=(SELECT current_schema()) AND table_type='BASE TABLE'`
	tables, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}
	defer func() {
		if errClose := tables.Close(); errClose != nil {
			err = multierror.Append(err, errClose)
		}
	}()

	// delete one table after another
	tableNames := make([]string, 0)
	for tables.Next() {
		var tableName string
		if err := tables.Scan(&tableName); err != nil {
			return err
		}
		if len(tableName) > 0 {
			tableNames = append(tableNames, tableName)
		}
	}
	if err := tables.Err(); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	if len(tableNames) > 0 {
		// delete one by one ...
		for _, t := range tableNames {
			if t == "spatial_ref_sys" {
				continue
			}

			query = `DROP TABLE IF EXISTS ` + pq.QuoteIdentifier(t) + ` CASCADE`
			if _, err := db.ExecContext(context.Background(), query); err != nil {
				return &database.Error{OrigErr: err, Query: []byte(query)}
			}
		}
	}

	return nil
}

func getTestDB(t *testing.T, dsnStr string) (*Database, database.Driver) {
	if currentPGDb != nil {
		return currentPGDb, pgDriver
	}

	db, dsn, err := GetConnectionFromDSN(context.Background(), dsnStr)
	require.NoError(t, err)

	db.SetMaxOpenConns(1000)
	currentPGDb, err := NewRepository(db, dsn, testLogger, testTracer)
	require.NoError(t, err)

	dbDriver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	require.NoError(t, err)
	pgDriver = dbDriver

	return currentPGDb, pgDriver
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func assertEqualPtr[T comparable](t *testing.T, expected, actual *T) {
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	require.NotNil(t, actual)
	require.Equal(t, *expected, *actual)
}
