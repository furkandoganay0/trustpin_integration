package httptransport

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"trustpin_integration/internal/domain"
)

type loginRequest struct {
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "invalid_json"})
		return
	}
	if req.TenantID == "" || req.Username == "" || req.Password == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_fields"})
		return
	}
	req.TenantID = strings.TrimSpace(req.TenantID)
	ctx := r.Context()
	res, err := s.Auth.Login(ctx, domain.TenantID(req.TenantID), req.Username, req.Password)
	if err != nil {
		writeError(w, &AppError{Status: 401, Code: "invalid_credentials", Message: "invalid_credentials"})
		return
	}
	token, exp, err := s.Tokens.Issue(ctx, req.TenantID, res.UserID)
	if err != nil {
		writeError(w, &AppError{Status: 500, Code: "token_error", Message: "token_issue_failed"})
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{AccessToken: token, ExpiresAt: exp.UTC().Format(time.RFC3339Nano)})
}
