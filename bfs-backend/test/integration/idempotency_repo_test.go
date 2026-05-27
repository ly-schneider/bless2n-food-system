package integration

import (
	"context"
	"testing"
	"time"

	"backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIdempotencyRepository_ClaimAndFillResponse(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	ctx := context.Background()

	t.Run("Claim of fresh key returns new row with existed=false", func(t *testing.T) {
		row, existed, err := repos.Idempotency.Claim(ctx, "order", "fresh-key-1", time.Hour)
		require.NoError(t, err)
		require.False(t, existed)
		require.NotEqual(t, uuid.Nil, row.ID)
		require.Empty(t, row.Response, "newly claimed row must have nil response")
	})

	t.Run("Claim of existing key returns existing row with existed=true", func(t *testing.T) {
		first, existed1, err := repos.Idempotency.Claim(ctx, "order", "shared-key", time.Hour)
		require.NoError(t, err)
		require.False(t, existed1)

		second, existed2, err := repos.Idempotency.Claim(ctx, "order", "shared-key", time.Hour)
		require.NoError(t, err)
		require.True(t, existed2, "second Claim of same key must report existed=true")
		require.Equal(t, first.ID, second.ID, "Claim conflict must return the original row")
	})

	t.Run("FillResponse populates the response payload", func(t *testing.T) {
		claimed, _, err := repos.Idempotency.Claim(ctx, "order", "fill-key", time.Hour)
		require.NoError(t, err)
		require.Empty(t, claimed.Response)

		payload := map[string]any{"orderId": "abc", "totalCents": 1500.0}
		require.NoError(t, repos.Idempotency.FillResponse(ctx, claimed.ID, payload))

		fetched, err := repos.Idempotency.Get(ctx, "order", "fill-key")
		require.NoError(t, err)
		gotPayload, err := repository.GetResponseMap(fetched)
		require.NoError(t, err)
		require.Equal(t, "abc", gotPayload["orderId"])
		require.InDelta(t, 1500.0, gotPayload["totalCents"], 0.001)
	})

	t.Run("Discard removes a claimed placeholder", func(t *testing.T) {
		claimed, _, err := repos.Idempotency.Claim(ctx, "order", "discard-key", time.Hour)
		require.NoError(t, err)

		require.NoError(t, repos.Idempotency.Discard(ctx, claimed.ID))

		// After Discard, the key is available again for a fresh Claim
		again, existed, err := repos.Idempotency.Claim(ctx, "order", "discard-key", time.Hour)
		require.NoError(t, err)
		require.False(t, existed, "after Discard, the same key should claim as fresh")
		require.NotEqual(t, claimed.ID, again.ID)
	})

	t.Run("scope isolates keys", func(t *testing.T) {
		_, existed1, err := repos.Idempotency.Claim(ctx, "scope-a", "same-key", time.Hour)
		require.NoError(t, err)
		require.False(t, existed1)

		_, existed2, err := repos.Idempotency.Claim(ctx, "scope-b", "same-key", time.Hour)
		require.NoError(t, err)
		require.False(t, existed2, "same key in different scope must be a fresh claim")
	})

	t.Run("Claim-then-FillResponse round-trip serves cached response on replay", func(t *testing.T) {
		claimed, _, err := repos.Idempotency.Claim(ctx, "order", "replay-key", time.Hour)
		require.NoError(t, err)
		require.NoError(t, repos.Idempotency.FillResponse(ctx, claimed.ID, map[string]any{"orderId": "xyz"}))

		replay, existed, err := repos.Idempotency.Claim(ctx, "order", "replay-key", time.Hour)
		require.NoError(t, err)
		require.True(t, existed)
		respMap, err := repository.GetResponseMap(replay)
		require.NoError(t, err)
		require.Equal(t, "xyz", respMap["orderId"])
	})
}
