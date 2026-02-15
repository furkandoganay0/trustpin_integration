package redis

import (
	"context"
	"errors"
	"time"

	"trustpin_integration/internal/domain"
)

type NonceStore struct{}

func (s *NonceStore) CheckAndSet(ctx context.Context, tenantID domain.TenantID, nonce string, ttl time.Duration) (bool, error) {
	return false, errors.New("not_implemented")
}
