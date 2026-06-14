package integration

import (
	"context"
	"testing"

	"backend/internal/generated/ent/product"

	nanoid "backend/internal/id"
	"github.com/stretchr/testify/require"
)

func TestMenuSlotRepository_GetByMenuProductIDs(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	ctx := context.Background()

	category := fixtures.CreateCategory("Menus", 1, true)
	optionCat := fixtures.CreateCategory("Options", 2, true)

	cola := fixtures.CreateProduct("Cola", optionCat.ID, 350, product.TypeSimple, nil)
	fries := fixtures.CreateProduct("Fries", optionCat.ID, 400, product.TypeSimple, nil)

	menuA := fixtures.CreateProduct("Combo A", category.ID, 1500, product.TypeMenu, nil)
	menuB := fixtures.CreateProduct("Combo B", category.ID, 1800, product.TypeMenu, nil)
	menuC := fixtures.CreateProduct("Combo C", category.ID, 2000, product.TypeMenu, nil)

	slotA1 := fixtures.CreateMenuSlot(menuA.ID, "Drink", 1)
	slotA2 := fixtures.CreateMenuSlot(menuA.ID, "Side", 2)
	slotB1 := fixtures.CreateMenuSlot(menuB.ID, "Drink", 1)

	fixtures.CreateMenuSlotOption(slotA1.ID, cola.ID)
	fixtures.CreateMenuSlotOption(slotA2.ID, fries.ID)
	fixtures.CreateMenuSlotOption(slotB1.ID, cola.ID)

	t.Run("returns slots across multiple menu products", func(t *testing.T) {
		slots, err := repos.MenuSlot.GetByMenuProductIDs(ctx, []string{menuA.ID, menuB.ID})
		require.NoError(t, err)
		require.Len(t, slots, 3)

		byMenu := map[string]int{}
		for _, s := range slots {
			byMenu[s.MenuProductID]++
		}
		require.Equal(t, 2, byMenu[menuA.ID])
		require.Equal(t, 1, byMenu[menuB.ID])
	})

	t.Run("loads options edge", func(t *testing.T) {
		slots, err := repos.MenuSlot.GetByMenuProductIDs(ctx, []string{menuA.ID})
		require.NoError(t, err)
		require.Len(t, slots, 2)
		for _, s := range slots {
			require.Len(t, s.Edges.Options, 1, "slot %s should have its option loaded", s.Name)
			require.NotNil(t, s.Edges.Options[0].Edges.OptionProduct)
		}
	})

	t.Run("excludes unrelated menus", func(t *testing.T) {
		slots, err := repos.MenuSlot.GetByMenuProductIDs(ctx, []string{menuA.ID})
		require.NoError(t, err)
		for _, s := range slots {
			require.Equal(t, menuA.ID, s.MenuProductID)
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		slots, err := repos.MenuSlot.GetByMenuProductIDs(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, slots)
	})

	t.Run("unknown ids return no rows", func(t *testing.T) {
		slots, err := repos.MenuSlot.GetByMenuProductIDs(ctx, []string{nanoid.New()})
		require.NoError(t, err)
		require.Empty(t, slots)
	})

	t.Run("matches single-id legacy method", func(t *testing.T) {
		batch, err := repos.MenuSlot.GetByMenuProductIDs(ctx, []string{menuC.ID})
		require.NoError(t, err)
		single, err := repos.MenuSlot.GetByMenuProductID(ctx, menuC.ID)
		require.NoError(t, err)
		require.Len(t, batch, len(single))
	})
}
