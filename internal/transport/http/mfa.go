package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"trustpin_integration/internal/adapters/trustpin"
	"trustpin_integration/internal/application"
	"trustpin_integration/internal/domain"
	"trustpin_integration/internal/middleware"
)

type cachedResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

func (s *Server) handleIdempotency(w http.ResponseWriter, r *http.Request, tenantID string, fn func() (any, *AppError)) {
	key := r.Header.Get("Idempotency-Key")
	if key != "" && s.MFA.IdemStore != nil {
		if cached, ok, err := s.MFA.IdemStore.Get(r.Context(), domain.TenantID(tenantID), key); err == nil && ok {
			var resp cachedResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(resp.Status)
				_, _ = w.Write(resp.Body)
				return
			}
		}
	}

	payload, appErr := fn()
	if appErr != nil {
		writeError(w, appErr)
		return
	}

	body, _ := json.Marshal(payload)
	writeJSON(w, http.StatusOK, payload)

	if key != "" && s.MFA.IdemStore != nil {
		resp := cachedResponse{Status: http.StatusOK, Body: body}
		if b, err := json.Marshal(resp); err == nil {
			_ = s.MFA.IdemStore.Set(r.Context(), domain.TenantID(tenantID), key, b, 5*time.Minute)
		}
	}
}

type enrollRequest struct {
	DeviceID string `json:"device_id"`
}

type activateRequest struct {
	DeviceID    string `json:"device_id"`
	PairingCode string `json:"pairing_code"`
	PublicKey   string `json:"public_key"`
	Label       string `json:"label"`
}

type challengeRequest struct {
	DeviceID  string         `json:"device_id"`
	Action    string         `json:"action"`
	Context   map[string]any `json:"context"`
}

type approveRequest struct {
	ChallengeID string         `json:"challenge_id"`
	DeviceID    string         `json:"device_id"`
	Signature   string         `json:"signature"`
	Payload     map[string]any `json:"payload"`
	TOTPCode    string         `json:"totp_code"`
}

func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req enrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "invalid_json"})
		return
	}
	if req.DeviceID == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_fields"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	userID, _ := middleware.UserID(r.Context())

	s.handleIdempotency(w, r, tenantID, func() (any, *AppError) {
		res, err := s.MFA.Enroll(r.Context(), domain.TenantID(tenantID), userID, application.TrustPinEnrollRequest{
			TenantID: tenantID,
			UserID:   userID,
			DeviceID: req.DeviceID,
		})
		if err != nil {
			return nil, mapError(err)
		}
		return res, nil
	})
}

func (s *Server) handleActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req activateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "invalid_json"})
		return
	}
	if req.DeviceID == "" || req.PairingCode == "" || req.PublicKey == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_fields"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	userID, _ := middleware.UserID(r.Context())

	s.handleIdempotency(w, r, tenantID, func() (any, *AppError) {
		res, err := s.MFA.Activate(r.Context(), domain.TenantID(tenantID), userID, application.TrustPinActivateRequest{
			TenantID:    tenantID,
			UserID:      userID,
			DeviceID:    req.DeviceID,
			PairingCode: req.PairingCode,
			PublicKey:   req.PublicKey,
			Label:       req.Label,
		})
		if err != nil {
			return nil, mapError(err)
		}
		return res, nil
	})
}

func (s *Server) handleCreateChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req challengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "invalid_json"})
		return
	}
	if req.DeviceID == "" || req.Action == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_fields"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	userID, _ := middleware.UserID(r.Context())

	s.handleIdempotency(w, r, tenantID, func() (any, *AppError) {
		res, err := s.MFA.CreateChallenge(r.Context(), domain.TenantID(tenantID), userID, application.TrustPinChallengeRequest{
			TenantID: tenantID,
			UserID:   userID,
			DeviceID: req.DeviceID,
			Action:   req.Action,
			Context:  req.Context,
		})
		if err != nil {
			return nil, mapError(err)
		}
		return res, nil
	})
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req approveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "invalid_json"})
		return
	}
	if req.ChallengeID == "" || req.DeviceID == "" || req.Signature == "" || req.Payload == nil {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_fields"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	userID, _ := middleware.UserID(r.Context())

	s.handleIdempotency(w, r, tenantID, func() (any, *AppError) {
		res, err := s.MFA.Approve(r.Context(), domain.TenantID(tenantID), userID, application.TrustPinApproveRequest{
			TenantID:    tenantID,
			UserID:      userID,
			DeviceID:    req.DeviceID,
			ChallengeID: req.ChallengeID,
			Signature:   req.Signature,
			Payload:     req.Payload,
			TOTPCode:    req.TOTPCode,
		})
		if err != nil {
			return nil, mapError(err)
		}
		return res, nil
	})
}

func (s *Server) handleGetChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/mfa/challenge/")
	if id == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_id"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	c, err := s.MFA.Challenges.GetByID(r.Context(), domain.TenantID(tenantID), id)
	if err != nil {
		writeError(w, &AppError{Status: 500, Code: "server_error", Message: "failed"})
		return
	}
	if c == nil {
		writeError(w, &AppError{Status: 404, Code: "not_found", Message: "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"challenge_id": c.ID,
		"status":       c.State,
		"updated_at":   c.UpdatedAt,
	})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/mfa/status/")
	if id == "" {
		writeError(w, &AppError{Status: 400, Code: "bad_request", Message: "missing_id"})
		return
	}
	tenantID, _ := middleware.TenantID(r.Context())
	d, err := s.MFA.Devices.GetByID(r.Context(), domain.TenantID(tenantID), id)
	if err != nil {
		writeError(w, &AppError{Status: 500, Code: "server_error", Message: "failed"})
		return
	}
	if d == nil {
		writeError(w, &AppError{Status: 404, Code: "not_found", Message: "not_found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"device_id": d.ID,
		"status":    d.State,
	})
}

func mapError(err error) *AppError {
	var tpErr *trustpin.Error
	if errors.As(err, &tpErr) {
		switch tpErr.Status {
		case 400:
			return &AppError{Status: 400, Code: "bad_request", Message: "invalid_payload", Details: string(tpErr.Body)}
		case 404:
			return &AppError{Status: 404, Code: "not_found", Message: "not_found", Details: string(tpErr.Body)}
		case 409:
			return &AppError{Status: 409, Code: "invalid_state", Message: "invalid_state", Details: string(tpErr.Body)}
		case 410:
			return &AppError{Status: 410, Code: "expired", Message: "expired", Details: string(tpErr.Body)}
		case 412:
			return &AppError{Status: 422, Code: "invalid_signature", Message: "signature_mismatch", Details: string(tpErr.Body)}
		case 429:
			return &AppError{Status: 429, Code: "rate_limited", Message: "rate_limited", Details: string(tpErr.Body)}
		case 503:
			return &AppError{Status: 503, Code: "push_failed", Message: "push_failed", Details: string(tpErr.Body)}
		default:
			return &AppError{Status: 502, Code: "trustpin_error", Message: "upstream_error", Details: string(tpErr.Body)}
		}
	}
	if err.Error() == "nonce_reuse" {
		return &AppError{Status: 409, Code: "nonce_reuse", Message: "nonce_reuse"}
	}
	if err.Error() == "invalid_state" {
		return &AppError{Status: 409, Code: "invalid_state", Message: "invalid_state"}
	}
	if err.Error() == "missing_nonce" {
		return &AppError{Status: 400, Code: "bad_request", Message: "missing_nonce"}
	}
	return &AppError{Status: 500, Code: "server_error", Message: "internal_error"}
}
