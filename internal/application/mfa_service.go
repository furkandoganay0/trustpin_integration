package application

import (
	"context"
	"errors"
	"time"

	"trustpin_integration/internal/domain"
)

type MFAService struct {
	Devices     DeviceRepository
	Challenges  ChallengeRepository
	NonceStore  NonceStore
	IdemStore   IdempotencyStore
	TrustPin    TrustPinAdapter
}

func (s *MFAService) Enroll(ctx context.Context, tenantID domain.TenantID, userID string, req TrustPinEnrollRequest) (*TrustPinEnrollResponse, error) {
	d := &domain.MFADevice{
		ID:         req.DeviceID,
		TenantID:   tenantID,
		UserID:     userID,
		DeviceName: "",
		PublicKey:  "",
		State:      "PENDING",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.Devices.Create(ctx, d); err != nil {
		return nil, err
	}

	res, err := s.TrustPin.Enroll(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.Devices.UpdateState(ctx, tenantID, d.ID, "PAIRING_PENDING"); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *MFAService) Activate(ctx context.Context, tenantID domain.TenantID, userID string, req TrustPinActivateRequest) (*TrustPinActivateResponse, error) {
	d, err := s.Devices.GetByID(ctx, tenantID, req.DeviceID)
	if err != nil {
		return nil, err
	}
	if d == nil || d.State != "PAIRING_PENDING" {
		return nil, errors.New("invalid_state")
	}

	res, err := s.TrustPin.Activate(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.Devices.UpdateState(ctx, tenantID, d.ID, "ACTIVE"); err != nil {
		return nil, err
	}
	if res.DeviceID != "" && res.DeviceID != req.DeviceID {
		alias := &domain.MFADevice{
			ID:         res.DeviceID,
			TenantID:   tenantID,
			UserID:     userID,
			DeviceName: d.DeviceName,
			PublicKey:  d.PublicKey,
			State:      "ACTIVE",
			CreatedAt:  d.CreatedAt,
			UpdatedAt:  time.Now(),
		}
		if err := s.Devices.Create(ctx, alias); err != nil && err.Error() != "device_exists" {
			return nil, err
		}
	}
	return res, nil
}

func (s *MFAService) CreateChallenge(ctx context.Context, tenantID domain.TenantID, userID string, req TrustPinChallengeRequest) (*TrustPinChallengeResponse, error) {
	d, err := s.Devices.GetByID(ctx, tenantID, req.DeviceID)
	if err != nil {
		return nil, err
	}
	if d == nil || d.State != "ACTIVE" {
		return nil, errors.New("invalid_state")
	}

	res, err := s.TrustPin.CreateChallenge(ctx, req)
	if err != nil {
		return nil, err
	}

	c := &domain.MFAChallenge{
		ID:                 res.ChallengeID,
		TenantID:           tenantID,
		UserID:             userID,
		DeviceID:           req.DeviceID,
		Action:             req.Action,
		State:              res.State,
		IssuedAt:           time.Now(),
		ExpiresAt:          time.Now().Add(2 * time.Minute),
		UpdatedAt:          time.Now(),
		TrustPinChallengeID: res.ChallengeID,
	}
	if err := s.Challenges.Create(ctx, c); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *MFAService) Approve(ctx context.Context, tenantID domain.TenantID, userID string, req TrustPinApproveRequest) (*TrustPinApproveResponse, error) {
	c, err := s.Challenges.GetByID(ctx, tenantID, req.ChallengeID)
	if err != nil {
		return nil, err
	}
	if c == nil || c.State != "PUSH_SENT" {
		return nil, errors.New("invalid_state")
	}
	if nonce, ok := extractNonce(req.Payload); ok {
		if okSet, err := s.NonceStore.CheckAndSet(ctx, tenantID, nonce, time.Minute*5); err != nil || !okSet {
			return nil, errors.New("nonce_reuse")
		}
	} else {
		return nil, errors.New("missing_nonce")
	}

	res, err := s.TrustPin.Approve(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.Challenges.UpdateState(ctx, tenantID, req.ChallengeID, res.Status); err != nil {
		return nil, err
	}
	return res, nil
}

func extractNonce(payload map[string]any) (string, bool) {
	if payload == nil {
		return "", false
	}
	if v, ok := payload["nonce"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, true
		}
	}
	return "", false
}
