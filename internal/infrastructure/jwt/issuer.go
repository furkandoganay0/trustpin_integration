package jwt

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	ttl        time.Duration
}

func NewIssuer(privateKeyPEM, issuer, audience string, ttl time.Duration) (*Issuer, error) {
	pk, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return nil, err
	}
	return &Issuer{privateKey: pk, issuer: issuer, audience: audience, ttl: ttl}, nil
}

func (i *Issuer) Issue(ctx context.Context, tenantID, userID string) (string, time.Time, error) {
	exp := time.Now().Add(i.ttl)
	claims := jwt.MapClaims{
		"sub":       userID,
		"tenant_id": tenantID,
		"iss":       i.issuer,
		"aud":       i.audience,
		"exp":       exp.Unix(),
		"iat":       time.Now().Unix(),
		"jti":       jwt.NewNumericDate(time.Now()).String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}
