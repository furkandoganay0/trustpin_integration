package trustpin

import (
	"context"
	"fmt"

	"trustpin_integration/internal/application"
)

type Adapter struct {
	client *Client
}

func NewAdapter(client *Client) *Adapter {
	return &Adapter{client: client}
}

type enrollmentInitRequest struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
}

type deviceActivateRequest struct {
	PairingCode string `json:"pairing_code"`
	PublicKey   string `json:"public_key"`
	Label       string `json:"label,omitempty"`
}

type challengeInitRequest struct {
	TenantID string         `json:"tenant_id"`
	UserID   string         `json:"user_id"`
	DeviceID string         `json:"device_id"`
	Action   string         `json:"action"`
	Context  map[string]any `json:"context,omitempty"`
}

type challengeDecisionRequest struct {
	DeviceID  string         `json:"device_id"`
	Signature string         `json:"signature"`
	Payload   map[string]any `json:"payload"`
	TOTPCode  string         `json:"totp_code,omitempty"`
}

func (a *Adapter) Enroll(ctx context.Context, req application.TrustPinEnrollRequest) (*application.TrustPinEnrollResponse, error) {
	payload := enrollmentInitRequest{
		TenantID: req.TenantID,
		UserID:   req.UserID,
	}
	var out struct {
		EnrollmentID string `json:"enrollment_id"`
		PairingCode  string `json:"pairing_code"`
		ExpiresAt    string `json:"expires_at"`
	}
	if err := a.client.do(ctx, "POST", "/v1/enrollments/init", req.TenantID, payload, &out); err != nil {
		return nil, err
	}
	return &application.TrustPinEnrollResponse{EnrollmentID: out.EnrollmentID, PairingCode: out.PairingCode, ExpiresAt: out.ExpiresAt}, nil
}

func (a *Adapter) Activate(ctx context.Context, req application.TrustPinActivateRequest) (*application.TrustPinActivateResponse, error) {
	payload := deviceActivateRequest{
		PairingCode: req.PairingCode,
		PublicKey:   req.PublicKey,
		Label:       req.Label,
	}
	var out struct {
		DeviceID string `json:"device_id"`
		State    string `json:"state"`
	}
	if err := a.client.do(ctx, "POST", "/v1/devices/activate", req.TenantID, payload, &out); err != nil {
		return nil, err
	}
	return &application.TrustPinActivateResponse{DeviceID: out.DeviceID, State: out.State}, nil
}

func (a *Adapter) CreateChallenge(ctx context.Context, req application.TrustPinChallengeRequest) (*application.TrustPinChallengeResponse, error) {
	payload := challengeInitRequest{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		DeviceID: req.DeviceID,
		Action:   req.Action,
		Context:  req.Context,
	}
	var out struct {
		ChallengeID string `json:"challenge_id"`
		State       string `json:"state"`
		IssuedAt    string `json:"issued_at"`
		ExpiresAt   string `json:"expires_at"`
	}
	if err := a.client.do(ctx, "POST", "/v1/auth/challenges/init", req.TenantID, payload, &out); err != nil {
		return nil, err
	}
	return &application.TrustPinChallengeResponse{ChallengeID: out.ChallengeID, State: out.State, IssuedAt: out.IssuedAt, ExpiresAt: out.ExpiresAt}, nil
}

func (a *Adapter) Approve(ctx context.Context, req application.TrustPinApproveRequest) (*application.TrustPinApproveResponse, error) {
	if req.ChallengeID == "" {
		return nil, fmt.Errorf("missing_challenge_id")
	}
	payload := challengeDecisionRequest{
		DeviceID:  req.DeviceID,
		Signature: req.Signature,
		Payload:   req.Payload,
		TOTPCode:  req.TOTPCode,
	}
	if err := a.client.do(ctx, "POST", "/v1/auth/challenges/"+req.ChallengeID+"/approve", req.TenantID, payload, nil); err != nil {
		return nil, err
	}
	return &application.TrustPinApproveResponse{ChallengeID: req.ChallengeID, Status: "APPROVED"}, nil
}
