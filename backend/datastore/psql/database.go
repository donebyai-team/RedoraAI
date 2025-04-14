package psql

import (
	"context"
	"crypto/aes"
	"fmt"
	"github.com/shank318/doota/models"

	"github.com/jmoiron/sqlx"
	"github.com/shank318/doota/datastore"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var preparedStmts = map[string]string{}

func registerFiles(files []string) {
	for _, file := range files {
		stmt := onDiskStatement(file)
		if _, found := preparedStmts[stmt]; found {
			panic(fmt.Errorf("statement %q already registered", file))
		}
		preparedStmts[file] = stmt
	}
}

var _ datastore.Repository = (*Database)(nil)

type Database struct {
	*sqlx.DB

	encryptionKey [aes.BlockSize]byte
	stmts         map[string]*sqlx.NamedStmt
	zlogger       *zap.Logger
	tracer        logging.Tracer
}

func NewRepository(dbConn *sqlx.DB, dsn *DSN, zlogger *zap.Logger, tracer logging.Tracer) (*Database, error) {
	// At some point in time, when encryption is fully required, this should switch to an error when
	// the encryption key is not set.
	//
	// Old message: "the Postgres DSN requires that the `encryptionKey` must be defined and must be valid but this was not the case"
	var decodedEncryptionKey [aes.BlockSize]byte
	if dsn.EncryptionKey() != "" {
		var err error
		decodedEncryptionKey, err = decodeEncryptionKey(dsn.EncryptionKey())
		if err != nil {
			return nil, fmt.Errorf("could not decode encryption key: %w", err)
		}
	}

	return &Database{
		DB:            dbConn,
		encryptionKey: decodedEncryptionKey,
		stmts:         make(map[string]*sqlx.NamedStmt),
		zlogger:       zlogger,
		tracer:        tracer,
	}, nil
}

func (r *Database) Setup() error {
	if err := r.setupPreparedStmt(preparedStmts); err != nil {
		return fmt.Errorf("failed to register prepared stmt: %w", err)
	}
	return nil
}

func (r *Database) setupPreparedStmt(stmts map[string]string) error {
	for k, s := range stmts {
		if _, found := r.stmts[k]; found {
			return fmt.Errorf("statement key %q already in use", k)
		}

		ps, err := r.PrepareNamed(s)
		if err != nil {
			return fmt.Errorf("failed to register prepared statement with key %q: %w", k, err)
		}

		r.stmts[k] = ps
	}

	return nil
}

func (r *Database) mustGetStmt(key string) *sqlx.NamedStmt {
	v, ok := r.stmts[key]
	if !ok {
		panic(fmt.Errorf("unable to find prepared stmt %q", key))
	}
	return v
}

func (r *Database) mustGetTxStmt(ctx context.Context, name string, tx *sqlx.Tx) *sqlx.NamedStmt {
	return tx.NamedStmtContext(ctx, r.mustGetStmt(name))
}
