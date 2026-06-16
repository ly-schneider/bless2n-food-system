package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/inventoryledger"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/orderpayment"
	"backend/internal/generated/ent/product"
	nanoid "backend/internal/id"
	"backend/internal/inventory"
	"backend/internal/payrexx"
	"backend/internal/qrsign"
	"backend/internal/repository"
	"backend/internal/trace"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type PaymentService interface {
	// IsPayrexxEnabled returns true if Payrexx is configured (both instance and API secret set).
	IsPayrexxEnabled() bool
	// PrepareAndCreateOrder validates items and creates a pending order with inventory reservation.
	PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error)
	// CreatePayrexxGateway creates a Payrexx payment gateway for the prepared order.
	CreatePayrexxGateway(ctx context.Context, prep *CheckoutPreparation, successURL, failedURL, cancelURL string) (*payrexx.Gateway, error)
	// MarkOrderPaidByPayrexx marks an order as paid based on Payrexx webhook data.
	MarkOrderPaidByPayrexx(ctx context.Context, orderID string, gatewayID, transactionID int, contactEmail *string) error
	// MarkOrderPaidDev marks an order as paid in dev mode (no Payrexx gateway/transaction IDs).
	MarkOrderPaidDev(ctx context.Context, orderID string) error
	// FindPendingOrderByAttemptID finds a pending order by payment attempt ID.
	FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error)
	// SetOrderAttemptID sets the payment attempt ID on an order.
	SetOrderAttemptID(ctx context.Context, orderID string, attemptID string) error
	// CleanupPendingOrderByID deletes a pending order and releases inventory.
	CleanupPendingOrderByID(ctx context.Context, orderID string) error
	// CleanupOtherPendingOrdersByAttemptID deletes other pending orders with the same attempt ID.
	CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID string) (int64, error)
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
	OrderID       string
	TotalCents    int64
	LineItems     []payrexx.InvoiceItem
	CustomerEmail *string
	UserID        *string
	Order         *ent.Order
	QRPayload     string
}

type paymentService struct {
	cfg              config.Config
	payrexxClient    *payrexx.Client
	orderRepo        repository.OrderRepository
	orderLineRepo    repository.OrderLineRepository
	orderPaymentRepo repository.OrderPaymentRepository
	products         ProductService
	menuSlotRepo     repository.MenuSlotRepository
	inventoryRepo    repository.InventoryLedgerRepository
	inventoryHub     *inventory.Hub
	emailService     EmailService
	qrKeys           QRKeyService
	logger           *zap.Logger
}

func NewPaymentService(
	cfg config.Config,
	orderRepo repository.OrderRepository,
	orderLineRepo repository.OrderLineRepository,
	orderPaymentRepo repository.OrderPaymentRepository,
	products ProductService,
	menuSlotRepo repository.MenuSlotRepository,
	inventoryRepo repository.InventoryLedgerRepository,
	inventoryHub *inventory.Hub,
	emailService EmailService,
	qrKeys QRKeyService,
	logger *zap.Logger,
) PaymentService {
	var client *payrexx.Client
	if cfg.Payrexx.InstanceName != "" && cfg.Payrexx.APISecret != "" {
		client = payrexx.NewClient(cfg.Payrexx.InstanceName, cfg.Payrexx.APISecret)
	}
	return &paymentService{
		cfg:              cfg,
		payrexxClient:    client,
		orderRepo:        orderRepo,
		orderLineRepo:    orderLineRepo,
		orderPaymentRepo: orderPaymentRepo,
		products:         products,
		menuSlotRepo:     menuSlotRepo,
		inventoryRepo:    inventoryRepo,
		inventoryHub:     inventoryHub,
		emailService:     emailService,
		qrKeys:           qrKeys,
		logger:           logger,
	}
}

func (s *paymentService) IsPayrexxEnabled() bool {
	return s.payrexxClient != nil
}

