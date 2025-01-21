package psql

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/drone/envsubst"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

func GetConnectionFromDSN(ctx context.Context, dsnStr string) (*sqlx.DB, *DSN, error) {
	dsn, err := parseDSN(dsnStr, os.Getenv)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse dsn string: %w", err)
	}

	db, err := initDBConnection(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize db: %w", err)
	}

	return db, dsn, nil
}

// initDBConnection instantiates a new database connection
func initDBConnection(ctx context.Context, dsn *DSN) (*sqlx.DB, error) {
	ctx, cancelCallback := context.WithTimeout(ctx, 5*time.Second)
	defer cancelCallback()

	db, err := otelsqlx.ConnectContext(ctx, "postgres", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if dsn.MaxConnections() > 0 {
		db.SetMaxOpenConns(int(dsn.MaxConnections()))
	}

	return db, nil
}

func parseDSN(dsnStr string, mapper func(s string) string) (*DSN, error) {
	expanded, err := envsubst.Eval(dsnStr, mapper)
	if err != nil {
		return nil, fmt.Errorf("failed to expand variables: %w", err)
	}

	dsnURL, err := url.Parse(expanded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	isValidScheme := dsnURL.Scheme == "postgresql" || dsnURL.Scheme == "postgres"
	if !isValidScheme {
		return nil, fmt.Errorf(`invalid scheme %q, should be "postgresql" or "postgres"`, dsnURL.Scheme)
	}

	host := dsnURL.Hostname()

	port := int64(5432)
	if strings.Contains(dsnURL.Host, ":") {
		port, _ = strconv.ParseInt(dsnURL.Port(), 10, 32)
	}

	username := dsnURL.User.Username()
	password, _ := dsnURL.User.Password()
	database := strings.TrimPrefix(dsnURL.EscapedPath(), "/")

	query := dsnURL.Query()
	keys := make([]string, 0, len(query))
	for key := range dsnURL.Query() {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	maxConnections := uint64(0)
	encryptionKey := ""
	options := make([]string, len(query))

	for i, key := range keys {
		if key == "maxConnections" {
			maxConn, err := strconv.ParseUint(query[key][0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse max number of connections: %w", err)
			}
			maxConnections = maxConn
			dsnStr = strings.Replace(dsnStr, fmt.Sprintf("maxConnections=%s", query[key][0]), "", 1)
			continue
		}

		if key == "encryptionKey" {
			encryptionKey = query[key][0]
			dsnStr = strings.Replace(dsnStr, fmt.Sprintf("encryptionKey=%s", query[key][0]), "", 1)
			continue
		}

		options[i] = fmt.Sprintf("%s=%s", key, strings.Join(query[key], ","))
	}

	return &DSN{dsnStr, host, port, database, username, password, maxConnections, encryptionKey, options}, nil
}

type DSN struct {
	original string

	host           string
	port           int64
	database       string
	username       string
	password       string
	maxConnections uint64
	encryptionKey  string
	options        []string
}

// String returns the full dsn connection string
func (c *DSN) String() string {
	out := fmt.Sprintf("host=%s port=%d user=%s dbname=%s %s", c.host, c.port, c.username, c.database, strings.Join(c.options, " "))
	if c.password != "" {
		out = out + " password=" + c.password
	}
	return out
}

// String returns the original dsn connection string configured
func (c *DSN) Original() string {
	return c.original
}

// MaxConnections returns the max number of connections
func (c *DSN) MaxConnections() uint64 {
	return c.maxConnections
}

// EncryptionKey returns the encryption/decryption key that should be used
// to encrypt/decrypt data that needs secure storage in the database.
func (c *DSN) EncryptionKey() string {
	return c.encryptionKey
}
