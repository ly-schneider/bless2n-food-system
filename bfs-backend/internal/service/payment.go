package service

import (
	"context"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/product"
	"backend/internal/inventory"
	"backend/internal/payrexx"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type PaymentService interface {
	// IsPayrexxEnabled returns true if Payrexx is configured (both instance and API secret set).
	IsPayrexxEnabled() bool
	// PrepareAndCreateOrder validates items and creates a pending order with inventory reservation.
	PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error)
	// CreatePayrexxGateway creates a Payrexx payment gateway for the prepared order.
	CreatePayrexxGateway(ctx context.Context, prep *CheckoutPreparation, successURL, failedURL, cancelURL string) (*payrexx.Gateway, error)
	// MarkOrderPaidByPayrexx marks an order as paid based on Payrexx webhook data.
	MarkOrderPaidByPayrexx(ctx context.Context, orderID uuid.UUID, gatewayID, transactionID int, contactEmail *string) error
	// MarkOrderPaidDev marks an order as paid in dev mode (no Payrexx gateway/transaction IDs).
	MarkOrderPaidDev(ctx context.Context, orderID uuid.UUID) error
	// FindPendingOrderByAttemptID finds a pending order by payment attempt ID.
	FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error)
	// SetOrderAttemptID sets the payment attempt ID on an order.
	SetOrderAttemptID(ctx context.Context, orderID uuid.UUID, attemptID string) error
	// CleanupPendingOrderByID deletes a pending order and releases inventory.
	CleanupPendingOrderByID(ctx context.Context, orderID uuid.UUID) error
	// CleanupOtherPendingOrdersByAttemptID deletes other pending orders with the same attempt ID.
	CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID uuid.UUID) (int64, error)
	// GetPayrexxGateway retrieves a Payrexx gateway by ID.
	GetPayrexxGateway(ctx context.Context, gatewayID int) (*payrexx.Gateway, error)
}

type CreateCheckoutInput struct {
	Items []CheckoutItemInput `json:"items"`
	// CustomerEmail is optional; if user is logged in, their email is used.
	CustomerEmail *string `json:"customerEmail,omitempty"`
	// Origin is the order origin (shop or pos). Defaults to shop if empty.
	Origin order.Origin `json:"-"`
}

type CheckoutItemInput struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
	// Configuration maps slotID -> selected productID for menu items.
	Configuration map[string]string `json:"configuration,omitempty"`
}

type CheckoutPreparation struct {
	OrderID       uuid.UUID
	TotalCents    int64
	LineItems     []payrexx.InvoiceItem
	CustomerEmail *string
	UserID        *string
}

type paymentService struct {
	cfg           config.Config
	payrexxClient *payrexx.Client
	orderRepo     repository.OrderRepository
	orderLineRepo repository.OrderLineRepository
	productRepo   *repository.ProductRepository
	menuSlotRepo  repository.MenuSlotRepository
	inventoryRepo repository.InventoryLedgerRepository
	inventoryHub  *inventory.Hub
}

func NewPaymentService(
	cfg config.Config,
	orderRepo repository.OrderRepository,
	orderLineRepo repository.OrderLineRepository,
	productRepo *repository.ProductRepository,
	menuSlotRepo repository.MenuSlotRepository,
	inventoryRepo repository.InventoryLedgerRepository,
	inventoryHub *inventory.Hub,
) PaymentService {
	var client *payrexx.Client
	if cfg.Payrexx.InstanceName != "" && cfg.Payrexx.APISecret != "" {
		client = payrexx.NewClient(cfg.Payrexx.InstanceName, cfg.Payrexx.APISecret)
	}
	return &paymentService{
		cfg:           cfg,
		payrexxClient: client,
		orderRepo:     orderRepo,
		orderLineRepo: orderLineRepo,
		productRepo:   productRepo,
		menuSlotRepo:  menuSlotRepo,
		inventoryRepo: inventoryRepo,
		inventoryHub:  inventoryHub,
	}
}

func (s *paymentService) IsPayrexxEnabled() bool {
	return s.payrexxClient != nil
}

