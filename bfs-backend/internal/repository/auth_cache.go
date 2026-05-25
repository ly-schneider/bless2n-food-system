package repository

import (
	"context"
	"sync"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/devicebinding"

	"github.com/google/uuid"
)

const authCacheTTL = 30 * time.Second

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

type ttlCache[V any] struct {
	mu      sync.RWMutex
	entries map[string]ttlEntry[V]
}

func newTTLCache[V any]() *ttlCache[V] {
	return &ttlCache[V]{entries: make(map[string]ttlEntry[V])}
}

func (c *ttlCache[V]) get(key string, now time.Time) (V, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || now.After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.value, true
}

func (c *ttlCache[V]) set(key string, value V, expiresAt time.Time) {
	c.mu.Lock()
	c.entries[key] = ttlEntry[V]{value: value, expiresAt: expiresAt}
	if len(c.entries) > 4096 {
		now := time.Now()
		for k, v := range c.entries {
			if now.After(v.expiresAt) {
				delete(c.entries, k)
			}
		}
	}
	c.mu.Unlock()
}

func (c *ttlCache[V]) invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

type cachedSessionRepo struct {
	inner SessionRepository
	cache *ttlCache[*SessionWithUser]
}

// NewCachedSessionRepository wraps a SessionRepository with a short-TTL in-memory
// cache for GetByToken. The cache is per-replica; tokens revoked via the auth
// service take at most authCacheTTL to fully propagate.
func NewCachedSessionRepository(inner SessionRepository) SessionRepository {
	return &cachedSessionRepo{
		inner: inner,
		cache: newTTLCache[*SessionWithUser](),
	}
}

func (r *cachedSessionRepo) GetByToken(ctx context.Context, token string) (*SessionWithUser, error) {
	now := time.Now()
	if v, ok := r.cache.get(token, now); ok {
		return v, nil
	}
	v, err := r.inner.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	r.cache.set(token, v, now.Add(authCacheTTL))
	return v, nil
}

func (r *cachedSessionRepo) RefreshSession(ctx context.Context, token string, expiresIn time.Duration) error {
	r.cache.invalidate(token)
	return r.inner.RefreshSession(ctx, token, expiresIn)
}

func (r *cachedSessionRepo) CreateSession(ctx context.Context, userID string, expiresIn time.Duration) (string, error) {
	return r.inner.CreateSession(ctx, userID, expiresIn)
}

type cachedDeviceBindingRepo struct {
	inner DeviceBindingRepository
	cache *ttlCache[*ent.DeviceBinding]
}

// NewCachedDeviceBindingRepository wraps a DeviceBindingRepository with a
// short-TTL in-memory cache for GetByTokenHash. Revocations propagate within
// authCacheTTL.
func NewCachedDeviceBindingRepository(inner DeviceBindingRepository) DeviceBindingRepository {
	return &cachedDeviceBindingRepo{
		inner: inner,
		cache: newTTLCache[*ent.DeviceBinding](),
	}
}

func (r *cachedDeviceBindingRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.DeviceBinding, error) {
	return r.inner.GetByID(ctx, id)
}

func (r *cachedDeviceBindingRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*ent.DeviceBinding, error) {
	now := time.Now()
	if v, ok := r.cache.get(tokenHash, now); ok {
		return v, nil
	}
	v, err := r.inner.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	r.cache.set(tokenHash, v, now.Add(authCacheTTL))
	return v, nil
}

func (r *cachedDeviceBindingRepo) Create(ctx context.Context, deviceType devicebinding.DeviceType, tokenHash, createdByUserID string, name *string, deviceID, stationID *uuid.UUID) (*ent.DeviceBinding, error) {
	return r.inner.Create(ctx, deviceType, tokenHash, createdByUserID, name, deviceID, stationID)
}

func (r *cachedDeviceBindingRepo) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	return r.inner.UpdateLastSeen(ctx, id)
}

func (r *cachedDeviceBindingRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	r.cache.mu.Lock()
	for k, v := range r.cache.entries {
		if v.value != nil && v.value.ID == id {
			delete(r.cache.entries, k)
		}
	}
	r.cache.mu.Unlock()
	return r.inner.Revoke(ctx, id)
}

func (r *cachedDeviceBindingRepo) ListActive(ctx context.Context) ([]*ent.DeviceBinding, error) {
	return r.inner.ListActive(ctx)
}

func (r *cachedDeviceBindingRepo) ListByType(ctx context.Context, deviceType devicebinding.DeviceType) ([]*ent.DeviceBinding, error) {
	return r.inner.ListByType(ctx, deviceType)
}
