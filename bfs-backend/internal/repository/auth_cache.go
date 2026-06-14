package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/devicebinding"
	"backend/internal/trace"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	authCacheTTL        = 30 * time.Second
	refreshAsyncBudget  = 5 * time.Second
	lastSeenDebounce    = 60 * time.Second
	lastSeenAsyncBudget = 5 * time.Second
)

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

func recordCacheResult(ctx context.Context, layer string, hit bool) {
	outcome := "miss"
	if hit {
		outcome = "hit"
	}
	trace.Data(ctx, "cache.layer", layer)
	trace.Data(ctx, "cache.outcome", outcome)
	trace.Tag(ctx, "cache.outcome", outcome)
}

type cachedSessionRepo struct {
	inner     SessionRepository
	cache     *ttlCache[*SessionWithUser]
	refreshSF singleflight.Group
}

// NewCachedSessionRepository: cache is per-replica; revoked tokens take up to
// authCacheTTL to propagate.
func NewCachedSessionRepository(inner SessionRepository) SessionRepository {
	return &cachedSessionRepo{
		inner: inner,
		cache: newTTLCache[*SessionWithUser](),
	}
}

func (r *cachedSessionRepo) GetByToken(ctx context.Context, token string) (*SessionWithUser, error) {
	ctx, finish := trace.StartSpan(ctx, "cache", "session.GetByToken")
	defer finish()

	now := time.Now()
	if v, ok := r.cache.get(token, now); ok {
		recordCacheResult(ctx, "session", true)
		if v == nil {
			return nil, ErrNotFound
		}
		return v, nil
	}
	recordCacheResult(ctx, "session", false)

	v, err := r.inner.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			r.cache.set(token, nil, now.Add(authCacheTTL))
		}
		return nil, err
	}
	r.cache.set(token, v, now.Add(authCacheTTL))
	return v, nil
}

func (r *cachedSessionRepo) RefreshSession(ctx context.Context, token string, expiresIn time.Duration) error {
	r.markRefreshed(token)
	return r.inner.RefreshSession(ctx, token, expiresIn)
}

func (r *cachedSessionRepo) RefreshSessionAsync(token string, expiresIn time.Duration, logger *zap.Logger) {
	r.markRefreshed(token)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), refreshAsyncBudget)
		defer cancel()
		_, err, _ := r.refreshSF.Do("refresh:"+token, func() (any, error) {
			return nil, r.inner.RefreshSession(ctx, token, expiresIn)
		})
		if err != nil && logger != nil {
			logger.Warn("async session refresh failed", zap.Error(err))
		}
	}()
}

// markRefreshed bumps cached UpdatedAt before the DB write so concurrent
// requests skip the sliding-refresh check while the write is in flight.
func (r *cachedSessionRepo) markRefreshed(token string) {
	now := time.Now().UTC()
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()
	e, ok := r.cache.entries[token]
	if !ok || e.value == nil {
		return
	}
	updated := *e.value
	updated.UpdatedAt = now
	e.value = &updated
	r.cache.entries[token] = e
}

func (r *cachedSessionRepo) CreateSession(ctx context.Context, userID string, expiresIn time.Duration) (string, error) {
	return r.inner.CreateSession(ctx, userID, expiresIn)
}

type AsyncSessionRefresher interface {
	RefreshSessionAsync(token string, expiresIn time.Duration, logger *zap.Logger)
}

type cachedDeviceBindingRepo struct {
	inner    DeviceBindingRepository
	cache    *ttlCache[*ent.DeviceBinding]
	lastSeen sync.Map
}

func NewCachedDeviceBindingRepository(inner DeviceBindingRepository) DeviceBindingRepository {
	return &cachedDeviceBindingRepo{
		inner: inner,
		cache: newTTLCache[*ent.DeviceBinding](),
	}
}

func (r *cachedDeviceBindingRepo) UpdateLastSeenDebounced(id string, logger *zap.Logger) {
	now := time.Now()
	if prev, ok := r.lastSeen.Load(id); ok {
		if last, ok := prev.(time.Time); ok && now.Sub(last) < lastSeenDebounce {
			return
		}
	}
	r.lastSeen.Store(id, now)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), lastSeenAsyncBudget)
		defer cancel()
		if err := r.inner.UpdateLastSeen(ctx, id); err != nil && logger != nil {
			logger.Warn("debounced last_seen update failed", zap.Error(err))
		}
	}()
}

type DebouncedLastSeenWriter interface {
	UpdateLastSeenDebounced(id string, logger *zap.Logger)
}

func (r *cachedDeviceBindingRepo) GetByID(ctx context.Context, id string) (*ent.DeviceBinding, error) {
	return r.inner.GetByID(ctx, id)
}

func (r *cachedDeviceBindingRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*ent.DeviceBinding, error) {
	ctx, finish := trace.StartSpan(ctx, "cache", "device_binding.GetByTokenHash")
	defer finish()

	now := time.Now()
	if v, ok := r.cache.get(tokenHash, now); ok {
		recordCacheResult(ctx, "device_binding", true)
		if v == nil {
			return nil, ErrNotFound
		}
		return v, nil
	}
	recordCacheResult(ctx, "device_binding", false)

	v, err := r.inner.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			r.cache.set(tokenHash, nil, now.Add(authCacheTTL))
		}
		return nil, err
	}
	r.cache.set(tokenHash, v, now.Add(authCacheTTL))
	return v, nil
}

func (r *cachedDeviceBindingRepo) Create(ctx context.Context, deviceType devicebinding.DeviceType, tokenHash, createdByUserID string, name *string, deviceID, stationID *string) (*ent.DeviceBinding, error) {
	created, err := r.inner.Create(ctx, deviceType, tokenHash, createdByUserID, name, deviceID, stationID)
	if err != nil {
		return nil, err
	}
	r.cache.set(tokenHash, created, time.Now().Add(authCacheTTL))
	return created, nil
}

func (r *cachedDeviceBindingRepo) UpdateLastSeen(ctx context.Context, id string) error {
	return r.inner.UpdateLastSeen(ctx, id)
}

func (r *cachedDeviceBindingRepo) Revoke(ctx context.Context, id string) error {
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
