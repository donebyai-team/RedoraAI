package crypto

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	gcpjwt "github.com/someone1/gcp-jwt-go"
)

type SigningKeyGetter interface {
	GetName() string
	GetKey(context.Context) interface{}
	GetKeyVerificationFunc(context.Context) (jwt.Keyfunc, error)
}

type KMSSigningKeyGetter struct {
	kmsSignAPIKeyPath string
	config            *gcpjwt.KMSConfig
}

func NewKMSSigningKeyGetter(kmsSignKeyPath string) *KMSSigningKeyGetter {
	return &KMSSigningKeyGetter{
		kmsSignAPIKeyPath: kmsSignKeyPath,
		config: &gcpjwt.KMSConfig{
			KeyPath: kmsSignKeyPath,
		},
	}
}

func (g *KMSSigningKeyGetter) GetName() string {
	return "KMS Signing"
}

func (g *KMSSigningKeyGetter) GetKey(ctx context.Context) interface{} {
	return gcpjwt.NewKMSContext(ctx, g.config)
}

func (g *KMSSigningKeyGetter) GetKeyVerificationFunc(ctx context.Context) (jwt.Keyfunc, error) {
	return gcpjwt.KMSVerfiyKeyfunc(context.Background(), &gcpjwt.KMSConfig{
		KeyPath: g.kmsSignAPIKeyPath,
	})
}
