package service

import (
	"context"
	"fmt"
	"time"

	"backend/internal/auth"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/product"
	nanoid "backend/internal/id"
	"backend/internal/inventory"
	"backend/internal/repository"
)

type ProductService interface {
	ListProducts(ctx context.Context, categoryID *string, limit, offset int) ([]*ent.Product, error)
	GetByID(ctx context.Context, id string) (*ent.Product, error)
	GetByIDWithRelations(ctx context.Context, id string) (*ent.Product, error)
	GetAll(ctx context.Context) ([]*ent.Product, error)
	GetByCategory(ctx context.Context, categoryID string) ([]*ent.Product, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.Product, error)
	Create(ctx context.Context, categoryID string, productType product.Type, name string, priceCents int64, isActive bool, image *string, description *string, jetonID *string) (*ent.Product, error)
	Update(ctx context.Context, id, categoryID string, productType product.Type, name string, priceCents int64, isActive bool, image *string, description *string, jetonID *string) (*ent.Product, error)
	Delete(ctx context.Context, id string) error
	CountActiveWithoutJeton(ctx context.Context) (int64, error)
	CountByJetonIDs(ctx context.Context, ids []string) (map[string]int64, error)
	UpdateJeton(ctx context.Context, id string, jetonID *string) error

	// Inventory
	GetStock(ctx context.Context, id string) (int64, error)
	GetStockBatch(ctx context.Context, ids []string) (map[string]int, error)
	AdjustStock(ctx context.Context, id string, delta int64, reason string) error
	ListInventoryHistory(ctx context.Context, productID string, limit, offset int) ([]*ent.InventoryLedger, error)

	// Menus
	GetMenus(ctx context.Context) ([]*ent.Product, error)

	// Menu slots
	CreateMenuSlot(ctx context.Context, menuID string, name string) (*ent.MenuSlot, error)
	UpdateMenuSlot(ctx context.Context, menuID, slotID string, name string) (*ent.MenuSlot, error)
	DeleteMenuSlot(ctx context.Context, menuID, slotID string) error
	ReorderMenuSlots(ctx context.Context, menuID string, positions map[string]int) error

	// Menu slot options
	AddSlotOption(ctx context.Context, menuID, slotID string, productID string) (*ent.MenuSlotOption, error)
	RemoveSlotOption(ctx context.Context, menuID, slotID, optionProductID string) error
}

type productService struct {
	productRepo        *repository.ProductRepository
	categoryRepo       repository.CategoryRepository
	menuSlotRepo       repository.MenuSlotRepository
	menuSlotOptionRepo repository.MenuSlotOptionRepository
	inventoryRepo      repository.InventoryLedgerRepository
	jetonRepo          repository.JetonRepository
	inventoryHub       *inventory.Hub
	cache              *catalogCache
}

func NewProductService(
	productRepo *repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	menuSlotRepo repository.MenuSlotRepository,
	menuSlotOptionRepo repository.MenuSlotOptionRepository,
	inventoryRepo repository.InventoryLedgerRepository,
	jetonRepo repository.JetonRepository,
	inventoryHub *inventory.Hub,
) ProductService {
	return &productService{
		productRepo:        productRepo,
		categoryRepo:       categoryRepo,
		menuSlotRepo:       menuSlotRepo,
		menuSlotOptionRepo: menuSlotOptionRepo,
		inventoryRepo:      inventoryRepo,
		jetonRepo:          jetonRepo,
		inventoryHub:       inventoryHub,
		cache:              newCatalogCache(),
	}
}

func (s *productService) ListProducts(ctx context.Context, categoryID *string, limit, offset int) ([]*ent.Product, error) {
	if categoryID != nil {
		catID := *categoryID
		if !nanoid.Valid(catID) {
			return nil, fmt.Errorf("invalid category id")
		}
		return s.cache.load(ctx, catalogKeyByCategory(catID), func(ctx context.Context) ([]*ent.Product, error) {
			return s.productRepo.GetByCategory(ctx, catID)
		})
	}
	return s.cache.load(ctx, catalogKeyAll(), func(ctx context.Context) ([]*ent.Product, error) {
		return s.productRepo.GetAll(ctx)
	})
}

func (s *productService) GetByID(ctx context.Context, id string) (*ent.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *productService) GetByIDWithRelations(ctx context.Context, id string) (*ent.Product, error) {
	return s.productRepo.GetByIDWithRelations(ctx, id)
}

