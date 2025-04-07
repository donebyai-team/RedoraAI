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

type MockKeyGetter struct {
	key string
}

func NewMockKeyGetter(key string) *MockKeyGetter {
	return &MockKeyGetter{key: key}
}

func (m MockKeyGetter) GetName() string {
	return "Mock KMS Signing"
}

func (m MockKeyGetter) GetKey(ctx context.Context) interface{} {
	return m.key
}

func (m MockKeyGetter) GetKeyVerificationFunc(ctx context.Context) (jwt.Keyfunc, error) {
	//TODO implement me
	panic("implement me")
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
