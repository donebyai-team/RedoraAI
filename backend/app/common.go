package app

import (
	"context"
	"fmt"
	"time"

	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/auth/crypto"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/datastore/psql"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func SetupKMS(ctx context.Context, kmsKeyPath string, logger *zap.Logger) (crypto.SigningKeyGetter, auth.TokenValidationFunc, error) {
	signingKeyGetter := crypto.NewKMSSigningKeyGetter(kmsKeyPath)

	logger.Info("setting up authenticator", zap.Any("kmsKeyPath", kmsKeyPath))
	tokenValidator, err := auth.NewKMSTokenValidator(ctx, kmsKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup kms token validator: %w", err)
	}

	return signingKeyGetter, tokenValidator, nil
}

func SetupMockKMS(ctx context.Context, kmsKeyPath string, logger *zap.Logger) (crypto.SigningKeyGetter, auth.TokenValidationFunc, error) {
	mockKey := "dummy_key"
	signingKeyGetter := crypto.NewMockKeyGetter(mockKey)

	logger.Info("setting up mock authenticator", zap.Any("kmsKeyPath", kmsKeyPath))
	tokenValidator, err := auth.NewMockKMSTokenValidator(ctx, mockKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup kms token validator: %w", err)
	}

	return signingKeyGetter, tokenValidator, nil
}

type PostgresDSNString string

func newDataStore(ctx context.Context, dsnStr PostgresDSNString, logger *zap.Logger, tracer logging.Tracer) (datastore.Repository, error) {
	return SetupDataStore(ctx, string(dsnStr), logger, tracer)
}

func SetupDataStore(ctx context.Context, dsnStr string, logger *zap.Logger, tracer logging.Tracer) (datastore.Repository, error) {
	ctx, cnl := context.WithTimeout(ctx, 3*time.Second)
	defer cnl()

	logger.Info("setting up datastore", zap.String("dsn", dsnStr))
	dbConn, dsn, err := psql.GetConnectionFromDSN(ctx, dsnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to setup pg conn: %w", err)
	}
	dataStore, err := psql.NewRepository(dbConn, dsn, logger, tracer)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore: %w", err)
	}

	if err = dataStore.Setup(); err != nil {
		return nil, fmt.Errorf("failed to setup datastore: %w", err)
	}

	return dataStore, nil
}
