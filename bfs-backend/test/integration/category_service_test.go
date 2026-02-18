package integration

import (
	"context"
	"testing"

	"backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCategoryService_CRUD(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	svc := service.NewCategoryService(repos.Category)
	ctx := context.Background()

	t.Run("Create category", func(t *testing.T) {
		cat, err := svc.Create(ctx, "Drinks", 1)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, cat.ID)
		require.Equal(t, "Drinks", cat.Name)
		require.Equal(t, 1, cat.Position)
		require.True(t, cat.IsActive)
	})

	t.Run("GetAll returns created categories", func(t *testing.T) {
		// Create additional category
		_, err := svc.Create(ctx, "Food", 2)
		require.NoError(t, err)

		cats, err := svc.GetAll(ctx)
		require.NoError(t, err)
		require.Len(t, cats, 2)
		// Should be ordered by position
		require.Equal(t, "Drinks", cats[0].Name)
		require.Equal(t, "Food", cats[1].Name)
	})

	t.Run("GetActive returns only active categories", func(t *testing.T) {
		// Create an inactive category
		cat, err := svc.Create(ctx, "Desserts", 3)
		require.NoError(t, err)

		// Deactivate it
		_, err = svc.Update(ctx, cat.ID, "Desserts", 3, false)
		require.NoError(t, err)

		active, err := svc.GetActive(ctx)
		require.NoError(t, err)
		require.Len(t, active, 2) // Only Drinks and Food
	})

	t.Run("GetByID returns specific category", func(t *testing.T) {
		cats, err := svc.GetAll(ctx)
		require.NoError(t, err)
		require.True(t, len(cats) > 0)

		cat, err := svc.GetByID(ctx, cats[0].ID)
		require.NoError(t, err)
		require.Equal(t, cats[0].Name, cat.Name)
	})

	t.Run("Update modifies category", func(t *testing.T) {
		cats, err := svc.GetAll(ctx)
		require.NoError(t, err)
		require.True(t, len(cats) > 0)

		updated, err := svc.Update(ctx, cats[0].ID, "Beverages", 10, true)
		require.NoError(t, err)
		require.Equal(t, "Beverages", updated.Name)
		require.Equal(t, 10, updated.Position)
	})

	t.Run("Delete removes category", func(t *testing.T) {
		cats, err := svc.GetAll(ctx)
		require.NoError(t, err)
		initialCount := len(cats)

		err = svc.Delete(ctx, cats[0].ID)
		require.NoError(t, err)

		cats, err = svc.GetAll(ctx)
		require.NoError(t, err)
		require.Len(t, cats, initialCount-1)
	})

	t.Run("List with pagination", func(t *testing.T) {
		tdb.Cleanup(t)

		// Create 5 categories
		for i := 0; i < 5; i++ {
			_, err := svc.Create(ctx, "Category"+string(rune('A'+i)), i)
			require.NoError(t, err)
		}

		// Test pagination
		items, total, err := svc.List(ctx, 2, 0)
		require.NoError(t, err)
		require.Equal(t, int64(5), total)
		require.Len(t, items, 2)

		// Second page
		items, total, err = svc.List(ctx, 2, 2)
		require.NoError(t, err)
		require.Equal(t, int64(5), total)
		require.Len(t, items, 2)

		// Third page
		items, total, err = svc.List(ctx, 2, 4)
		require.NoError(t, err)
		require.Equal(t, int64(5), total)
		require.Len(t, items, 1)
	})
}

func TestCategoryService_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	svc := service.NewCategoryService(repos.Category)
	ctx := context.Background()

	t.Run("GetByID returns error for non-existent ID", func(t *testing.T) {
		_, err := svc.GetByID(ctx, uuid.Must(uuid.NewV7()))
		require.Error(t, err)
	})

	t.Run("Update returns error for non-existent ID", func(t *testing.T) {
		_, err := svc.Update(ctx, uuid.Must(uuid.NewV7()), "Test", 1, true)
		require.Error(t, err)
	})
}
