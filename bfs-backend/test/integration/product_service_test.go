package integration

import (
	"context"
	"testing"

	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/product"
	"backend/internal/service"

	"github.com/stretchr/testify/require"
)

func TestProductService_ListProducts(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewProductService(
		repos.Product,
		repos.Category,
		repos.MenuSlot,
		repos.MenuSlotOption,
		repos.Inventory,
		repos.Jeton,
		nil,
	)
	ctx := context.Background()

	// Setup test data
	category := fixtures.CreateCategory("Drinks", 1, true)
	jeton := fixtures.CreateJeton("Red", "#EF4444")

	fixtures.CreateProduct("Cola", category.ID, 350, product.TypeSimple, &jeton.ID)
	fixtures.CreateProduct("Sprite", category.ID, 350, product.TypeSimple, &jeton.ID)

	// Add inventory
	fixtures.AddInventory(fixtures.CreateProduct("InvTest", category.ID, 100, product.TypeSimple, nil).ID, 100, inventoryledger.ReasonOpeningBalance)

	t.Run("List all products", func(t *testing.T) {
		products, err := svc.ListProducts(ctx, nil, 50, 0)
		require.NoError(t, err)
		require.Len(t, products, 3)
	})

	t.Run("List products with category filter", func(t *testing.T) {
		// Create another category with products
		otherCat := fixtures.CreateCategory("Food", 2, true)
		fixtures.CreateProduct("Burger", otherCat.ID, 1200, product.TypeSimple, nil)

		catIDStr := category.ID.String()
		products, err := svc.ListProducts(ctx, &catIDStr, 50, 0)
		require.NoError(t, err)
		// Only products from the Drinks category
		for _, p := range products {
			require.Equal(t, category.ID, p.CategoryID)
		}
	})

	t.Run("List products eager-loads jeton edge", func(t *testing.T) {
		products, err := svc.ListProducts(ctx, nil, 50, 0)
		require.NoError(t, err)

		var foundJeton bool
		for _, p := range products {
			if p.JetonID != nil && *p.JetonID == jeton.ID {
				foundJeton = true
				require.NotNil(t, p.Edges.Jeton)
				require.Equal(t, "Red", p.Edges.Jeton.Name)
				require.Equal(t, "#EF4444", p.Edges.Jeton.Color)
			}
		}
		require.True(t, foundJeton, "Expected to find product with jeton")
	})

	t.Run("List with invalid category ID returns error", func(t *testing.T) {
		invalidID := "not-a-uuid"
		_, err := svc.ListProducts(ctx, &invalidID, 50, 0)
		require.Error(t, err)
	})
}

func TestProductService_MenuProducts(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)

	svc := service.NewProductService(
		repos.Product,
		repos.Category,
		repos.MenuSlot,
		repos.MenuSlotOption,
		repos.Inventory,
		repos.Jeton,
		nil,
	)
	ctx := context.Background()

	// Setup test data
	category := fixtures.CreateCategory("Menus", 1, true)

	// Create simple products for menu options
	optionCat := fixtures.CreateCategory("Options", 2, true)
	cola := fixtures.CreateProduct("Cola", optionCat.ID, 350, product.TypeSimple, nil)
	sprite := fixtures.CreateProduct("Sprite", optionCat.ID, 350, product.TypeSimple, nil)
	fries := fixtures.CreateProduct("Fries", optionCat.ID, 400, product.TypeSimple, nil)

	fixtures.AddInventory(cola.ID, 50, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(sprite.ID, 50, inventoryledger.ReasonOpeningBalance)
	fixtures.AddInventory(fries.ID, 30, inventoryledger.ReasonOpeningBalance)

	// Create menu product
	menu := fixtures.CreateProduct("Combo Menu", category.ID, 1500, product.TypeMenu, nil)

	// Create menu slots
	drinkSlot := fixtures.CreateMenuSlot(menu.ID, "Drink", 1)
	sideSlot := fixtures.CreateMenuSlot(menu.ID, "Side", 2)

	// Add options to slots
	fixtures.CreateMenuSlotOption(drinkSlot.ID, cola.ID)
	fixtures.CreateMenuSlotOption(drinkSlot.ID, sprite.ID)
	fixtures.CreateMenuSlotOption(sideSlot.ID, fries.ID)

	t.Run("GetMenus returns menu products", func(t *testing.T) {
		menus, err := svc.GetMenus(ctx)
		require.NoError(t, err)
		require.Len(t, menus, 1)
		require.Equal(t, "Combo Menu", menus[0].Name)
		require.Equal(t, product.TypeMenu, menus[0].Type)
	})

	t.Run("Menu slot operations", func(t *testing.T) {
		// Create a new slot
		newSlot, err := svc.CreateMenuSlot(ctx, menu.ID, "Dessert")
		require.NoError(t, err)
		require.Equal(t, "Dessert", newSlot.Name)

		// Update the slot
		updated, err := svc.UpdateMenuSlot(ctx, menu.ID, newSlot.ID, "Sweet Treat")
		require.NoError(t, err)
		require.Equal(t, "Sweet Treat", updated.Name)

		// Delete the slot
		err = svc.DeleteMenuSlot(ctx, menu.ID, newSlot.ID)
		require.NoError(t, err)
	})
}

func TestProductService_EmptyResults(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	repos := NewRepositories(tdb.Client)

	svc := service.NewProductService(
		repos.Product,
		repos.Category,
		repos.MenuSlot,
		repos.MenuSlotOption,
		repos.Inventory,
		repos.Jeton,
		nil,
	)
	ctx := context.Background()

	t.Run("List products returns empty for no products", func(t *testing.T) {
		products, err := svc.ListProducts(ctx, nil, 50, 0)
		require.NoError(t, err)
		require.Empty(t, products)
	})
}