func (s *paymentService) PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error) {
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("no items")
	}

	// Collect all product IDs
	productIDSet := make(map[uuid.UUID]struct{})
	for _, it := range in.Items {
		pid, err := uuid.Parse(it.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid productId: %s", it.ProductID)
		}
		productIDSet[pid] = struct{}{}
		// Also collect configured child products
		for _, childID := range it.Configuration {
			if childID == "" {
				continue
			}
			cid, err := uuid.Parse(childID)
			if err != nil {
				return nil, fmt.Errorf("invalid configuration productId: %s", childID)
			}
			productIDSet[cid] = struct{}{}
		}
	}

	ids := make([]uuid.UUID, 0, len(productIDSet))
	for id := range productIDSet {
		ids = append(ids, id)
	}

	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load products: %w", err)
	}
	productMap := make(map[uuid.UUID]*ent.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	// Preload menu slots for menu products
	menuProductIDs := make([]uuid.UUID, 0)
	for _, it := range in.Items {
		pid, _ := uuid.Parse(it.ProductID)
		if p, ok := productMap[pid]; ok && p.Type == product.TypeMenu {
			menuProductIDs = append(menuProductIDs, pid)
		}
	}

	slotByID := make(map[uuid.UUID]*ent.MenuSlot)
	allowedBySlot := make(map[uuid.UUID]map[uuid.UUID]struct{})
	for _, menuProdID := range menuProductIDs {
		slots, err := s.menuSlotRepo.GetByMenuProductID(ctx, menuProdID)
		if err != nil {
			return nil, fmt.Errorf("load menu slots: %w", err)
		}
		for _, slot := range slots {
			slotByID[slot.ID] = slot
			allowed := make(map[uuid.UUID]struct{})
			for _, opt := range slot.Edges.Options {
				allowed[opt.OptionProductID] = struct{}{}
			}
			allowedBySlot[slot.ID] = allowed
		}
	}

	// Calculate total and validate
	var totalCents int64
	for _, it := range in.Items {
		pid, _ := uuid.Parse(it.ProductID)
		p, ok := productMap[pid]
		if !ok {
			return nil, fmt.Errorf("unknown product: %s", it.ProductID)
		}
		totalCents += p.PriceCents * int64(it.Quantity)

		// TWINT limit: 5000 CHF per transaction
		if p.PriceCents*int64(it.Quantity) > 500000 {
			return nil, fmt.Errorf("item exceeds TWINT max: %s", p.Name)
		}
	}
	if totalCents > 500000 {
		return nil, fmt.Errorf("order exceeds TWINT max (5000 CHF)")
	}

	// Validate inventory availability
	requiredQuantities := make(map[uuid.UUID]int)
	for _, it := range in.Items {
		pid, _ := uuid.Parse(it.ProductID)
		p := productMap[pid]
		if p.Type == product.TypeSimple {
			requiredQuantities[pid] += it.Quantity
		}
		for _, childIDStr := range it.Configuration {
			if childIDStr == "" {
				continue
			}
			childID, _ := uuid.Parse(childIDStr)
			requiredQuantities[childID] += it.Quantity
		}
	}

	if len(requiredQuantities) > 0 {
		productIDsToCheck := make([]uuid.UUID, 0, len(requiredQuantities))
		for pid := range requiredQuantities {
			productIDsToCheck = append(productIDsToCheck, pid)
		}
		currentStock, err := s.inventoryRepo.GetCurrentStockBatch(ctx, productIDsToCheck)
		if err != nil {
			return nil, fmt.Errorf("check inventory: %w", err)
		}
		for pid, required := range requiredQuantities {
			available := currentStock[pid]
			if available < required {
				pName := "unknown"
				if p, ok := productMap[pid]; ok {
					pName = p.Name
				}
				return nil, fmt.Errorf("insufficient inventory for %s: requested %d, available %d", pName, required, available)
			}
		}
	}

	origin := in.Origin
	if origin == "" {
		origin = order.OriginShop
	}

	// Create the order
	ord, err := s.orderRepo.Create(ctx, totalCents, order.StatusPending, origin, userID, in.CustomerEmail, attemptID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Build order lines
	var orderLines []repository.OrderLineCreateParams
	var inventoryEntries []repository.InventoryLedgerCreateParams

	for _, it := range in.Items {
		pid, _ := uuid.Parse(it.ProductID)
		p := productMap[pid]

		// Determine parent line type
		lt := orderline.LineTypeSimple
		if p.Type == product.TypeMenu {
			lt = orderline.LineTypeBundle
		}

		parentLineID := uuid.Must(uuid.NewV7())
		parentLine := repository.OrderLineCreateParams{
			ID:             &parentLineID,
			OrderID:        ord.ID,
			LineType:       lt,
			ProductID:      p.ID,
			Title:          p.Name,
			Quantity:       it.Quantity,
			UnitPriceCents: p.PriceCents,
		}
		orderLines = append(orderLines, parentLine)

		// Reserve inventory for simple products
		if p.Type == product.TypeSimple && it.Quantity > 0 {
			inventoryEntries = append(inventoryEntries, repository.InventoryLedgerCreateParams{
				ProductID: p.ID,
				Delta:     -it.Quantity,
				Reason:    inventoryledger.ReasonSale,
				OrderID:   &ord.ID,
			})
		}

		// Process menu configurations
		if p.Type == product.TypeMenu && len(it.Configuration) > 0 {
			for slotIDStr, childProdIDStr := range it.Configuration {
				if childProdIDStr == "" {
					continue
				}
				slotID, err := uuid.Parse(slotIDStr)
				if err != nil {
					return nil, fmt.Errorf("invalid menu slot id: %s", slotIDStr)
				}
				slot, ok := slotByID[slotID]
				if !ok || slot.MenuProductID != p.ID {
					return nil, fmt.Errorf("slot does not belong to product: %s", slotIDStr)
				}
				childProdID, err := uuid.Parse(childProdIDStr)
				if err != nil {
					return nil, fmt.Errorf("invalid configured product id: %s", childProdIDStr)
				}
				// Validate allowed
				if allowed := allowedBySlot[slotID]; allowed != nil {
					if _, ok := allowed[childProdID]; !ok {
						return nil, fmt.Errorf("product not allowed in slot")
					}
				}

				childProd := productMap[childProdID]
				slotName := slot.Name
				childLine := repository.OrderLineCreateParams{
					OrderID:        ord.ID,
					LineType:       orderline.LineTypeComponent,
					ProductID:      childProdID,
					Title:          childProd.Name,
					Quantity:       it.Quantity,
					UnitPriceCents: 0,
					ParentLineID:   &parentLineID,
					MenuSlotID:     &slotID,
					MenuSlotName:   &slotName,
				}
				orderLines = append(orderLines, childLine)

				// Reserve inventory for component
				if it.Quantity > 0 {
					inventoryEntries = append(inventoryEntries, repository.InventoryLedgerCreateParams{
						ProductID: childProdID,
						Delta:     -it.Quantity,
						Reason:    inventoryledger.ReasonSale,
						OrderID:   &ord.ID,
					})
				}
			}
		}
	}

	// Insert order lines
	if _, err := s.orderLineRepo.CreateBatch(ctx, orderLines); err != nil {
		return nil, fmt.Errorf("insert order lines: %w", err)
	}

	// Reserve inventory
	if len(inventoryEntries) > 0 {
		if _, err := s.inventoryRepo.CreateMany(ctx, inventoryEntries); err != nil {
			return nil, fmt.Errorf("reserve inventory: %w", err)
		}
		s.publishInventoryUpdates(ctx, inventoryEntries)
	}

	// Prepare Payrexx line items from the params (we have all the data we need)
	lineItems := make([]payrexx.InvoiceItem, 0)
	for _, line := range orderLines {
		if line.ParentLineID == nil && line.UnitPriceCents > 0 && line.Quantity > 0 {
			lineItems = append(lineItems, payrexx.InvoiceItem{
				Name:     line.Title,
				Quantity: line.Quantity,
				Amount:   int(line.UnitPriceCents),
			})
		}
	}

	return &CheckoutPreparation{
		OrderID:       ord.ID,
		TotalCents:    totalCents,
		LineItems:     lineItems,
		CustomerEmail: in.CustomerEmail,
		UserID:        userID,
	}, nil
}

func (s *paymentService) CreatePayrexxGateway(ctx context.Context, prep *CheckoutPreparation, successURL, failedURL, cancelURL string) (*payrexx.Gateway, error) {
	if s.payrexxClient == nil {
		return nil, fmt.Errorf("payrexx client not configured")
	}

	gateway, err := s.payrexxClient.CreateGateway(payrexx.CreateGatewayParams{
		Amount:             int(prep.TotalCents),
		Currency:           "CHF",
		ReferenceID:        prep.OrderID.String(),
		SuccessRedirectURL: successURL,
		FailedRedirectURL:  failedURL,
		CancelRedirectURL:  cancelURL,
		PaymentMeans:       []string{"twint"},
		InvoiceItems:       prep.LineItems,
		CustomerEmail:      safeStr(prep.CustomerEmail),
		Purpose:            "BlessThun Food Order",
		ValidityMinutes:    15,
	})
	if err != nil {
		return nil, fmt.Errorf("create payrexx gateway: %w", err)
	}

	// Update order with gateway ID - get current order first to preserve all fields
	ord, err := s.orderRepo.GetByID(ctx, prep.OrderID)
	if err != nil {
		return nil, err
	}
	gwID := gateway.ID
	if _, err := s.orderRepo.Update(ctx, ord.ID, ord.TotalCents, ord.Status, ord.Origin, ord.CustomerID, ord.ContactEmail, ord.PaymentAttemptID, &gwID, ord.PayrexxTransactionID); err != nil {
		return nil, fmt.Errorf("update order with gateway: %w", err)
	}

	return gateway, nil
}

func (s *paymentService) MarkOrderPaidByPayrexx(ctx context.Context, orderID uuid.UUID, gatewayID, transactionID int, contactEmail *string) error {
	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Merge contact email: use provided one if non-empty, otherwise keep existing
	ce := ord.ContactEmail
	if contactEmail != nil && *contactEmail != "" {
		ce = contactEmail
	}

	_, err = s.orderRepo.Update(ctx, ord.ID, ord.TotalCents, order.StatusPaid, ord.Origin, ord.CustomerID, ce, ord.PaymentAttemptID, &gatewayID, &transactionID)
	return err
}

func (s *paymentService) MarkOrderPaidDev(ctx context.Context, orderID uuid.UUID) error {
	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if ord.Status != order.StatusPending {
		return fmt.Errorf("order is not pending")
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, order.StatusPaid)
}

func (s *paymentService) FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error) {
	if attemptID == "" {
		return nil, fmt.Errorf("missing attempt id")
	}
	return s.orderRepo.FindPendingByAttemptID(ctx, attemptID)
}