func (s *paymentService) PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error) {
	ctx, finish := trace.StartSpan(ctx, "service", "payment.prepare_checkout")
	defer finish()
	trace.Data(ctx, "checkout.item_count", len(in.Items))
	trace.Data(ctx, "checkout.origin", string(in.Origin))
	if userID != nil {
		trace.Data(ctx, "checkout.user_id", *userID)
	}

	if len(in.Items) == 0 {
		trace.Err(ctx, fmt.Errorf("no items"))
		return nil, fmt.Errorf("no items")
	}

	// Collect all product IDs
	productIDSet := make(map[string]struct{})
	for _, it := range in.Items {
		if !nanoid.Valid(it.ProductID) {
			return nil, fmt.Errorf("invalid productId: %s", it.ProductID)
		}
		productIDSet[it.ProductID] = struct{}{}
		// Also collect configured child products
		for _, childID := range it.Configuration {
			if childID == "" {
				continue
			}
			if !nanoid.Valid(childID) {
				return nil, fmt.Errorf("invalid configuration productId: %s", childID)
			}
			productIDSet[childID] = struct{}{}
		}
	}

	ids := make([]string, 0, len(productIDSet))
	for id := range productIDSet {
		ids = append(ids, id)
	}

	var (
		products       []*ent.Product
		slots          []*ent.MenuSlot
		preloadedStock map[string]int
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		products, err = s.products.GetByIDs(gctx, ids)
		if err != nil {
			return fmt.Errorf("load products: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		slots, err = s.menuSlotRepo.GetByMenuProductIDs(gctx, ids)
		if err != nil {
			return fmt.Errorf("load menu slots: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		preloadedStock, err = s.inventoryRepo.GetCurrentStockBatch(gctx, ids)
		if err != nil {
			return fmt.Errorf("check inventory: %w", err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	productMap := make(map[string]*ent.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	slotByID := make(map[string]*ent.MenuSlot, len(slots))
	allowedBySlot := make(map[string]map[string]struct{}, len(slots))
	for _, slot := range slots {
		slotByID[slot.ID] = slot
		allowed := make(map[string]struct{})
		for _, opt := range slot.Edges.Options {
			allowed[opt.OptionProductID] = struct{}{}
		}
		allowedBySlot[slot.ID] = allowed
	}

	// Calculate total and validate
	var totalCents int64
	for _, it := range in.Items {
		pid := it.ProductID
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
	requiredQuantities := make(map[string]int)
	for _, it := range in.Items {
		pid := it.ProductID
		p := productMap[pid]
		if p.Type == product.TypeSimple {
			requiredQuantities[pid] += it.Quantity
		}
		for _, childIDStr := range it.Configuration {
			if childIDStr == "" {
				continue
			}
			childID := childIDStr
			requiredQuantities[childID] += it.Quantity
		}
	}

	for pid, required := range requiredQuantities {
		available := preloadedStock[pid]
		if available < required {
			pName := "unknown"
			if p, ok := productMap[pid]; ok {
				pName = p.Name
			}
			return nil, fmt.Errorf("insufficient inventory for %s: requested %d, available %d", pName, required, available)
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
		pid := it.ProductID
		p := productMap[pid]

		// Determine parent line type
		lt := orderline.LineTypeSimple
		if p.Type == product.TypeMenu {
			lt = orderline.LineTypeBundle
		}

		parentLineID := nanoid.New()
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
				slotID := slotIDStr
				if !nanoid.Valid(slotID) {
					return nil, fmt.Errorf("invalid menu slot id: %s", slotIDStr)
				}
				slot, ok := slotByID[slotID]
				if !ok || slot.MenuProductID != p.ID {
					return nil, fmt.Errorf("slot does not belong to product: %s", slotIDStr)
				}
				childProdID := childProdIDStr
				if !nanoid.Valid(childProdID) {
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
		s.publishInventoryUpdates(ctx, inventoryEntries, preloadedStock)
	}

	// An order is redeemed offline from its token, so a signing/persist failure
	// must fail the order rather than create an unredeemable one.
	qrToken, err := s.signOrderQR(ctx, ord.ID, orderLines)
	if err != nil {
		return nil, fmt.Errorf("sign order qr: %w", err)
	}
	if err := s.orderRepo.SetQRPayload(ctx, ord.ID, qrToken); err != nil {
		return nil, fmt.Errorf("persist qr payload: %w", err)
	}
	ord.QrPayload = &qrToken

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

	trace.Data(ctx, "checkout.order_id", ord.ID)
	trace.Data(ctx, "checkout.total_cents", totalCents)
	trace.Data(ctx, "checkout.line_count", len(orderLines))

	return &CheckoutPreparation{
		OrderID:       ord.ID,
		TotalCents:    totalCents,
		LineItems:     lineItems,
		CustomerEmail: in.CustomerEmail,
		UserID:        userID,
		Order:         ord,
		QRPayload:     qrToken,
	}, nil
}

// signOrderQR signs over the redeemable physical items: simple products and
// chosen menu components. A menu's bundle parent is omitted (its components carry
// the items) unless it has none, then it falls back to the parent.
func (s *paymentService) signOrderQR(ctx context.Context, orderID string, lines []repository.OrderLineCreateParams) (string, error) {
	_, finish := trace.StartSpan(ctx, "service", "payment.sign_qr")
	defer finish()

	if s.qrKeys == nil {
		return "", fmt.Errorf("qr signing not configured")
	}

	hasComponents := make(map[string]bool, len(lines))
	for _, l := range lines {
		if l.ParentLineID != nil {
			hasComponents[*l.ParentLineID] = true
		}
	}

	redeemable := make([]qrsign.Line, 0, len(lines))
	for _, l := range lines {
		switch l.LineType {
		case orderline.LineTypeSimple, orderline.LineTypeComponent:
			redeemable = append(redeemable, qrsign.Line{ProductID: l.ProductID, Quantity: l.Quantity})
		case orderline.LineTypeBundle:
			if l.ID == nil || !hasComponents[*l.ID] {
				redeemable = append(redeemable, qrsign.Line{ProductID: l.ProductID, Quantity: l.Quantity})
			}
		}
	}

	priv, _ := s.qrKeys.SigningKey()

	token, err := qrsign.Sign(priv, qrsign.Payload{
		Version:  qrsign.Version,
		OrderID:  orderID,
		IssuedAt: time.Now().Unix(),
		Lines:    redeemable,
	})
	if err != nil {
		return "", fmt.Errorf("qr sign: sign payload: %w", err)
	}
	return token, nil
}

func (s *paymentService) CreatePayrexxGateway(ctx context.Context, prep *CheckoutPreparation, successURL, failedURL, cancelURL string) (*payrexx.Gateway, error) {
	ctx, finish := trace.StartSpan(ctx, "service", "payment.create_gateway")
	defer finish()
	trace.Data(ctx, "gateway.order_id", prep.OrderID)
	trace.Data(ctx, "gateway.amount_cents", prep.TotalCents)

	if s.payrexxClient == nil {
		trace.Err(ctx, fmt.Errorf("payrexx client not configured"))
		return nil, fmt.Errorf("payrexx client not configured")
	}

	gatewayCtx, gatewayCancel := context.WithTimeout(ctx, payrexx.DefaultRequestTimeout)
	defer gatewayCancel()
	gateway, err := s.payrexxClient.CreateGateway(gatewayCtx, payrexx.CreateGatewayParams{
		Amount:             int(prep.TotalCents),
		Currency:           "CHF",
		ReferenceID:        prep.OrderID,
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

func (s *paymentService) MarkOrderPaidByPayrexx(ctx context.Context, orderID string, gatewayID, transactionID int, contactEmail *string) error {
	ctx, finish := trace.StartSpan(ctx, "service", "payment.mark_paid_payrexx")
	defer finish()
	trace.Data(ctx, "payment.order_id", orderID)
	trace.Data(ctx, "payment.gateway_id", gatewayID)
	trace.Data(ctx, "payment.transaction_id", transactionID)

	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	ce := ord.ContactEmail
	if contactEmail != nil && *contactEmail != "" {
		ce = contactEmail
	}

	now := time.Now()

	if _, err := s.orderPaymentRepo.Create(ctx, ord.ID, orderpayment.MethodTWINT, ord.TotalCents, now, nil); err != nil {
		return fmt.Errorf("create order payment: %w", err)
	}

	if _, err = s.orderRepo.Update(ctx, ord.ID, ord.TotalCents, order.StatusPaid, ord.Origin, ord.CustomerID, ce, ord.PaymentAttemptID, &gatewayID, &transactionID); err != nil {
		return err
	}

	email := ""
	if ce != nil {
		email = *ce
	}
	if email != "" {
		s.logger.Info("scheduling receipt email",
			zap.String("orderId", ord.ID),
			zap.String("to", email),
		)
		go s.sendReceipt(ord.ID, email, ord.TotalCents, now, "TWINT")
	} else {
		s.logger.Info("no contact email, skipping receipt",
			zap.String("orderId", ord.ID),
		)
	}

	return nil
}

func (s *paymentService) MarkOrderPaidDev(ctx context.Context, orderID string) error {
	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if ord.Status != order.StatusPending {
		return fmt.Errorf("order is not pending")
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, order.StatusPaid); err != nil {
		return err
	}

	email := ""
	if ord.ContactEmail != nil {
		email = *ord.ContactEmail
	}
	if email != "" {
		s.logger.Info("scheduling receipt email (dev)",
			zap.String("orderId", ord.ID),
			zap.String("to", email),
		)
		go s.sendReceipt(ord.ID, email, ord.TotalCents, time.Now(), "TWINT (Dev)")
	}

	return nil
}

func (s *paymentService) FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error) {
	if attemptID == "" {
		return nil, fmt.Errorf("missing attempt id")
	}
	return s.orderRepo.FindPendingByAttemptID(ctx, attemptID)
}

func (s *paymentService) SetOrderAttemptID(ctx context.Context, orderID string, attemptID string) error {
	return s.orderRepo.SetPaymentAttemptID(ctx, orderID, attemptID)
}

func (s *paymentService) CleanupPendingOrderByID(ctx context.Context, orderID string) error {
	ctx, finish := trace.StartSpan(ctx, "service", "payment.cleanup_pending")
	defer finish()
	trace.Data(ctx, "cleanup.order_id", orderID)

	ord, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		trace.Err(ctx, err)
		return err
	}
	if ord.Status != order.StatusPending {
		trace.Data(ctx, "cleanup.skipped", "not_pending")
		return nil
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
		s.publishInventoryUpdates(ctx, releaseEntries, nil)
	}

	// Delete order (cascade deletes order lines)
	_, err = s.orderRepo.DeleteIfPending(ctx, orderID)
	return err
}

func (s *paymentService) CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID string) (int64, error) {
	return s.orderRepo.DeletePendingByAttemptIDExcept(ctx, attemptID, keepOrderID)
}

func (s *paymentService) GetPayrexxGateway(ctx context.Context, gatewayID int) (*payrexx.Gateway, error) {
	if s.payrexxClient == nil {
		return nil, fmt.Errorf("payrexx client not configured")
	}
	return s.payrexxClient.GetGateway(ctx, gatewayID)
}

func (s *paymentService) sendReceipt(orderID string, to string, totalCents int64, paidAt time.Time, method string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lines, err := s.orderLineRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		s.logger.Error("receipt: failed to load order lines", zap.Error(err), zap.String("orderId", orderID))
		return
	}

	childrenByParent := make(map[string][]ReceiptLineItem)
	var roots []*ent.OrderLine
	for _, l := range lines {
		if l.ParentLineID != nil {
			childrenByParent[*l.ParentLineID] = append(childrenByParent[*l.ParentLineID], ReceiptLineItem{
				Title:    l.Title,
				Quantity: l.Quantity,
				Cents:    l.UnitPriceCents,
			})
		} else {
			roots = append(roots, l)
		}
	}

	items := make([]ReceiptLineItem, 0, len(roots))
	for _, r := range roots {
		items = append(items, ReceiptLineItem{
			Title:    r.Title,
			Quantity: r.Quantity,
			Cents:    r.UnitPriceCents,
			Children: childrenByParent[r.ID],
		})
	}

	baseURL := strings.TrimRight(s.cfg.App.PublicBaseURL, "/")
	orderURL := baseURL + "/food/orders/" + orderID

	data := ReceiptEmailData{
		Brand:      "BlessThun Food",
		OrderID:    orderID,
		OrderURL:   orderURL,
		OrderDate:  formatOrderDate(paidAt),
		Items:      items,
		TotalCents: totalCents,
		Method:     method,
	}

	if err := s.emailService.SendReceiptEmail(ctx, to, data); err != nil {
		s.logger.Error("receipt: failed to send", zap.Error(err), zap.String("orderId", orderID), zap.String("to", to))
	}
}

func safeStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *paymentService) publishInventoryUpdates(ctx context.Context, entries []repository.InventoryLedgerCreateParams, preStock map[string]int) {
	if s.inventoryHub == nil {
		return
	}
	productIDs := make([]string, 0, len(entries))
	deltaByProduct := make(map[string]int)
	for _, entry := range entries {
		if _, seen := deltaByProduct[entry.ProductID]; !seen {
			productIDs = append(productIDs, entry.ProductID)
		}
		deltaByProduct[entry.ProductID] += entry.Delta
	}
	stocks := make(map[string]int, len(productIDs))
	missing := make([]string, 0)
	for _, pid := range productIDs {
		if pre, ok := preStock[pid]; ok {
			stocks[pid] = pre + deltaByProduct[pid]
		} else {
			missing = append(missing, pid)
		}
	}
	if len(missing) > 0 {
		fetched, err := s.inventoryRepo.GetCurrentStockBatch(ctx, missing)
		if err != nil {
			return
		}
		for pid, v := range fetched {
			stocks[pid] = v
		}
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
