package service

import (
    "backend/internal/config"
    "backend/internal/domain"
    "backend/internal/repository"
    "context"
    "fmt"

    stripe "github.com/stripe/stripe-go/v82"
    "github.com/stripe/stripe-go/v82/checkout/session"
    "github.com/stripe/stripe-go/v82/customer"
    "github.com/stripe/stripe-go/v82/paymentintent"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckoutItem struct {
    Name        string
    AmountCents int64
    Currency    string
    Quantity    int64
}

type CreateCheckoutSessionParams struct {
    Items       []CheckoutItem
    SuccessURL  string
    CancelURL   string
    ClientRefID *string
    CustomerEmail *string
    // If provided, Stripe Checkout will use this customer and not show email input
    CustomerID   *string
    // Optional metadata to attach to the underlying PaymentIntent
    PaymentIntentMetadata map[string]string
}

type PaymentService interface {
    CreateTWINTCheckoutSession(ctx context.Context, p CreateCheckoutSessionParams) (*stripe.CheckoutSession, error)
    PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error)
    CreateStripeCheckoutForOrder(ctx context.Context, prep *CheckoutPreparation, successURL, cancelURL string) (*stripe.CheckoutSession, error)
    // PaymentIntents flow
    CreatePaymentIntentForOrder(ctx context.Context, prep *CheckoutPreparation, receiptEmail *string) (*stripe.PaymentIntent, error)
    UpdatePaymentIntentReceiptEmail(ctx context.Context, paymentIntentID string, email *string) (*stripe.PaymentIntent, error)
    GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error)
    MarkOrderPaid(ctx context.Context, clientReferenceID string, contactEmail *string) error
    PersistPaymentSuccessByOrderID(ctx context.Context, orderIDHex string, paymentIntentID string, chargeID *string, customerID *string, receiptEmail *string) error
    FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*domain.Order, error)
    SetOrderAttemptID(ctx context.Context, orderID primitive.ObjectID, attemptID string) error
    CreatePaymentIntentForExistingPendingOrder(ctx context.Context, ord *domain.Order, userID *string, receiptEmail *string) (*stripe.PaymentIntent, error)
    CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID primitive.ObjectID) (int64, error)
    CleanupPendingOrderByID(ctx context.Context, clientReferenceID string) error
    CleanupPendingOrderBySessionID(ctx context.Context, sessionID string) error
}

type paymentService struct {
    cfg           config.Config
    orderRepo     repository.OrderRepository
    orderItemRepo repository.OrderItemRepository
    productRepo   repository.ProductRepository
    menuSlotRepo  repository.MenuSlotRepository
    menuSlotItemRepo repository.MenuSlotItemRepository
    userRepo      repository.UserRepository
    inventoryRepo repository.InventoryLedgerRepository
}

func NewPaymentService(cfg config.Config, orderRepo repository.OrderRepository, orderItemRepo repository.OrderItemRepository, productRepo repository.ProductRepository, menuSlotRepo repository.MenuSlotRepository, menuSlotItemRepo repository.MenuSlotItemRepository, userRepo repository.UserRepository, inventoryRepo repository.InventoryLedgerRepository) PaymentService {
    // Set global Stripe key for SDK
    stripe.Key = cfg.Stripe.SecretKey
    return &paymentService{cfg: cfg, orderRepo: orderRepo, orderItemRepo: orderItemRepo, productRepo: productRepo, menuSlotRepo: menuSlotRepo, menuSlotItemRepo: menuSlotItemRepo, userRepo: userRepo, inventoryRepo: inventoryRepo}
}