func (s *productService) GetAll(ctx context.Context) ([]*ent.Product, error) {
	return s.cache.load(ctx, catalogKeyAll(), func(ctx context.Context) ([]*ent.Product, error) {
		return s.productRepo.GetAll(ctx)
	})
}

func (s *productService) GetByCategory(ctx context.Context, categoryID string) ([]*ent.Product, error) {
	return s.cache.load(ctx, catalogKeyByCategory(categoryID), func(ctx context.Context) ([]*ent.Product, error) {
		return s.productRepo.GetByCategory(ctx, categoryID)
	})
}

func (s *productService) GetByIDs(ctx context.Context, ids []string) ([]*ent.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	all, err := s.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*ent.Product, len(all))
	for _, p := range all {
		byID[p.ID] = p
	}
	result := make([]*ent.Product, 0, len(ids))
	missing := make([]string, 0)
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			result = append(result, p)
			continue
		}
		missing = append(missing, id)
	}
	if len(missing) > 0 {
		fetched, err := s.productRepo.GetByIDs(ctx, missing)
		if err != nil {
			return nil, err
		}
		result = append(result, fetched...)
	}
	return result, nil
}

func (s *productService) Create(ctx context.Context, categoryID string, productType product.Type, name string, priceCents int64, isActive bool, image *string, description *string, jetonID *string) (*ent.Product, error) {
	created, err := s.productRepo.Create(ctx, categoryID, productType, name, priceCents, isActive, image, description, jetonID)
	if err == nil {
		s.cache.invalidate()
	}
	return created, err
}

func (s *productService) Update(ctx context.Context, id, categoryID string, productType product.Type, name string, priceCents int64, isActive bool, image *string, description *string, jetonID *string) (*ent.Product, error) {
	updated, err := s.productRepo.Update(ctx, id, categoryID, productType, name, priceCents, isActive, image, description, jetonID)
	if err == nil {
		s.cache.invalidate()
	}
	return updated, err
}

func (s *productService) Delete(ctx context.Context, id string) error {
	slots, _ := s.menuSlotRepo.GetByMenuProductID(ctx, id)
	for _, slot := range slots {
		_ = s.menuSlotOptionRepo.DeleteByMenuSlotID(ctx, slot.ID)
	}
	if len(slots) > 0 {
		_ = s.menuSlotRepo.DeleteByMenuProductID(ctx, id)
	}
	err := s.productRepo.Delete(ctx, id)
	if err == nil {
		s.cache.invalidate()
	}
	return err
}

func (s *productService) CountActiveWithoutJeton(ctx context.Context) (int64, error) {
	return s.productRepo.CountActiveWithoutJeton(ctx)
}

func (s *productService) CountByJetonIDs(ctx context.Context, ids []string) (map[string]int64, error) {
	return s.productRepo.CountByJetonIDs(ctx, ids)
}

func (s *productService) UpdateJeton(ctx context.Context, id string, jetonID *string) error {
	err := s.productRepo.UpdateJeton(ctx, id, jetonID)
	if err == nil {
		s.cache.invalidate()
	}
	return err
}

// ---------------------------------------------------------------------------
// Inventory
// ---------------------------------------------------------------------------

func (s *productService) GetStock(ctx context.Context, id string) (int64, error) {
	stock, err := s.inventoryRepo.GetCurrentStock(ctx, id)
	if err != nil {
		return 0, err
	}
	return int64(stock), nil
}

func (s *productService) GetStockBatch(ctx context.Context, ids []string) (map[string]int, error) {
	return s.inventoryRepo.GetCurrentStockBatch(ctx, ids)
}

func (s *productService) ListInventoryHistory(ctx context.Context, productID string, limit, offset int) ([]*ent.InventoryLedger, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.inventoryRepo.GetByProductIDWithPagination(ctx, productID, limit, offset)
}

