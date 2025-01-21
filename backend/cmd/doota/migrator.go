package main

import (
	"fmt"
	"os"
	"time"

	"github.com/drone/envsubst"
	"github.com/golang-migrate/migrate/source"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/shank318/doota/datastore/psql/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	"github.com/streamingfast/derr"
	"go.uber.org/zap"
)

var MigratorCmd = cli.Command(migratorCmdE,
	"migrator",
	"Starts the migrator",
	cli.ArbitraryArgs(),
	cli.Flags(func(flags *pflag.FlagSet) {
		flags.Uint64("desired-version", 0, "The desired migration version you would like to be at")
		flags.Bool("enable-auto-migration", false, "When enable the mirgator will ALWAYS auto migrate the database")
	}),
)

func migratorCmdE(cmd *cobra.Command, args []string) error {
	desiredVersion := sflags.MustGetUint64(cmd, "desired-version")
	enableAutoMIgration := sflags.MustGetBool(cmd, "enable-auto-migration")

	pgDSN := sflags.MustGetString(cmd, "pg-dsn")
	zlog.Info("starting migrator",
		zap.Uint64("desired_version", desiredVersion),
		zap.Bool("enable_auto_migration", enableAutoMIgration),
		zap.String("pg_dsn", pgDSN),
	)

	m, err := setupMigrator(pgDSN)
	if err != nil {
		return fmt.Errorf("failed to setup migrator: %w", err)
	}
	runMigration(desiredVersion, enableAutoMIgration, m)

	signalHandler := derr.SetupSignalHandler(0 * time.Second)
	zlog.Info("ready, waiting for signal to quit")
	select {
	case <-signalHandler:
		zlog.Info("received termination signal, quitting application")
	}
	zlog.Info("migrator terminated")
	return nil
}

func setupMigrator(pgdsn string) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("new source migrations: %w", err)
	}

	pgdsn, err = envsubst.Eval(pgdsn, os.Getenv)
	if err != nil {
		return nil, fmt.Errorf("failed to expand dsn variables: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("embed", source, pgdsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	m.Log = &Logger{zlog: zlog}
	return m, nil
}

func runMigration(desiredVersion uint64, autoDeploy bool, m *migrate.Migrate) {

	currentVersion, err := getCurrentVersion(m)
	if err != nil {
		zlog.Error("failed to get current", zap.Error(err))
		return
	}

	maxVersion := retrieveLastMigration()
	zlog.Info("database connected",
		zap.Uint64("current_version", currentVersion),
		zap.Uint64("max_version", maxVersion),
	)

	version := uint(0)
	if autoDeploy {
		zlog.Info("auto deploy is enabled, migrating to the max version",
			zap.Uint64("current_version", currentVersion),
			zap.Uint64("max_version", maxVersion),
		)
		version = uint(maxVersion)

	} else {
		zlog.Info("auto deploy is disabled, migrating to the desired version",
			zap.Uint64("current_version", currentVersion),
			zap.Uint64("desired_version", desiredVersion),
		)
		version = uint(desiredVersion)
	}
	t0 := time.Now()
	if err := m.Migrate(version); err != nil {
		if err == migrate.ErrNoChange {
			zlog.Info("nothing to migrate!", zap.Error(err))
			return
		}
		zlog.Error("failed to run migrations", zap.Error(err))
		return
	}
	zlog.Info("migrations ran successfully",
		zap.Int("migrated_version", int(version)),
		zap.Duration("duration", time.Since(t0)),
	)
	return
}

func getCurrentVersion(m *migrate.Migrate) (uint64, error) {
	currentVersion, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			zlog.Info("no migration version found, starting from scratch")
			currentVersion = 0
		} else {
			return 0, fmt.Errorf("failed to get the current version: %w", err)
		}
	}
	if dirty {
		zlog.Error("database is dirty, cannot check migrations")
	}
	return uint64(currentVersion), nil
}
func retrieveLastMigration() uint64 {
	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		panic(fmt.Errorf("failed to read migrations, they are embed so this should not happen: %w", err))
	}

	version := uint(0)
	for _, fi := range entries {
		m, err := source.DefaultParse(fi.Name())
		if err != nil {
			continue // ignore files that we can't parse
		}
		if m.Version > uint(version) {
			version = m.Version
		}
	}
	return uint64(version)
}

type Logger struct {
	zlog *zap.Logger
}

// Printf implements migrate.Logger.
func (*Logger) Printf(format string, v ...interface{}) {
	out := fmt.Sprintf(format, v...)
	zlog.Info(out)
}

// Verbose implements migrate.Logger.
func (*Logger) Verbose() bool {
	return true
}
