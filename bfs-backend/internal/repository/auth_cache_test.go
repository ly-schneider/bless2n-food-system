package repository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/devicebinding"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeSessionRepo struct {
	mu              sync.Mutex
	getCalls        int
	refreshCalls    int32
	refreshDelay    time.Duration
	refreshReleased chan struct{}
	session         *SessionWithUser
	getErr          error
	refreshErr      error
}

func (f *fakeSessionRepo) GetByToken(_ context.Context, _ string) (*SessionWithUser, error) {
	f.mu.Lock()
	f.getCalls++
	f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	copy := *f.session
	return &copy, nil
}

func (f *fakeSessionRepo) RefreshSession(_ context.Context, _ string, _ time.Duration) error {
	atomic.AddInt32(&f.refreshCalls, 1)
	if f.refreshDelay > 0 {
		time.Sleep(f.refreshDelay)
	}
	if f.refreshReleased != nil {
		<-f.refreshReleased
	}
	return f.refreshErr
}

func (f *fakeSessionRepo) CreateSession(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}

type fakeBindingRepo struct {
	mu            sync.Mutex
	getCalls      int
	lastSeenCalls int32
	lastSeenDelay time.Duration
	binding       *ent.DeviceBinding
	getErr        error
	lastSeenErr   error
}

func (f *fakeBindingRepo) GetByID(_ context.Context, _ uuid.UUID) (*ent.DeviceBinding, error) {
	return f.binding, f.getErr
}

func (f *fakeBindingRepo) GetByTokenHash(_ context.Context, _ string) (*ent.DeviceBinding, error) {
	f.mu.Lock()
	f.getCalls++
	f.mu.Unlock()
	return f.binding, f.getErr
}

func (f *fakeBindingRepo) Create(_ context.Context, _ devicebinding.DeviceType, _, _ string, _ *string, _, _ *uuid.UUID) (*ent.DeviceBinding, error) {
	return f.binding, nil
}

func (f *fakeBindingRepo) UpdateLastSeen(_ context.Context, _ uuid.UUID) error {
	atomic.AddInt32(&f.lastSeenCalls, 1)
	if f.lastSeenDelay > 0 {
		time.Sleep(f.lastSeenDelay)
	}
	return f.lastSeenErr
}

func (f *fakeBindingRepo) Revoke(_ context.Context, _ uuid.UUID) error { return nil }

func (f *fakeBindingRepo) ListActive(_ context.Context) ([]*ent.DeviceBinding, error) {
	return nil, nil
}

func (f *fakeBindingRepo) ListByType(_ context.Context, _ devicebinding.DeviceType) ([]*ent.DeviceBinding, error) {
	return nil, nil
}

func newTestSession() *SessionWithUser {
	return &SessionWithUser{
		UserID:    "user-1",
		UpdatedAt: time.Now().Add(-48 * time.Hour).UTC(),
	}
}

func TestCachedSessionRepo_GetByToken_HitAfterMiss(t *testing.T) {
	inner := &fakeSessionRepo{session: newTestSession()}
	repo := NewCachedSessionRepository(inner)
	ctx := context.Background()

	_, err := repo.GetByToken(ctx, "tok")
	require.NoError(t, err)
	_, err = repo.GetByToken(ctx, "tok")
	require.NoError(t, err)

	require.Equal(t, 1, inner.getCalls, "second call should be served from cache")
}

func TestCachedSessionRepo_GetByToken_DistinctTokensMiss(t *testing.T) {
	inner := &fakeSessionRepo{session: newTestSession()}
	repo := NewCachedSessionRepository(inner)
	ctx := context.Background()

	_, _ = repo.GetByToken(ctx, "a")
	_, _ = repo.GetByToken(ctx, "b")

	require.Equal(t, 2, inner.getCalls)
}

func TestCachedSessionRepo_GetByToken_ErrorNotCached(t *testing.T) {
	inner := &fakeSessionRepo{getErr: errors.New("boom")}
	repo := NewCachedSessionRepository(inner)
	ctx := context.Background()

	_, err1 := repo.GetByToken(ctx, "tok")
	_, err2 := repo.GetByToken(ctx, "tok")
	require.Error(t, err1)
	require.Error(t, err2)
	require.Equal(t, 2, inner.getCalls, "errors must not be cached")
}

func TestCachedSessionRepo_RefreshSession_SyncBumpsCache(t *testing.T) {
	inner := &fakeSessionRepo{session: newTestSession()}
	repo := NewCachedSessionRepository(inner)
	ctx := context.Background()

	cached, err := repo.GetByToken(ctx, "tok")
	require.NoError(t, err)
	originalUpdated := cached.UpdatedAt

	require.NoError(t, repo.RefreshSession(ctx, "tok", time.Hour))
	require.Equal(t, int32(1), atomic.LoadInt32(&inner.refreshCalls))

	fresh, err := repo.GetByToken(ctx, "tok")
	require.NoError(t, err)
	require.True(t, fresh.UpdatedAt.After(originalUpdated), "cache entry UpdatedAt should be bumped after refresh")
}