func (s *paymentService) CreateTWINTCheckoutSession(ctx context.Context, p CreateCheckoutSessionParams) (*stripe.CheckoutSession, error) {
    if len(p.Items) == 0 {
        return nil, fmt.Errorf("no items provided")
    }

    lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(p.Items))
    for _, it := range p.Items {
        if it.Currency == "" {
            it.Currency = "chf"
        }
        if it.AmountCents <= 0 || it.Quantity <= 0 {
            return nil, fmt.Errorf("invalid amount or quantity for item %q", it.Name)
        }
        // TWINT limit: 5000 CHF per transaction
        if it.Currency == "chf" && it.AmountCents*it.Quantity > 500000 { // 5000 CHF in cents
            return nil, fmt.Errorf("item %q exceeds TWINT maximum amount", it.Name)
        }
        lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
            PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                Currency: stripe.String(it.Currency),
                ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                    Name: stripe.String(it.Name),
                },
                UnitAmount: stripe.Int64(it.AmountCents),
            },
            Quantity: stripe.Int64(it.Quantity),
        })
    }

    params := &stripe.CheckoutSessionParams{
        Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
        SuccessURL: stripe.String(p.SuccessURL),
        CancelURL:  stripe.String(p.CancelURL),
        PaymentMethodTypes: []*string{
            stripe.String("twint"),
        },
        LineItems: lineItems,
    }
    // Attach PaymentIntent metadata so we can correlate payment_intent events
    if len(p.PaymentIntentMetadata) > 0 {
        params.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{}
        // populate metadata map
        for k, v := range p.PaymentIntentMetadata {
            params.PaymentIntentData.AddMetadata(k, v)
        }
    }
    if p.ClientRefID != nil && *p.ClientRefID != "" {
        params.ClientReferenceID = stripe.String(*p.ClientRefID)
    }
    // Prefer explicit Customer over email to suppress email field in Checkout
    if p.CustomerID != nil && *p.CustomerID != "" {
        params.Customer = stripe.String(*p.CustomerID)
    } else if p.CustomerEmail != nil {
        params.CustomerEmail = stripe.String(*p.CustomerEmail)
    }

    // Create the Checkout Session
    return session.New(params)
}

// ----- Checkout orchestration -----

// CreateCheckoutInput is the structured payload from frontend
type CreateCheckoutInput struct {
    Items []struct {
        ProductID string             `json:"productId"`
        Quantity  int64              `json:"quantity"`
        // Optional configuration map: slotId -> selected productId
        Configuration map[string]string `json:"configuration,omitempty"`
    } `json:"items"`
    CustomerEmail *string `json:"customerEmail,omitempty"`
}

// CheckoutPreparation bundles computed order and pricing for Stripe
type CheckoutPreparation struct {
    OrderID      primitive.ObjectID
    LineItems    []CheckoutItem
    CustomerEmail *string
    // If the order is associated to a logged-in user, store their ID for potential Stripe Customer usage
    UserID       *primitive.ObjectID
}

