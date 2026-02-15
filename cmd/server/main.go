package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trustpin_integration/internal/adapters/trustpin"
	"trustpin_integration/internal/application"
	"trustpin_integration/internal/config"
	"trustpin_integration/internal/infrastructure/idempotency"
	"trustpin_integration/internal/infrastructure/jwt"
	"trustpin_integration/internal/infrastructure/memory"
	"trustpin_integration/internal/infrastructure/postgres"
	"trustpin_integration/internal/infrastructure/redis"
	"trustpin_integration/internal/domain"
	"trustpin_integration/internal/middleware"
	"trustpin_integration/internal/transport/http"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	jwtValidator, err := middleware.NewJWTValidator(cfg.JWTPublicKeyPEM, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		logger.Error("jwt_validator", "error", err)
		os.Exit(1)
	}
	issuer, err := jwt.NewIssuer(cfg.JWTPrivateKeyPEM, cfg.JWTIssuer, cfg.JWTAudience, 15*time.Minute)
	if err != nil {
		logger.Error("jwt_issuer", "error", err)
		os.Exit(1)
	}

	trustpinClient := trustpin.NewClient(cfg.TrustPinBaseURL, cfg.TrustPinAPIKey, cfg.HTTPTimeout, trustpin.RetryConfig{Max: cfg.RetryMax, Backoff: cfg.RetryBackoff})
	trustpinAdapter := trustpin.NewAdapter(trustpinClient)

	var (
		users     application.UserRepository
		sessions  application.SessionRepository
		devices   application.DeviceRepository
		challenges application.ChallengeRepository
		nonceStore application.NonceStore
		idemStore  application.IdempotencyStore
	)

	if cfg.DBDSN == "" || cfg.RedisAddr == "" {
		memUsers := memory.NewUserRepo()
		memUsers.Seed(&domain.User{ID: "demo-user", TenantID: "demo-tenant", Username: "demo", Status: "ACTIVE", CreatedAt: time.Now()})
		users = memUsers
		sessions = memory.NewSessionRepo()
		devices = memory.NewDeviceRepo()
		challenges = memory.NewChallengeRepo()
		nonceStore = memory.NewNonceStore()
		idemStore = memory.NewIdempotencyStore()
	} else {
		users = &postgres.UserRepo{}
		sessions = &postgres.SessionRepo{}
		devices = &postgres.DeviceRepo{}
		challenges = &postgres.ChallengeRepo{}
		nonceStore = &redis.NonceStore{}
		idemStore = &idempotency.Store{}
	}

	authSvc := &application.AuthService{Users: users, Sessions: sessions}
	mfaSvc := &application.MFAService{Devices: devices, Challenges: challenges, NonceStore: nonceStore, IdemStore: idemStore, TrustPin: trustpinAdapter}

	server := &httptransport.Server{Auth: authSvc, MFA: mfaSvc, JWT: jwtValidator, Log: logger, Tokens: issuer}

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server_start", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server_error", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	logger.Info("server_shutdown")
}
