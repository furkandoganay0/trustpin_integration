package application

import (
	"context"
	"time"

	"trustpin_integration/internal/domain"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, tenantID domain.TenantID, username string) (*domain.User, error)
	GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	RevokeByJWTID(ctx context.Context, tenantID domain.TenantID, jwtID string, revokedAt time.Time) error
}

type DeviceRepository interface {
	Create(ctx context.Context, d *domain.MFADevice) error
	GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFADevice, error)
	UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error
}

type ChallengeRepository interface {
	Create(ctx context.Context, c *domain.MFAChallenge) error
	GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFAChallenge, error)
	UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error
}

type NonceStore interface {
	CheckAndSet(ctx context.Context, tenantID domain.TenantID, nonce string, ttl time.Duration) (bool, error)
}

type IdempotencyStore interface {
	Get(ctx context.Context, tenantID domain.TenantID, key string) ([]byte, bool, error)
	Set(ctx context.Context, tenantID domain.TenantID, key string, value []byte, ttl time.Duration) error
}

type TrustPinAdapter interface {
	Enroll(ctx context.Context, req TrustPinEnrollRequest) (*TrustPinEnrollResponse, error)
	Activate(ctx context.Context, req TrustPinActivateRequest) (*TrustPinActivateResponse, error)
	CreateChallenge(ctx context.Context, req TrustPinChallengeRequest) (*TrustPinChallengeResponse, error)
	Approve(ctx context.Context, req TrustPinApproveRequest) (*TrustPinApproveResponse, error)
}

// Request/response DTOs abstracted from adapter.
type TrustPinEnrollRequest struct {
	TenantID   string
	UserID     string
	DeviceID   string
}

type TrustPinEnrollResponse struct {
	EnrollmentID string `json:"enrollment_id"`
	PairingCode  string `json:"pairing_code"`
	ExpiresAt    string `json:"expires_at"`
}

type TrustPinActivateRequest struct {
	TenantID    string
	UserID      string
	DeviceID    string
	PairingCode string
	PublicKey   string
	Label       string
}

type TrustPinActivateResponse struct {
	DeviceID string `json:"device_id"`
	State    string `json:"state"`
}

type TrustPinChallengeRequest struct {
	TenantID   string
	UserID     string
	DeviceID   string
	Action     string
	Context    map[string]any
}

type TrustPinChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	State       string `json:"state"`
	IssuedAt    string `json:"issued_at"`
	ExpiresAt   string `json:"expires_at"`
}

type TrustPinApproveRequest struct {
	TenantID    string
	UserID      string
	DeviceID    string
	ChallengeID string
	Signature   string
	Payload     map[string]any
	TOTPCode    string
}

type TrustPinApproveResponse struct {
	ChallengeID string `json:"challenge_id"`
	Status      string `json:"status"`
}