func (s *paymentService) PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string, attemptID *string) (*CheckoutPreparation, error) {
    if len(in.Items) == 0 {
        return nil, fmt.Errorf("no items")
    }

    // Resolve user
    var customerOID *primitive.ObjectID
    if userID != nil && *userID != "" {
        if oid, err := primitive.ObjectIDFromHex(*userID); err == nil {
            customerOID = &oid
        }
    }
    // If logged in and no email provided, fetch and set customer email to skip email collection (prefill)
    if customerOID != nil && in.CustomerEmail == nil {
        if u, err := s.userRepo.FindByID(ctx, *customerOID); err == nil {
            in.CustomerEmail = &u.Email
        }
    }

    // Build CreateOrderDTO for validation and totals
    dto := &domain.CreateOrderDTO{ ContactEmail: in.CustomerEmail }
    // Track all referenced product IDs (parents + configured children)
    childIDs := make([]primitive.ObjectID, 0)
    for _, it := range in.Items {
        pid, err := primitive.ObjectIDFromHex(it.ProductID)
        if err != nil { return nil, fmt.Errorf("invalid productId: %s", it.ProductID) }
        dto.OrderItems = append(dto.OrderItems, domain.CreateOrderItemDTO{ ProductID: pid, Quantity: int(it.Quantity) })
        for _, child := range it.Configuration {
            if child == "" { continue }
            cid, err := primitive.ObjectIDFromHex(child)
            if err != nil { return nil, fmt.Errorf("invalid configuration productId: %s", child) }
            childIDs = append(childIDs, cid)
        }
    }

    // Validate and compute totals
    ids := make([]primitive.ObjectID, 0, len(dto.OrderItems))
    seen := map[primitive.ObjectID]struct{}{}
    for _, it := range dto.OrderItems {
        if _, ok := seen[it.ProductID]; !ok { ids = append(ids, it.ProductID); seen[it.ProductID] = struct{}{} }
    }
    // Include child configured products in the load set
    ids = append(ids, childIDs...)
    products, err := s.productRepo.GetByIDs(ctx, ids)
    if err != nil { return nil, fmt.Errorf("load products: %w", err) }
    pm := map[primitive.ObjectID]*domain.Product{}
    for _, p := range products { pm[p.ID] = p }

    var total domain.Cents
    // Preload menu slots/items for menu-type products
    menuProdIDs := make([]primitive.ObjectID, 0)
    for _, it := range dto.OrderItems {
        if p := pm[it.ProductID]; p != nil && p.Type == domain.ProductTypeMenu {
            menuProdIDs = append(menuProdIDs, p.ID)
        }
    }
    slotsByProduct := map[primitive.ObjectID][]*domain.MenuSlot{}
    slotByID := map[primitive.ObjectID]*domain.MenuSlot{}
    allowedBySlot := map[primitive.ObjectID]map[primitive.ObjectID]struct{}{}
    if len(menuProdIDs) > 0 {
        slots, err := s.menuSlotRepo.FindByProductIDs(ctx, menuProdIDs)
        if err != nil { return nil, fmt.Errorf("load menu slots: %w", err) }
        slotIDs := make([]primitive.ObjectID, 0, len(slots))
        for _, sl := range slots {
            slotsByProduct[sl.ProductID] = append(slotsByProduct[sl.ProductID], sl)
            slotByID[sl.ID] = sl
            slotIDs = append(slotIDs, sl.ID)
        }
        if len(slotIDs) > 0 {
            mitems, err := s.menuSlotItemRepo.FindByMenuSlotIDs(ctx, slotIDs)
            if err != nil { return nil, fmt.Errorf("load menu slot items: %w", err) }
            for _, mi := range mitems {
                set := allowedBySlot[mi.MenuSlotID]
                if set == nil { set = map[primitive.ObjectID]struct{}{} }
                set[mi.ProductID] = struct{}{}
                allowedBySlot[mi.MenuSlotID] = set
            }
        }
    }

    for idx, it := range in.Items {
        // Lookup parent product
        parentDTO := dto.OrderItems[idx]
        p := pm[parentDTO.ProductID]
        if p == nil { return nil, fmt.Errorf("unknown product: %s", parentDTO.ProductID.Hex()) }
        // pricing
        total += p.PriceCents * domain.Cents(it.Quantity)
        if p.PriceCents*domain.Cents(it.Quantity) > 500000 { // 5000 CHF
            return nil, fmt.Errorf("item exceeds TWINT max: %s", p.Name)
        }
    }
    if total > 500000 { return nil, fmt.Errorf("order exceeds TWINT max (5000 CHF)") }

    // Create order
    ord := &domain.Order{
        CustomerID:   customerOID,
        ContactEmail: in.CustomerEmail,
        TotalCents:   total,
        Status:       domain.OrderStatusPending,
    }
    if attemptID != nil && *attemptID != "" {
        aid := *attemptID
        ord.PaymentAttemptID = &aid
    }
    id, err := s.orderRepo.Create(ctx, ord)
    if err != nil { return nil, fmt.Errorf("create order: %w", err) }

    // Build and insert order items
    oitems := make([]*domain.OrderItem, 0, len(dto.OrderItems))
    for idx, it := range in.Items {
        p := pm[dto.OrderItems[idx].ProductID]
        // Create the parent order item first
        parentOrderItem := &domain.OrderItem{
            ID:                primitive.NewObjectID(),
            OrderID:           id,
            ProductID:         p.ID,
            Title:             p.Name,
            Quantity:          int(it.Quantity),
            PricePerUnitCents: p.PriceCents,
            IsRedeemed:        false,
        }
        oitems = append(oitems, parentOrderItem)
        // Add configuration children if menu product
        if p.Type == domain.ProductTypeMenu && len(it.Configuration) > 0 {
            // Ensure slot belongs to this product and child allowed
            for slotIDStr, childProdIDStr := range it.Configuration {
                if childProdIDStr == "" { continue }
                slotOID, err := primitive.ObjectIDFromHex(slotIDStr)
                if err != nil { return nil, fmt.Errorf("invalid menu slot id: %s", slotIDStr) }
                sl := slotByID[slotOID]
                if sl == nil || sl.ProductID != p.ID { return nil, fmt.Errorf("slot does not belong to product: %s", slotIDStr) }
                childOID, err := primitive.ObjectIDFromHex(childProdIDStr)
                if err != nil { return nil, fmt.Errorf("invalid configured product id: %s", childProdIDStr) }
                // Validate allowed
                if set := allowedBySlot[slotOID]; set != nil {
                    if _, ok := set[childOID]; !ok { return nil, fmt.Errorf("product not allowed in slot") }
                }
                childP := pm[childOID]
                title := "Configured Item"
                if childP != nil { title = childP.Name }
                // Parent ID must be the parent order item (not the previously appended child)
                parentID := parentOrderItem.ID
                oitems = append(oitems, &domain.OrderItem{
                    ID:                primitive.NewObjectID(),
                    OrderID:           id,
                    ProductID:         childOID,
                    Title:             title,
                    Quantity:          int(it.Quantity),
                    PricePerUnitCents: 0,
                    ParentItemID:      &parentID,
                    MenuSlotID:        &slotOID,
                    MenuSlotName:      &sl.Name,
                    IsRedeemed:        false,
                })
            }
        }
    }
    if err := s.orderItemRepo.InsertMany(ctx, oitems); err != nil {
        return nil, fmt.Errorf("insert order items: %w", err)
    }

    // Consume inventory immediately on pending order creation
    // For parent simple products: consume; for menu parents: skip; for children: always consume
    // Build consumption entries from created items
    entries := make([]*domain.InventoryLedger, 0)
    for _, oi := range oitems {
        if oi.ProductID.IsZero() { continue }
        consume := false
        if oi.ParentItemID == nil {
            // parent item: check product type
            if p := pm[oi.ProductID]; p != nil && p.Type == domain.ProductTypeSimple {
                consume = true
            }
        } else {
            // child item (menu component) always consumes
            consume = true
        }
        if consume && oi.Quantity > 0 {
            entries = append(entries, &domain.InventoryLedger{
                ProductID: oi.ProductID,
                Delta:     -oi.Quantity,
                Reason:    domain.InventoryReasonSale,
            })
        }
    }
    // Best-effort append; if this fails, return error to avoid placing order without reservation
    if err := s.inventoryRepo.AppendMany(ctx, entries); err != nil {
        return nil, fmt.Errorf("reserve inventory: %w", err)
    }

    // Prepare Stripe line items (only priced parent items)
    lis := make([]CheckoutItem, 0)
    for _, oi := range oitems {
        if oi.ParentItemID == nil && oi.PricePerUnitCents > 0 && oi.Quantity > 0 {
            lis = append(lis, CheckoutItem{
                Name:        oi.Title,
                AmountCents: int64(oi.PricePerUnitCents),
                Currency:    "chf",
                Quantity:    int64(oi.Quantity),
            })
        }
    }

    return &CheckoutPreparation{ OrderID: id, LineItems: lis, CustomerEmail: in.CustomerEmail, UserID: customerOID }, nil
}

