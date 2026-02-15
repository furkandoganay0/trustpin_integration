package idempotency

import (
	"context"
	"errors"
	"time"

	"trustpin_integration/internal/domain"
)

type Store struct{}

func (s *Store) Get(ctx context.Context, tenantID domain.TenantID, key string) ([]byte, bool, error) {
	return nil, false, errors.New("not_implemented")
}

func (s *Store) Set(ctx context.Context, tenantID domain.TenantID, key string, value []byte, ttl time.Duration) error {
	return errors.New("not_implemented")
}