func (s *paymentService) SetOrderAttemptID(ctx context.Context, orderID uuid.UUID, attemptID string) error {
	return s.orderRepo.SetPaymentAttemptID(ctx, orderID, attemptID)
}

func (s *paymentService) CleanupPendingOrderByID(ctx context.Context, orderID uuid.UUID) error {
	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return nil // Only cleanup pending orders
	}

	// Load order lines to release inventory
	lines, err := s.orderLineRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// Release inventory
	var releaseEntries []repository.InventoryLedgerCreateParams
	for _, line := range lines {
		if line.Quantity > 0 {
			// Only release for simple products or components (not menu bundles themselves)
			if line.LineType == orderline.LineTypeSimple || line.LineType == orderline.LineTypeComponent {
				releaseEntries = append(releaseEntries, repository.InventoryLedgerCreateParams{
					ProductID: line.ProductID,
					Delta:     line.Quantity, // Positive to add back
					Reason:    inventoryledger.ReasonCorrection,
					OrderID:   &orderID,
				})
			}
		}
	}
	if len(releaseEntries) > 0 {
		_, _ = s.inventoryRepo.CreateMany(ctx, releaseEntries)
		s.publishInventoryUpdates(ctx, releaseEntries)
	}

	// Delete order (cascade deletes order lines)
	_, err = s.orderRepo.DeleteIfPending(ctx, orderID)
	return err
}