func TestCachedSessionRepo_RefreshSessionAsync_DoesNotBlock(t *testing.T) {
	released := make(chan struct{})
	inner := &fakeSessionRepo{
		session:         newTestSession(),
		refreshReleased: released,
	}
	cached := NewCachedSessionRepository(inner).(*cachedSessionRepo)
	_, _ = cached.GetByToken(context.Background(), "tok")

	start := time.Now()
	cached.RefreshSessionAsync("tok", time.Hour, zap.NewNop())
	require.Less(t, time.Since(start), 50*time.Millisecond, "async refresh must not block caller")

	close(released)
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&inner.refreshCalls) == 1
	}, time.Second, 5*time.Millisecond)
}

func TestCachedSessionRepo_RefreshSessionAsync_Singleflight(t *testing.T) {
	inner := &fakeSessionRepo{
		session:      newTestSession(),
		refreshDelay: 300 * time.Millisecond,
	}
	cached := NewCachedSessionRepository(inner).(*cachedSessionRepo)
	_, _ = cached.GetByToken(context.Background(), "tok")

	var spawned sync.WaitGroup
	for i := 0; i < 50; i++ {
		spawned.Add(1)
		go func() {
			defer spawned.Done()
			cached.RefreshSessionAsync("tok", time.Hour, zap.NewNop())
		}()
	}
	spawned.Wait()

	time.Sleep(700 * time.Millisecond)

	require.Equal(t, int32(1), atomic.LoadInt32(&inner.refreshCalls), "concurrent refreshes must collapse to one DB write")
}

func TestCachedSessionRepo_RefreshSessionAsync_OptimisticUpdatedAtBump(t *testing.T) {
	released := make(chan struct{})
	inner := &fakeSessionRepo{
		session:         newTestSession(),
		refreshReleased: released,
	}
	cached := NewCachedSessionRepository(inner).(*cachedSessionRepo)

	first, _ := cached.GetByToken(context.Background(), "tok")
	originalUpdated := first.UpdatedAt

	cached.RefreshSessionAsync("tok", time.Hour, zap.NewNop())

	bumped, _ := cached.GetByToken(context.Background(), "tok")
	require.True(t, bumped.UpdatedAt.After(originalUpdated), "subsequent reads should see bumped UpdatedAt before DB write completes")

	close(released)
}

func TestCachedBindingRepo_GetByTokenHash_HitAfterMiss(t *testing.T) {
	inner := &fakeBindingRepo{binding: &ent.DeviceBinding{ID: uuid.Must(uuid.NewV7())}}
	repo := NewCachedDeviceBindingRepository(inner)
	ctx := context.Background()

	_, err := repo.GetByTokenHash(ctx, "hash")
	require.NoError(t, err)
	_, err = repo.GetByTokenHash(ctx, "hash")
	require.NoError(t, err)

	require.Equal(t, 1, inner.getCalls)
}

func TestCachedBindingRepo_UpdateLastSeenDebounced_FiresOnceWithinWindow(t *testing.T) {
	inner := &fakeBindingRepo{}
	repo := NewCachedDeviceBindingRepository(inner).(*cachedDeviceBindingRepo)
	id := uuid.Must(uuid.NewV7())

	for i := 0; i < 25; i++ {
		repo.UpdateLastSeenDebounced(id, zap.NewNop())
	}

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&inner.lastSeenCalls) == 1
	}, time.Second, 5*time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&inner.lastSeenCalls), "calls within debounce window must collapse to one DB write")
}

func TestCachedBindingRepo_UpdateLastSeenDebounced_DifferentIDsAllFire(t *testing.T) {
	inner := &fakeBindingRepo{}
	repo := NewCachedDeviceBindingRepository(inner).(*cachedDeviceBindingRepo)

	for i := 0; i < 5; i++ {
		repo.UpdateLastSeenDebounced(uuid.Must(uuid.NewV7()), zap.NewNop())
	}

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&inner.lastSeenCalls) == 5
	}, time.Second, 5*time.Millisecond)
}

func TestCachedBindingRepo_Revoke_ClearsCacheEntry(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	inner := &fakeBindingRepo{binding: &ent.DeviceBinding{ID: id}}
	repo := NewCachedDeviceBindingRepository(inner)
	ctx := context.Background()

	_, _ = repo.GetByTokenHash(ctx, "hash")
	require.Equal(t, 1, inner.getCalls)
	_, _ = repo.GetByTokenHash(ctx, "hash")
	require.Equal(t, 1, inner.getCalls)

	require.NoError(t, repo.Revoke(ctx, id))

	_, _ = repo.GetByTokenHash(ctx, "hash")
	require.Equal(t, 2, inner.getCalls, "Revoke must invalidate cached entries for that binding")
}