func (s *productService) AdjustStock(ctx context.Context, id string, delta int64, reason string) error {
	r := inventoryledger.ReasonManualAdjust
	switch reason {
	case string(inventoryledger.ReasonOpeningBalance):
		r = inventoryledger.ReasonOpeningBalance
	case string(inventoryledger.ReasonSale):
		r = inventoryledger.ReasonSale
	case string(inventoryledger.ReasonRefund):
		r = inventoryledger.ReasonRefund
	case string(inventoryledger.ReasonCorrection):
		r = inventoryledger.ReasonCorrection
	case string(inventoryledger.ReasonManualAdjust):
		r = inventoryledger.ReasonManualAdjust
	}
	var createdBy *string
	if uid, ok := auth.GetUserID(ctx); ok {
		createdBy = &uid
	}
	_, err := s.inventoryRepo.Create(ctx, id, int(delta), r, nil, nil, nil, createdBy)
	if err != nil {
		return err
	}
	if s.inventoryHub != nil {
		newStock, stockErr := s.inventoryRepo.GetCurrentStock(ctx, id)
		if stockErr == nil {
			s.inventoryHub.Publish(inventory.Update{
				ProductID: id,
				NewStock:  newStock,
				Delta:     int(delta),
				Timestamp: time.Now(),
			})
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Menus
// ---------------------------------------------------------------------------

func (s *productService) GetMenus(ctx context.Context) ([]*ent.Product, error) {
	menus, _, err := s.productRepo.GetMenus(ctx, nil, nil, 1000, 0)
	return menus, err
}

// ---------------------------------------------------------------------------
// Menu Slots
// ---------------------------------------------------------------------------

func (s *productService) CreateMenuSlot(ctx context.Context, menuID string, name string) (*ent.MenuSlot, error) {
	existing, err := s.menuSlotRepo.GetByMenuProductID(ctx, menuID)
	if err != nil {
		return nil, err
	}
	seq := 0
	for _, slot := range existing {
		if slot.Sequence >= seq {
			seq = slot.Sequence + 1
		}
	}
	created, err := s.menuSlotRepo.Create(ctx, menuID, name, seq)
	if err == nil {
		s.cache.invalidate()
	}
	return created, err
}

func (s *productService) UpdateMenuSlot(ctx context.Context, menuID, slotID string, name string) (*ent.MenuSlot, error) {
	slot, err := s.menuSlotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = slot.Name
	}
	updated, err := s.menuSlotRepo.Update(ctx, slotID, menuID, name, slot.Sequence)
	if err == nil {
		s.cache.invalidate()
	}
	return updated, err
}

func (s *productService) DeleteMenuSlot(ctx context.Context, menuID, slotID string) error {
	// Verify the slot belongs to this menu.
	slot, err := s.menuSlotRepo.GetByID(ctx, slotID)
	if err != nil {
		return err
	}
	if slot.MenuProductID != menuID {
		return fmt.Errorf("slot does not belong to menu")
	}
	// Delete options first, then the slot.
	if err := s.menuSlotOptionRepo.DeleteByMenuSlotID(ctx, slotID); err != nil {
		return err
	}
	if err := s.menuSlotRepo.Delete(ctx, slotID); err != nil {
		return err
	}
	// Re-normalize remaining slot sequences to eliminate gaps.
	remaining, err := s.menuSlotRepo.GetByMenuProductID(ctx, menuID)
	if err != nil {
		return err
	}
	for i, sl := range remaining {
		if sl.Sequence != i {
			if _, err := s.menuSlotRepo.Update(ctx, sl.ID, menuID, sl.Name, i); err != nil {
				return err
			}
		}
	}
	s.cache.invalidate()
	return nil
}

func (s *productService) ReorderMenuSlots(ctx context.Context, menuID string, positions map[string]int) error {
	slots, err := s.menuSlotRepo.GetByMenuProductID(ctx, menuID)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		if newSeq, ok := positions[slot.ID]; ok {
			if _, err := s.menuSlotRepo.Update(ctx, slot.ID, menuID, slot.Name, newSeq); err != nil {
				return err
			}
		}
	}
	s.cache.invalidate()
	return nil
}

func (s *productService) AddSlotOption(ctx context.Context, menuID, slotID string, productID string) (*ent.MenuSlotOption, error) {
	opt, err := s.menuSlotOptionRepo.Create(ctx, slotID, productID)
	if err == nil {
		s.cache.invalidate()
	}
	return opt, err
}

func (s *productService) RemoveSlotOption(ctx context.Context, menuID, slotID, optionProductID string) error {
	err := s.menuSlotOptionRepo.Delete(ctx, slotID, optionProductID)
	if err == nil {
		s.cache.invalidate()
	}
	return err
}