func (s *paymentService) CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID uuid.UUID) (int64, error) {
	return s.orderRepo.DeletePendingByAttemptIDExcept(ctx, attemptID, keepOrderID)
}

func (s *paymentService) GetPayrexxGateway(ctx context.Context, gatewayID int) (*payrexx.Gateway, error) {
	if s.payrexxClient == nil {
		return nil, fmt.Errorf("payrexx client not configured")
	}
	return s.payrexxClient.GetGateway(gatewayID)
}

func safeStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *paymentService) publishInventoryUpdates(ctx context.Context, entries []repository.InventoryLedgerCreateParams) {
	if s.inventoryHub == nil {
		return
	}
	productIDs := make([]uuid.UUID, 0, len(entries))
	deltaByProduct := make(map[uuid.UUID]int)
	for _, entry := range entries {
		if _, seen := deltaByProduct[entry.ProductID]; !seen {
			productIDs = append(productIDs, entry.ProductID)
		}
		deltaByProduct[entry.ProductID] += entry.Delta
	}
	stocks, err := s.inventoryRepo.GetCurrentStockBatch(ctx, productIDs)
	if err != nil {
		return
	}
	now := time.Now()
	for _, productID := range productIDs {
		s.inventoryHub.Publish(inventory.Update{
			ProductID: productID,
			NewStock:  stocks[productID],
			Delta:     deltaByProduct[productID],
			Timestamp: now,
		})
	}
}