func (s *paymentService) CreateStripeCheckoutForOrder(ctx context.Context, prep *CheckoutPreparation, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
    idHex := prep.OrderID.Hex()
    var customerID *string
    var customerEmail = prep.CustomerEmail
    // If order has an associated user, prefer using a Stripe Customer to hide email input on Checkout
    if prep.UserID != nil {
        if u, err := s.userRepo.FindByID(ctx, *prep.UserID); err == nil && u != nil {
            if u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
                customerID = u.StripeCustomerID
            } else {
                // Create a Stripe Customer for this user (best effort). If it fails, fall back to email.
                c, cerr := customer.New(&stripe.CustomerParams{Email: stripe.String(u.Email)})
                if cerr == nil && c != nil {
                    // persist the mapping (best effort)
                    _ = s.userRepo.UpdateStripeCustomerID(ctx, u.ID, c.ID)
                    customerID = &c.ID
                } else {
                    // Fallback to using email
                    ce := u.Email
                    customerEmail = &ce
                }
            }
        }
    }

    sess, err := s.CreateTWINTCheckoutSession(ctx, CreateCheckoutSessionParams{
        Items:         prep.LineItems,
        SuccessURL:    successURL,
        CancelURL:     cancelURL,
        ClientRefID:   &idHex,
        CustomerEmail: customerEmail,
        CustomerID:    customerID,
        PaymentIntentMetadata: map[string]string{"order_id": idHex},
    })
    if err != nil { return nil, err }
    _ = s.orderRepo.SetStripeSession(ctx, prep.OrderID, sess.ID)
    return sess, nil
}

