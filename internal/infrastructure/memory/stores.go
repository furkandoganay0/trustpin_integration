package memory

import (
	"context"
	"sync"
	"time"

	"trustpin_integration/internal/domain"
)

type NonceStore struct {
	mu    sync.Mutex
	items map[string]time.Time
}

func NewNonceStore() *NonceStore {
	return &NonceStore{items: make(map[string]time.Time)}
}

func (s *NonceStore) CheckAndSet(ctx context.Context, tenantID domain.TenantID, nonce string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(tenantID) + ":" + nonce
	if exp, ok := s.items[key]; ok && time.Now().Before(exp) {
		return false, nil
	}
	s.items[key] = time.Now().Add(ttl)
	return true, nil
}

type IdempotencyStore struct {
	mu    sync.Mutex
	items map[string]idempotentItem
}

type idempotentItem struct {
	value []byte
	exp   time.Time
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{items: make(map[string]idempotentItem)}
}

func (s *IdempotencyStore) Get(ctx context.Context, tenantID domain.TenantID, key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := string(tenantID) + ":" + key
	item, ok := s.items[k]
	if !ok {
		return nil, false, nil
	}
	if time.Now().After(item.exp) {
		delete(s.items, k)
		return nil, false, nil
	}
	return item.value, true, nil
}

func (s *IdempotencyStore) Set(ctx context.Context, tenantID domain.TenantID, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := string(tenantID) + ":" + key
	s.items[k] = idempotentItem{value: value, exp: time.Now().Add(ttl)}
	return nil
}
