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
	// support reading JWT keys either directly from env variables or from
	// separate files. two environment variables are recognized for each key:
	//   JWT_PUBLIC_KEY, JWT_PRIVATE_KEY          // raw PEM string
	//   JWT_PUBLIC_KEY_FILE, JWT_PRIVATE_KEY_FILE // path to file containing PEM
	//
	// If both a file and a raw value are provided, the file takes precedence.
	// As a convenience when running locally we also look for the default
	// files `config/jwt_public.pem` and `config/jwt_private.pem` if nothing
	// else is set.

	pubPem := getenv("JWT_PUBLIC_KEY", "")
	if path := os.Getenv("JWT_PUBLIC_KEY_FILE"); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			pubPem = string(data)
		}
	}
	if pubPem == "" {
		// try default file relative to project root
		if data, err := os.ReadFile("config/jwt_public.pem"); err == nil {
			pubPem = string(data)
		}
	}

	privPem := getenv("JWT_PRIVATE_KEY", "")
	if path := os.Getenv("JWT_PRIVATE_KEY_FILE"); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			privPem = string(data)
		}
	}
	if privPem == "" {
		if data, err := os.ReadFile("config/jwt_private.pem"); err == nil {
			privPem = string(data)
		}
	}

	return Config{
		Env:             getenv("APP_ENV", "dev"),
		Port:            getenv("PORT", "8080"),
		DBDSN:           getenv("DB_DSN", ""),
		RedisAddr:       getenv("REDIS_ADDR", ""),
		TrustPinBaseURL: getenv("TRUSTPIN_BASE_URL", "http://trustpin.kaizen3.online"),
		TrustPinAPIKey:  getenv("TRUSTPIN_API_KEY", ""),
		JWTIssuer:       getenv("JWT_ISSUER", "trustpin"),
		JWTAudience:     getenv("JWT_AUDIENCE", "mobile"),
		JWTPublicKeyPEM: normalizePEM(pubPem),
		JWTPrivateKeyPEM: normalizePEM(privPem),
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