// ---- Payment Intents (Payment Element) ----

// CreatePaymentIntentForOrder creates a CHF PaymentIntent constrained to TWINT for the given prepared order.
// Optionally attaches a receipt_email and links to a Stripe Customer when the user is logged in.
func (s *paymentService) CreatePaymentIntentForOrder(ctx context.Context, prep *CheckoutPreparation, receiptEmail *string) (*stripe.PaymentIntent, error) {
    idHex := prep.OrderID.Hex()

    // Determine Stripe Customer to attach if user exists
    var customerID *string
    fallbackEmail := receiptEmail
    if prep.UserID != nil {
        if u, err := s.userRepo.FindByID(ctx, *prep.UserID); err == nil && u != nil {
            if u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
                customerID = u.StripeCustomerID
            } else {
                // Create Stripe Customer best-effort with account email
                c, cerr := customer.New(&stripe.CustomerParams{Email: stripe.String(u.Email)})
                if cerr == nil && c != nil {
                    _ = s.userRepo.UpdateStripeCustomerID(ctx, u.ID, c.ID)
                    customerID = &c.ID
                } else {
                    // default receipt to account email if none explicitly provided
                    if fallbackEmail == nil || *fallbackEmail == "" {
                        e := u.Email
                        fallbackEmail = &e
                    }
                }
            }
        }
    }

    // Compute total amount from prepared line items
    var total int64
    for _, li := range prep.LineItems {
        total += li.AmountCents * li.Quantity
    }
    if total <= 0 {
        return nil, fmt.Errorf("invalid order total")
    }

    params := &stripe.PaymentIntentParams{
        Amount:             stripe.Int64(total), // rappen
        Currency:           stripe.String(string(stripe.CurrencyCHF)),
        PaymentMethodTypes: stripe.StringSlice([]string{"twint"}),
    }
    // Attach metadata for correlation
    params.AddMetadata("order_id", idHex)
    if prep.UserID != nil {
        params.AddMetadata("user_id", prep.UserID.Hex())
    }
    if customerID != nil {
        params.Customer = stripe.String(*customerID)
    }
    // If we have an explicit or fallback email, attach as receipt_email
    if fallbackEmail != nil && *fallbackEmail != "" {
        params.ReceiptEmail = stripe.String(*fallbackEmail)
    }
    // Idempotency on order id to avoid duplicate PIs on retry
    params.SetIdempotencyKey("create_pi:" + idHex)
    pi, err := paymentintent.New(params)
    if err != nil {
        return nil, err
    }
    // Best effort: persist references on order
    _ = s.orderRepo.SetStripePaymentIntent(ctx, prep.OrderID, pi.ID, customerID, params.ReceiptEmail)
    return pi, nil
}

// UpdatePaymentIntentReceiptEmail sets or clears receipt_email on an existing PI.
func (s *paymentService) UpdatePaymentIntentReceiptEmail(ctx context.Context, paymentIntentID string, email *string) (*stripe.PaymentIntent, error) {
    if paymentIntentID == "" {
        return nil, fmt.Errorf("missing paymentIntentId")
    }
    params := &stripe.PaymentIntentParams{}
    if email != nil {
        // Setting empty string is treated as clearing by Stripe API
        params.ReceiptEmail = stripe.String(*email)
    } else {
        // Explicitly clear by setting to empty string
        empty := ""
        params.ReceiptEmail = &empty
    }
    params.SetIdempotencyKey("attach_email:" + paymentIntentID + ":" + safeStr(email))
    return paymentintent.Update(paymentIntentID, params)
}

// GetPaymentIntent fetches a PI by id
func (s *paymentService) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
    if paymentIntentID == "" {
        return nil, fmt.Errorf("missing id")
    }
    return paymentintent.Get(paymentIntentID, nil)
}

func safeStr(p *string) string {
    if p == nil { return "" }
    return *p
}

