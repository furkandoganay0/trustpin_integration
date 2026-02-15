package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

func NewJWTValidator(publicKeyPEM, issuer, audience string) (*JWTValidator, error) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return nil, err
	}
	return &JWTValidator{publicKey: pub, issuer: issuer, audience: audience}, nil
}

func (v *JWTValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("invalid_signing_method")
			}
			return v.publicKey, nil
		})
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !verifyIssuer(claims, v.issuer) || !verifyAudience(claims, v.audience) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tenantID, _ := claims["tenant_id"].(string)
		userID, _ := claims["sub"].(string)
		ctx := WithTenantID(r.Context(), tenantID)
		ctx = WithUserID(ctx, userID)
		r2 := r.WithContext(ctx)
		// also set tenant/user headers on the request so other middlewares
		// and loggers can read them from headers when context isn't available
		if tenantID != "" {
			r2.Header.Set("X-Tenant-ID", tenantID)
		}
		if userID != "" {
			r2.Header.Set("X-User-ID", userID)
		}
		next.ServeHTTP(w, r2)
	})
}

func verifyIssuer(claims jwt.MapClaims, issuer string) bool {
	v, ok := claims["iss"].(string)
	return ok && v == issuer
}

func verifyAudience(claims jwt.MapClaims, audience string) bool {
	switch aud := claims["aud"].(type) {
	case string:
		return aud == audience
	case []any:
		for _, v := range aud {
			if s, ok := v.(string); ok && s == audience {
				return true
			}
		}
	}
	return false
}
