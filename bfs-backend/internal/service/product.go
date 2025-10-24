package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductService interface {
	ListProducts(ctx context.Context, categoryID *string, limit int, offset int) (*domain.ListResponse[domain.ProductDTO], error)
}

type productService struct {
	productRepo      repository.ProductRepository
	categoryRepo     repository.CategoryRepository
	menuSlotRepo     repository.MenuSlotRepository
	menuSlotItemRepo repository.MenuSlotItemRepository
	inventoryRepo    repository.InventoryLedgerRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	menuSlotRepo repository.MenuSlotRepository,
	menuSlotItemRepo repository.MenuSlotItemRepository,
	inventoryRepo repository.InventoryLedgerRepository,
) ProductService {
	return &productService{
		productRepo:      productRepo,
		categoryRepo:     categoryRepo,
		menuSlotRepo:     menuSlotRepo,
		menuSlotItemRepo: menuSlotItemRepo,
		inventoryRepo:    inventoryRepo,
	}
}

func (s *productService) ListProducts(
	ctx context.Context,
	categoryID *string,
	limit int,
	offset int,
) (*domain.ListResponse[domain.ProductDTO], error) {

	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}

	var (
		products []*domain.Product
		err      error
	)

	if categoryID != nil {
		catID, convErr := bson.ObjectIDFromHex(*categoryID)
		if convErr != nil {
			return nil, errors.New("invalid category ID format")
		}

		cat, getErr := s.categoryRepo.GetByID(ctx, catID)
		if getErr != nil {
			return nil, fmt.Errorf("failed to check category: %w", getErr)
		}
		if cat == nil {
			return nil, errors.New("category not found")
		}

		products, err = s.productRepo.GetByCategoryID(ctx, cat.ID, limit, offset)
	} else {
		products, err = s.productRepo.GetAll(ctx, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	if len(products) == 0 {
		return &domain.ListResponse[domain.ProductDTO]{Items: []domain.ProductDTO{}, Count: 0}, nil
	}

	baseCatIDs := make(map[bson.ObjectID]struct{}, len(products))
	menuIDs := make([]bson.ObjectID, 0, len(products))
	for _, p := range products {
		baseCatIDs[p.CategoryID] = struct{}{}
		if p.Type == domain.ProductTypeMenu {
			menuIDs = append(menuIDs, p.ID)
		}
	}

	slotsByMenu := make(map[bson.ObjectID][]*domain.MenuSlot, len(menuIDs))
	slotIDs := make([]bson.ObjectID, 0, 16)
	if len(menuIDs) > 0 {
		slots, err := s.menuSlotRepo.FindByProductIDs(ctx, menuIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load menu slots: %w", err)
		}
		if len(slots) > 0 {
			slotIDs = make([]bson.ObjectID, 0, len(slots))
			for _, sl := range slots {
				slotsByMenu[sl.ProductID] = append(slotsByMenu[sl.ProductID], sl)
				slotIDs = append(slotIDs, sl.ID)
			}
			for id := range slotsByMenu {
				sort.Slice(slotsByMenu[id], func(i, j int) bool {
					return slotsByMenu[id][i].Sequence < slotsByMenu[id][j].Sequence
				})
			}
		}
	}

	itemsBySlot := make(map[bson.ObjectID][]*domain.MenuSlotItem, len(slotIDs))
	optionProdIDs := make(map[bson.ObjectID]struct{})
	if len(slotIDs) > 0 {
		items, err := s.menuSlotItemRepo.FindByMenuSlotIDs(ctx, slotIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load slot items: %w", err)
		}
		for _, it := range items {
			itemsBySlot[it.MenuSlotID] = append(itemsBySlot[it.MenuSlotID], it)
			optionProdIDs[it.ProductID] = struct{}{}
		}
	}

	optionProductsByID := make(map[bson.ObjectID]*domain.Product, len(optionProdIDs))
	optionCatIDs := make(map[bson.ObjectID]struct{}, len(optionProdIDs))
	if len(optionProdIDs) > 0 {
		ids := make([]bson.ObjectID, 0, len(optionProdIDs))
		for id := range optionProdIDs {
			ids = append(ids, id)
		}
		optProducts, err := s.productRepo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("failed to load option products: %w", err)
		}
		for _, op := range optProducts {
			optionProductsByID[op.ID] = op
			optionCatIDs[op.CategoryID] = struct{}{}
		}
	}

	// Compute availability for option products (simple items used in menus)
	optionSimpleIDs := make([]bson.ObjectID, 0)
	for _, op := range optionProductsByID {
		if op != nil && op.Type == domain.ProductTypeSimple {
			optionSimpleIDs = append(optionSimpleIDs, op.ID)
		}
	}
	optionStockByID := map[bson.ObjectID]int64{}
	if len(optionSimpleIDs) > 0 && s.inventoryRepo != nil {
		if sums, err := s.inventoryRepo.SumByProductIDs(ctx, optionSimpleIDs); err == nil {
			optionStockByID = sums
		}
	}

	allCatIDs := make([]bson.ObjectID, 0, len(baseCatIDs)+len(optionCatIDs))
	for id := range baseCatIDs {
		allCatIDs = append(allCatIDs, id)
	}
	for id := range optionCatIDs {
		if _, seen := baseCatIDs[id]; !seen {
			allCatIDs = append(allCatIDs, id)
		}
	}

	categories, err := s.categoryRepo.GetByIDs(ctx, allCatIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get categories: %w", err)
	}
	catByID := make(map[bson.ObjectID]*domain.Category, len(categories))
	for _, c := range categories {
		catByID[c.ID] = c
	}

	catDTOByID := make(map[bson.ObjectID]domain.CategoryDTO, len(catByID))
	for id, c := range catByID {
		catDTOByID[id] = domain.CategoryDTO{ID: c.ID.Hex(), Name: c.Name, IsActive: c.IsActive, Position: c.Position}
	}

	out := make([]domain.ProductDTO, 0, len(products))

	toCatDTO := func(id bson.ObjectID) domain.CategoryDTO {
		if dto, ok := catDTOByID[id]; ok {
			return dto
		}
		return domain.CategoryDTO{ID: id.Hex(), Name: "", IsActive: false, Position: 0}
	}

	// Precompute availability for simple products
	simpleIDs := make([]bson.ObjectID, 0)
	for _, p := range products {
		if p.Type == domain.ProductTypeSimple {
			simpleIDs = append(simpleIDs, p.ID)
		}
	}
	stockByID := map[bson.ObjectID]int64{}
	if len(simpleIDs) > 0 && s.inventoryRepo != nil {
		if sums, err := s.inventoryRepo.SumByProductIDs(ctx, simpleIDs); err == nil {
			stockByID = sums
		}
	}

	for _, p := range products {
		summary := domain.ProductSummaryDTO{
			ID:         p.ID.Hex(),
			Type:       domain.ProductType(p.Type),
			Name:       p.Name,
			Image:      p.Image,
			PriceCents: domain.Cents(p.PriceCents),
			IsActive:   p.IsActive,
			Category:   toCatDTO(p.CategoryID),
		}
		// Attach availability for simple products
		if p.Type == domain.ProductTypeSimple {
			qty64 := stockByID[p.ID]
			qty := int(qty64)
			available := qty > 0
			low := available && qty <= 10
			summary.AvailableQuantity = &qty
			summary.IsAvailable = &available
			summary.IsLowStock = &low
		}

		dto := domain.ProductDTO{ProductSummaryDTO: summary}

		if p.Type == domain.ProductTypeMenu {
			if slots := slotsByMenu[p.ID]; len(slots) > 0 {
				menu := domain.MenuDTO{Slots: make([]domain.MenuSlotDTO, 0, len(slots))}
				for _, sl := range slots {
					slotDTO := domain.MenuSlotDTO{
						ID:       sl.ID.Hex(),
						Name:     sl.Name,
						Sequence: sl.Sequence,
					}
					if items := itemsBySlot[sl.ID]; len(items) > 0 {
						slotDTO.MenuSlotItem = make([]domain.ProductSummaryDTO, 0, len(items))
						for _, it := range items {
							if op := optionProductsByID[it.ProductID]; op != nil {
								// Base summary
								sum := domain.ProductSummaryDTO{
									ID:         op.ID.Hex(),
									Type:       domain.ProductType(op.Type),
									Name:       op.Name,
									Image:      op.Image,
									PriceCents: domain.Cents(op.PriceCents),
									IsActive:   op.IsActive,
									Category:   toCatDTO(op.CategoryID),
								}
								// Attach stock for simple option products
								if op.Type == domain.ProductTypeSimple {
									qty64 := optionStockByID[op.ID]
									qty := int(qty64)
									available := qty > 0
									low := available && qty <= 10
									sum.AvailableQuantity = &qty
									sum.IsAvailable = &available
									sum.IsLowStock = &low
								}
								slotDTO.MenuSlotItem = append(slotDTO.MenuSlotItem, sum)
							}
						}
					}
					menu.Slots = append(menu.Slots, slotDTO)
				}
				dto.Menu = &menu
			}
		}

		out = append(out, dto)
	}

	return &domain.ListResponse[domain.ProductDTO]{Items: out, Count: len(out)}, nil
}