func (s *paymentService) MarkOrderPaid(ctx context.Context, clientReferenceID string, contactEmail *string) error {
    if clientReferenceID == "" { return fmt.Errorf("missing client_reference_id") }
    oid, err := primitive.ObjectIDFromHex(clientReferenceID)
    if err != nil { return fmt.Errorf("invalid order id in client_reference_id") }
    return s.orderRepo.UpdateStatusAndContact(ctx, oid, domain.OrderStatusPaid, contactEmail)
}

func (s *paymentService) PersistPaymentSuccessByOrderID(ctx context.Context, orderIDHex string, paymentIntentID string, chargeID *string, customerID *string, receiptEmail *string) error {
    if orderIDHex == "" { return fmt.Errorf("missing order id") }
    oid, err := primitive.ObjectIDFromHex(orderIDHex)
    if err != nil { return fmt.Errorf("invalid order id: %w", err) }
    return s.orderRepo.SetStripePaymentSuccess(ctx, oid, paymentIntentID, chargeID, customerID, receiptEmail)
}

func (s *paymentService) FindPendingOrderByAttemptID(ctx context.Context, attemptID string) (*domain.Order, error) {
    if attemptID == "" { return nil, fmt.Errorf("missing attempt id") }
    return s.orderRepo.FindPendingByAttemptID(ctx, attemptID)
}

func (s *paymentService) SetOrderAttemptID(ctx context.Context, orderID primitive.ObjectID, attemptID string) error {
    if orderID.IsZero() || attemptID == "" { return nil }
    return s.orderRepo.SetPaymentAttemptID(ctx, orderID, attemptID)
}

// CreatePaymentIntentForExistingPendingOrder creates a PI for an already created pending order.
func (s *paymentService) CreatePaymentIntentForExistingPendingOrder(ctx context.Context, ord *domain.Order, userID *string, receiptEmail *string) (*stripe.PaymentIntent, error) {
    if ord == nil { return nil, fmt.Errorf("nil order") }
    idHex := ord.ID.Hex()
    // Determine Stripe Customer
    var customerID *string
    fallbackEmail := receiptEmail
    if userID != nil && *userID != "" {
        if oid, err := primitive.ObjectIDFromHex(*userID); err == nil {
            if u, err := s.userRepo.FindByID(ctx, oid); err == nil && u != nil {
                if u.StripeCustomerID != nil && *u.StripeCustomerID != "" {
                    customerID = u.StripeCustomerID
                } else {
                    c, cerr := customer.New(&stripe.CustomerParams{Email: stripe.String(u.Email)})
                    if cerr == nil && c != nil {
                        _ = s.userRepo.UpdateStripeCustomerID(ctx, u.ID, c.ID)
                        customerID = &c.ID
                    } else {
                        if fallbackEmail == nil || *fallbackEmail == "" {
                            e := u.Email
                            fallbackEmail = &e
                        }
                    }
                }
            }
        }
    }
    // Create PI from order total
    params := &stripe.PaymentIntentParams{
        Amount:             stripe.Int64(int64(ord.TotalCents)),
        Currency:           stripe.String(string(stripe.CurrencyCHF)),
        PaymentMethodTypes: stripe.StringSlice([]string{"twint"}),
    }
    params.AddMetadata("order_id", idHex)
    if userID != nil && *userID != "" { params.AddMetadata("user_id", *userID) }
    if customerID != nil { params.Customer = stripe.String(*customerID) }
    if fallbackEmail != nil && *fallbackEmail != "" { params.ReceiptEmail = stripe.String(*fallbackEmail) }
    params.SetIdempotencyKey("create_pi_existing:" + idHex)
    pi, err := paymentintent.New(params)
    if err != nil { return nil, err }
    _ = s.orderRepo.SetStripePaymentIntent(ctx, ord.ID, pi.ID, customerID, params.ReceiptEmail)
    return pi, nil
}

func (s *paymentService) CleanupOtherPendingOrdersByAttemptID(ctx context.Context, attemptID string, keepOrderID primitive.ObjectID) (int64, error) {
    if attemptID == "" || keepOrderID.IsZero() { return 0, nil }
    return s.orderRepo.DeletePendingByAttemptIDExcept(ctx, attemptID, keepOrderID)
}

