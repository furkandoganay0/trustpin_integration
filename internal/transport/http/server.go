package httptransport

import (
	"log/slog"
	"net/http"
	"time"

	"trustpin_integration/internal/application"
	"trustpin_integration/internal/middleware"
)

type Server struct {
	Auth *application.AuthService
	MFA  *application.MFAService
	JWT  *middleware.JWTValidator
	Log  *slog.Logger
	Tokens TokenIssuer
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/swagger/openapi.yaml", s.handleOpenAPI)
	mux.HandleFunc("/swagger/", s.handleSwaggerUI)
	mux.HandleFunc("/api/auth/login", s.handleLogin)

	secured := http.NewServeMux()
	secured.HandleFunc("/api/mfa/enroll", s.handleEnroll)
	secured.HandleFunc("/api/mfa/activate", s.handleActivate)
	secured.HandleFunc("/api/mfa/challenge", s.handleCreateChallenge)
	secured.HandleFunc("/api/mfa/approve", s.handleApprove)
	secured.HandleFunc("/api/mfa/challenge/", s.handleGetChallenge)
	secured.HandleFunc("/api/mfa/status/", s.handleGetStatus)

	var handler http.Handler = mux
	securedHandler := http.Handler(secured)
	securedHandler = middleware.EnforceTenant(securedHandler)
	securedHandler = s.JWT.Middleware(securedHandler)

	mux.Handle("/api/mfa/", securedHandler)

	handler = middleware.RequestID(handler)
	handler = middleware.Logging(s.Log)(handler)

	return handler
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
