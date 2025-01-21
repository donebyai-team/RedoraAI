package auth

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	gcpjwt "github.com/someone1/gcp-jwt-go"
)

type TokenValidationFunc = func(token string) (*Credentials, error)

func NewKMSTokenValidator(ctx context.Context, kmsKeyPath string) (TokenValidationFunc, error) {
	kmsVerificationKeyFunc, err := gcpjwt.KMSVerfiyKeyfunc(ctx, &gcpjwt.KMSConfig{
		KeyPath: kmsKeyPath,
	})
	if err != nil {
		return nil, fmt.Errorf("new kms verify func: %w", err)
	}

	tokenValidator := func(token string) (*Credentials, error) {
		credentials := &Credentials{}
		parsedToken, err := jwt.ParseWithClaims(token, credentials, kmsVerificationKeyFunc)
		if err != nil {
			return nil, fmt.Errorf("unable to parse JWT token: %w", err)
		}

		expectedSigningAlgorithm := gcpjwt.SigningMethodKMSES256.Alg()
		actualSigningAlgorithm := parsedToken.Header["alg"]

		if expectedSigningAlgorithm != actualSigningAlgorithm {
			return nil, fmt.Errorf("invalid JWT token: expected signing method %s, got %s", expectedSigningAlgorithm, actualSigningAlgorithm)
		}

		if !parsedToken.Valid {
			return nil, fmt.Errorf("invalid JWT token: invalid signature")
		}
		return credentials, nil
	}
	return tokenValidator, nil
}
