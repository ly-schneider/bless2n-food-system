package integration

import (
	"context"
	"testing"

	"backend/internal/generated/ent/product"
	"backend/internal/generated/ent/settings"
	"backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSettingsService_GetSettings(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)

	svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)
	ctx := context.Background()

	t.Run("GetSettings returns default settings when none exist", func(t *testing.T) {
		s, err := svc.GetSettings(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.Equal(t, "default", s.ID)
		require.Equal(t, settings.PosModeQR_CODE, s.PosMode)
	})
}

func TestSettingsService_SetPosMode(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)
	ctx := context.Background()

	t.Run("SetPosMode to QR_CODE succeeds", func(t *testing.T) {
		err := svc.SetPosMode(ctx, settings.PosModeQR_CODE)
		require.NoError(t, err)

		s, err := svc.GetSettings(ctx)
		require.NoError(t, err)
		require.Equal(t, settings.PosModeQR_CODE, s.PosMode)
	})

	t.Run("SetPosMode to JETON succeeds when no active products without jetons", func(t *testing.T) {
		tdb.Cleanup(t)

		err := svc.SetPosMode(ctx, settings.PosModeJETON)
		require.NoError(t, err)
	})

	t.Run("SetPosMode to JETON fails when active products lack jetons", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)

		err := svc.SetPosMode(ctx, settings.PosModeJETON)
		require.Error(t, err)

		var missingErr service.MissingJetonForActiveProductsError
		require.ErrorAs(t, err, &missingErr)
		require.EqualValues(t, 1, missingErr.Count)
	})

	t.Run("SetPosMode to JETON succeeds when all active products have jetons", func(t *testing.T) {
		tdb.Cleanup(t)

		category := fixtures.CreateCategory("Drinks", 1, true)
		jeton := fixtures.CreateJeton("Red", "#EF4444")
		fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, &jeton.ID)

		err := svc.SetPosMode(ctx, settings.PosModeJETON)
		require.NoError(t, err)
	})

	t.Run("SetPosMode with invalid mode fails", func(t *testing.T) {
		err := svc.SetPosMode(ctx, "INVALID_MODE")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid_mode")
	})
}

func TestSettingsService_JetonCRUD(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)

	svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)
	ctx := context.Background()

	t.Run("CreateJeton creates new jeton with hex color", func(t *testing.T) {
		jeton, err := svc.CreateJeton(ctx, "Red Token", "#EF4444")
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, jeton.ID)
		require.Equal(t, "Red Token", jeton.Name)
		require.Equal(t, "#EF4444", jeton.Color)
	})

	t.Run("CreateJeton normalizes hex color", func(t *testing.T) {
		jeton, err := svc.CreateJeton(ctx, "Normalized Token", "ff5733")
		require.NoError(t, err)
		require.Equal(t, "#FF5733", jeton.Color)
	})

	t.Run("CreateJeton fails with empty name", func(t *testing.T) {
		_, err := svc.CreateJeton(ctx, "", "#EF4444")
		require.Error(t, err)
		require.Contains(t, err.Error(), "name_required")
	})

	t.Run("CreateJeton fails without color", func(t *testing.T) {
		_, err := svc.CreateJeton(ctx, "No Color", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "color_required")
	})

	t.Run("CreateJeton fails with invalid hex color", func(t *testing.T) {
		_, err := svc.CreateJeton(ctx, "Invalid", "not-a-hex")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid_hex")
	})

	t.Run("ListJetons returns all jetons", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := svc.CreateJeton(ctx, "Jeton A", "#EF4444")
		require.NoError(t, err)
		_, err = svc.CreateJeton(ctx, "Jeton B", "#3B82F6")
		require.NoError(t, err)

		jetons, err := svc.ListJetons(ctx)
		require.NoError(t, err)
		require.Len(t, jetons, 2)

		for _, j := range jetons {
			require.NotEmpty(t, j.Name)
		}
	})

	t.Run("UpdateJeton modifies jeton", func(t *testing.T) {
		tdb.Cleanup(t)

		created, err := svc.CreateJeton(ctx, "Original", "#EF4444")
		require.NoError(t, err)

		updated, err := svc.UpdateJeton(ctx, created.ID, "Updated", "#3B82F6")
		require.NoError(t, err)
		require.Equal(t, "Updated", updated.Name)
		require.Equal(t, "#3B82F6", updated.Color)
	})

	t.Run("UpdateJeton fails for non-existent jeton", func(t *testing.T) {
		_, err := svc.UpdateJeton(ctx, uuid.Must(uuid.NewV7()), "Test", "#EF4444")
		require.Error(t, err)
	})

	t.Run("DeleteJeton removes unused jeton", func(t *testing.T) {
		tdb.Cleanup(t)

		created, err := svc.CreateJeton(ctx, "ToDelete", "#EF4444")
		require.NoError(t, err)

		err = svc.DeleteJeton(ctx, created.ID)
		require.NoError(t, err)

		jetons, err := svc.ListJetons(ctx)
		require.NoError(t, err)
		require.Empty(t, jetons)
	})

	t.Run("DeleteJeton fails for jeton in use", func(t *testing.T) {
		tdb.Cleanup(t)

		repos := NewRepositories(tdb.Client)
		fixtures := NewFixtures(repos)
		svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)

		jeton := fixtures.CreateJeton("InUse", "#EF4444")
		category := fixtures.CreateCategory("Drinks", 1, true)
		fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, &jeton.ID)

		err := svc.DeleteJeton(ctx, jeton.ID)
		require.Error(t, err)

		var inUseErr service.JetonInUseError
		require.ErrorAs(t, err, &inUseErr)
		require.EqualValues(t, 1, inUseErr.Count)
	})
}

