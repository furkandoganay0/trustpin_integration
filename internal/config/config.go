package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string
	Port            string
	DBDSN           string
	RedisAddr       string
	TrustPinBaseURL string
	TrustPinAPIKey  string
	JWTIssuer       string
	JWTAudience     string
	JWTPublicKeyPEM string
	JWTPrivateKeyPEM string
	HTTPTimeout     time.Duration
	RetryMax        int
	RetryBackoff    time.Duration
}

func Load() Config {
	return Config{
		Env:             getenv("APP_ENV", "dev"),
		Port:            getenv("PORT", "8080"),
		DBDSN:           getenv("DB_DSN", ""),
		RedisAddr:       getenv("REDIS_ADDR", ""),
		TrustPinBaseURL: getenv("TRUSTPIN_BASE_URL", "http://trustpin.kaizen3.online"),
		TrustPinAPIKey:  getenv("TRUSTPIN_API_KEY", ""),
		JWTIssuer:       getenv("JWT_ISSUER", "trustpin"),
		JWTAudience:     getenv("JWT_AUDIENCE", "mobile"),
		JWTPublicKeyPEM: normalizePEM(getenv("JWT_PUBLIC_KEY", "")),
		JWTPrivateKeyPEM: normalizePEM(getenv("JWT_PRIVATE_KEY", "")),
		HTTPTimeout:     getDuration("HTTP_TIMEOUT", 5*time.Second),
		RetryMax:        getInt("RETRY_MAX", 2),
		RetryBackoff:    getDuration("RETRY_BACKOFF", 200*time.Millisecond),
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return parsed
}

func normalizePEM(v string) string {
	if v == "" {
		return v
	}
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "\"")
	v = strings.TrimSuffix(v, "\"")
	v = strings.ReplaceAll(v, "\\n", "\n")
	return strings.TrimSpace(v)
}
