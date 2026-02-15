package postgres

import (
	"context"
	"errors"
	"time"

	"trustpin_integration/internal/domain"
)

type UserRepo struct{}

type SessionRepo struct{}

type DeviceRepo struct{}

type ChallengeRepo struct{}

func (r *UserRepo) GetByUsername(ctx context.Context, tenantID domain.TenantID, username string) (*domain.User, error) {
	return nil, errors.New("not_implemented")
}

func (r *UserRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.User, error) {
	return nil, errors.New("not_implemented")
}

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	return errors.New("not_implemented")
}

func (r *SessionRepo) RevokeByJWTID(ctx context.Context, tenantID domain.TenantID, jwtID string, revokedAt time.Time) error {
	return errors.New("not_implemented")
}

func (r *DeviceRepo) Create(ctx context.Context, d *domain.MFADevice) error {
	return errors.New("not_implemented")
}

func (r *DeviceRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFADevice, error) {
	return nil, errors.New("not_implemented")
}

func (r *DeviceRepo) UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error {
	return errors.New("not_implemented")
}

func (r *ChallengeRepo) Create(ctx context.Context, c *domain.MFAChallenge) error {
	return errors.New("not_implemented")
}

func (r *ChallengeRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFAChallenge, error) {
	return nil, errors.New("not_implemented")
}

func (r *ChallengeRepo) UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error {
	return errors.New("not_implemented")
}
