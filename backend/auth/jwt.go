package auth

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/shank318/doota/models"
	gcpjwt "github.com/someone1/gcp-jwt-go"
)

// JwtValiditySecond grace period for jwt validity
var JwtValiditySecond = int64(31 * 24 * 60 * 60)

const Issuer = "doota"
const version = 1

func NewTokenFromUser(user *models.User, signingKey interface{}) (*jwt.Token, string, error) {
	jti, err := newUUID()
	if err != nil {
		return nil, "", fmt.Errorf("unable to generate JWT token id(jti): %w", err)
	}

	var method jwt.SigningMethod

	// Default to KMS
	method = gcpjwt.SigningMethodKMSES256

	// Check if the signingKey is a string and contains "dummy"
	if strKey, ok := signingKey.(string); ok && strings.Contains(strKey, "dummy") {
		method = jwt.SigningMethodHS256
		signingKey = []byte(strKey) // 🔥 convert string to []byte
	}

	unsignedToken := newUnsignedJWTTokenFromUser(jti, user, method)

	signedToken, err := unsignedToken.SignedString(signingKey)
	if err != nil {
		return nil, "", fmt.Errorf("unable to sign jwt: %w", err)
	}

	return unsignedToken, signedToken, nil
}

func newUnsignedJWTTokenFromUser(jti string, user *models.User, method jwt.SigningMethod) (unsignedToken *jwt.Token) {
	nowInSeconds := jwt.TimeFunc().Unix()
	expiresAtInSeconds := nowInSeconds + JwtValiditySecond

	claims := &Credentials{
		StandardClaims: jwt.StandardClaims{
			Id:        jti,
			Issuer:    Issuer,
			Subject:   user.ID,
			IssuedAt:  nowInSeconds,
			ExpiresAt: expiresAtInSeconds,
		},
		Version: version,
		UserId:  user.ID,
	}

	return jwt.NewWithClaims(method, claims)
}

func defaultNewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// newUUID generates a random UUID according to RFC 4122
var newUUID = defaultNewUUID
