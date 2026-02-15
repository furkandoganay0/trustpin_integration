package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"trustpin_integration/internal/domain"
)

type UserRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{users: make(map[string]*domain.User)}
}

func (r *UserRepo) GetByUsername(ctx context.Context, tenantID domain.TenantID, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.TenantID == tenantID && u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (r *UserRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok || u.TenantID != tenantID {
		return nil, nil
	}
	return u, nil
}

// Seed adds a demo user for local boot.
func (r *UserRepo) Seed(user *domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
}

type SessionRepo struct {
	mu       sync.Mutex
	sessions map[string]*domain.Session
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{sessions: make(map[string]*domain.Session)}
}

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	return nil
}

func (r *SessionRepo) RevokeByJWTID(ctx context.Context, tenantID domain.TenantID, jwtID string, revokedAt time.Time) error {
	return nil
}

type DeviceRepo struct {
	mu      sync.RWMutex
	devices map[string]*domain.MFADevice
}

func NewDeviceRepo() *DeviceRepo {
	return &DeviceRepo{devices: make(map[string]*domain.MFADevice)}
}

func (r *DeviceRepo) Create(ctx context.Context, d *domain.MFADevice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.devices[d.ID]; ok {
		return errors.New("device_exists")
	}
	r.devices[d.ID] = d
	return nil
}

func (r *DeviceRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFADevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.devices[id]
	if !ok || d.TenantID != tenantID {
		return nil, nil
	}
	return d, nil
}

func (r *DeviceRepo) UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.devices[id]
	if !ok || d.TenantID != tenantID {
		return errors.New("not_found")
	}
	d.State = state
	d.UpdatedAt = time.Now()
	return nil
}

type ChallengeRepo struct {
	mu         sync.RWMutex
	challenges map[string]*domain.MFAChallenge
}

func NewChallengeRepo() *ChallengeRepo {
	return &ChallengeRepo{challenges: make(map[string]*domain.MFAChallenge)}
}

func (r *ChallengeRepo) Create(ctx context.Context, c *domain.MFAChallenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.challenges[c.ID] = c
	return nil
}

func (r *ChallengeRepo) GetByID(ctx context.Context, tenantID domain.TenantID, id string) (*domain.MFAChallenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.challenges[id]
	if !ok || c.TenantID != tenantID {
		return nil, nil
	}
	return c, nil
}

func (r *ChallengeRepo) UpdateState(ctx context.Context, tenantID domain.TenantID, id, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.challenges[id]
	if !ok || c.TenantID != tenantID {
		return errors.New("not_found")
	}
	c.State = state
	c.UpdatedAt = time.Now()
	return nil
}