// CleanupPendingOrderByID deletes a pending order and its items by order ID (hex string).
func (s *paymentService) CleanupPendingOrderByID(ctx context.Context, clientReferenceID string) error {
    if clientReferenceID == "" { return fmt.Errorf("missing order id") }
    oid, err := primitive.ObjectIDFromHex(clientReferenceID)
    if err != nil { return fmt.Errorf("invalid order id: %w", err) }
    // Check order status; only act on pending
    ord, err := s.orderRepo.FindByID(ctx, oid)
    if err != nil { return err }
    if ord.Status != domain.OrderStatusPending { return nil }
    // Load items to release inventory
    items, _ := s.orderItemRepo.FindByOrderID(ctx, oid)
    if len(items) > 0 {
        // Load product types to discriminate menu vs simple
        pidSet := map[primitive.ObjectID]struct{}{}
        for _, it := range items { if !it.ProductID.IsZero() { pidSet[it.ProductID] = struct{}{} } }
        pids := make([]primitive.ObjectID, 0, len(pidSet))
        for id := range pidSet { pids = append(pids, id) }
        pmap := map[primitive.ObjectID]*domain.Product{}
        if len(pids) > 0 {
            if prods, err := s.productRepo.GetByIDs(ctx, pids); err == nil {
                for _, p := range prods { pmap[p.ID] = p }
            }
        }
        rel := make([]*domain.InventoryLedger, 0)
        for _, it := range items {
            if it.ProductID.IsZero() || it.Quantity <= 0 { continue }
            if it.ParentItemID == nil {
                // parent: release only if simple (consumed earlier)
                if p := pmap[it.ProductID]; p != nil && p.Type == domain.ProductTypeSimple {
                    rel = append(rel, &domain.InventoryLedger{ ProductID: it.ProductID, Delta: it.Quantity, Reason: domain.InventoryReasonCorrection })
                }
            } else {
                // child always release
                rel = append(rel, &domain.InventoryLedger{ ProductID: it.ProductID, Delta: it.Quantity, Reason: domain.InventoryReasonCorrection })
            }
        }
        _ = s.inventoryRepo.AppendMany(ctx, rel)
    }
    // delete items first, then order
    _ = s.orderItemRepo.DeleteByOrderID(ctx, oid)
    _, _ = s.orderRepo.DeleteIfPending(ctx, oid)
    return nil
}

// CleanupPendingOrderBySessionID deletes a pending order matched by stripe_session_id.
func (s *paymentService) CleanupPendingOrderBySessionID(ctx context.Context, sessionID string) error {
    if sessionID == "" { return fmt.Errorf("missing session id") }
    o, err := s.orderRepo.FindPendingByStripeSessionID(ctx, sessionID)
    if err != nil { return nil } // nothing to do if not found
    // release inventory best-effort
    items, _ := s.orderItemRepo.FindByOrderID(ctx, o.ID)
    if len(items) > 0 {
        // load product types
        pidSet := map[primitive.ObjectID]struct{}{}
        for _, it := range items { if !it.ProductID.IsZero() { pidSet[it.ProductID] = struct{}{} } }
        pids := make([]primitive.ObjectID, 0, len(pidSet))
        for id := range pidSet { pids = append(pids, id) }
        pmap := map[primitive.ObjectID]*domain.Product{}
        if len(pids) > 0 {
            if prods, err := s.productRepo.GetByIDs(ctx, pids); err == nil {
                for _, p := range prods { pmap[p.ID] = p }
            }
        }
        rel := make([]*domain.InventoryLedger, 0)
        for _, it := range items {
            if it.ProductID.IsZero() || it.Quantity <= 0 { continue }
            if it.ParentItemID == nil {
                if p := pmap[it.ProductID]; p != nil && p.Type == domain.ProductTypeSimple {
                    rel = append(rel, &domain.InventoryLedger{ ProductID: it.ProductID, Delta: it.Quantity, Reason: domain.InventoryReasonCorrection })
                }
            } else {
                rel = append(rel, &domain.InventoryLedger{ ProductID: it.ProductID, Delta: it.Quantity, Reason: domain.InventoryReasonCorrection })
            }
        }
        _ = s.inventoryRepo.AppendMany(ctx, rel)
    }
    // delete items first, then order
    _ = s.orderItemRepo.DeleteByOrderID(ctx, o.ID)
    _, _ = s.orderRepo.DeleteIfPending(ctx, o.ID)
    return nil
}