func TestSettingsService_SetProductJeton(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)
	ctx := context.Background()

	t.Run("SetProductJeton assigns jeton to product", func(t *testing.T) {
		category := fixtures.CreateCategory("Drinks", 1, true)
		p := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, nil)
		jeton := fixtures.CreateJeton("Red", "#EF4444")

		err := svc.SetProductJeton(ctx, p.ID, &jeton.ID)
		require.NoError(t, err)

		updated, err := repos.Product.GetByID(ctx, p.ID)
		require.NoError(t, err)
		require.NotNil(t, updated.JetonID)
		require.Equal(t, jeton.ID, *updated.JetonID)
	})

	t.Run("SetProductJeton removes jeton from product", func(t *testing.T) {
		category := fixtures.CreateCategory("Food", 2, true)
		jeton := fixtures.CreateJeton("Blue", "#3B82F6")
		p := fixtures.CreateProduct("Burger", category.ID, 1200, product.TypeSimple, &jeton.ID)

		err := svc.SetPosMode(ctx, settings.PosModeQR_CODE)
		require.NoError(t, err)

		err = svc.SetProductJeton(ctx, p.ID, nil)
		require.NoError(t, err)

		updated, err := repos.Product.GetByID(ctx, p.ID)
		require.NoError(t, err)
		require.Nil(t, updated.JetonID)
	})

	t.Run("SetProductJeton fails for non-existent product", func(t *testing.T) {
		jeton := fixtures.CreateJeton("Green", "#22C55E")
		err := svc.SetProductJeton(ctx, uuid.Must(uuid.NewV7()), &jeton.ID)
		require.Error(t, err)
	})

	t.Run("SetProductJeton fails for non-existent jeton", func(t *testing.T) {
		category := fixtures.CreateCategory("Snacks", 3, true)
		p := fixtures.CreateProduct("Chips", category.ID, 200, product.TypeSimple, nil)

		nonExistentJeton := uuid.Must(uuid.NewV7())
		err := svc.SetProductJeton(ctx, p.ID, &nonExistentJeton)
		require.Error(t, err)
	})

	t.Run("SetProductJeton fails to remove jeton in JETON mode for active simple product", func(t *testing.T) {
		tdb.Cleanup(t)

		repos := NewRepositories(tdb.Client)
		fixtures := NewFixtures(repos)
		svc := service.NewSettingsService(repos.Settings, repos.Jeton, repos.Product)

		category := fixtures.CreateCategory("Drinks", 1, true)
		jeton := fixtures.CreateJeton("Red", "#EF4444")
		p := fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, &jeton.ID)

		err := svc.SetPosMode(ctx, settings.PosModeJETON)
		require.NoError(t, err)

		err = svc.SetProductJeton(ctx, p.ID, nil)
		require.Error(t, err)
		require.Equal(t, service.ErrJetonRequired, err)
	})
}
