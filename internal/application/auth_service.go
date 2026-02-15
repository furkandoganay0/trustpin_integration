package application

import (
	"context"
	"errors"
	"time"

	"trustpin_integration/internal/domain"
)

type AuthService struct {
	Users    UserRepository
	Sessions SessionRepository
	// JWT signing handled elsewhere; this service focuses on domain flow.
}

func (s *AuthService) Login(ctx context.Context, tenantID domain.TenantID, username, password string) (*domain.Session, error) {
	user, err := s.Users.GetByUsername(ctx, tenantID, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid_credentials")
	}
	// Password verification intentionally omitted; plug in password hasher.
	if password == "" {
		return nil, errors.New("invalid_credentials")
	}

	session := &domain.Session{
		ID:        "",
		TenantID:  tenantID,
		UserID:    user.ID,
		JWTID:     "",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.Sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}
