package service

import (
    "backend/internal/config"
    "backend/internal/domain"
    "backend/internal/repository"
    "context"
    "fmt"

    stripe "github.com/stripe/stripe-go/v82"
    "github.com/stripe/stripe-go/v82/checkout/session"
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
}

type PaymentService interface {
    CreateTWINTCheckoutSession(ctx context.Context, p CreateCheckoutSessionParams) (*stripe.CheckoutSession, error)
    PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string) (*CheckoutPreparation, error)
    CreateStripeCheckoutForOrder(ctx context.Context, prep *CheckoutPreparation, successURL, cancelURL string) (*stripe.CheckoutSession, error)
    MarkOrderPaid(ctx context.Context, clientReferenceID string, contactEmail *string) error
}

type paymentService struct {
    cfg           config.Config
    orderRepo     repository.OrderRepository
    orderItemRepo repository.OrderItemRepository
    productRepo   repository.ProductRepository
    menuSlotRepo  repository.MenuSlotRepository
    menuSlotItemRepo repository.MenuSlotItemRepository
    userRepo      repository.UserRepository
}

func NewPaymentService(cfg config.Config, orderRepo repository.OrderRepository, orderItemRepo repository.OrderItemRepository, productRepo repository.ProductRepository, menuSlotRepo repository.MenuSlotRepository, menuSlotItemRepo repository.MenuSlotItemRepository, userRepo repository.UserRepository) PaymentService {
    // Set global Stripe key for SDK
    stripe.Key = cfg.Stripe.SecretKey
    return &paymentService{cfg: cfg, orderRepo: orderRepo, orderItemRepo: orderItemRepo, productRepo: productRepo, menuSlotRepo: menuSlotRepo, menuSlotItemRepo: menuSlotItemRepo, userRepo: userRepo}
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
    if p.ClientRefID != nil && *p.ClientRefID != "" {
        params.ClientReferenceID = stripe.String(*p.ClientRefID)
    }
    if p.CustomerEmail != nil {
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
}

func (s *paymentService) PrepareAndCreateOrder(ctx context.Context, in CreateCheckoutInput, userID *string) (*CheckoutPreparation, error) {
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
    // If logged in and no email provided, fetch and set customer email to skip email collection
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
    id, err := s.orderRepo.Create(ctx, ord)
    if err != nil { return nil, fmt.Errorf("create order: %w", err) }

    // Build and insert order items
    oitems := make([]*domain.OrderItem, 0, len(dto.OrderItems))
    for idx, it := range in.Items {
        p := pm[dto.OrderItems[idx].ProductID]
        oitems = append(oitems, &domain.OrderItem{
            ID:                primitive.NewObjectID(),
            OrderID:           id,
            ProductID:         p.ID,
            Title:             p.Name,
            Quantity:          int(it.Quantity),
            PricePerUnitCents: p.PriceCents,
            IsRedeemed:        false,
        })
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
                // Parent ID is the last inserted parent (we can reference via created OrderItem's ID)
                parentID := oitems[len(oitems)-1].ID
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

    return &CheckoutPreparation{ OrderID: id, LineItems: lis, CustomerEmail: in.CustomerEmail }, nil
}

func (s *paymentService) CreateStripeCheckoutForOrder(ctx context.Context, prep *CheckoutPreparation, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
    idHex := prep.OrderID.Hex()
    sess, err := s.CreateTWINTCheckoutSession(ctx, CreateCheckoutSessionParams{
        Items:         prep.LineItems,
        SuccessURL:    successURL,
        CancelURL:     cancelURL,
        ClientRefID:   &idHex,
        CustomerEmail: prep.CustomerEmail,
    })
    if err != nil { return nil, err }
    _ = s.orderRepo.SetStripeSession(ctx, prep.OrderID, sess.ID)
    return sess, nil
}

func (s *paymentService) MarkOrderPaid(ctx context.Context, clientReferenceID string, contactEmail *string) error {
    if clientReferenceID == "" { return fmt.Errorf("missing client_reference_id") }
    oid, err := primitive.ObjectIDFromHex(clientReferenceID)
    if err != nil { return fmt.Errorf("invalid order id in client_reference_id") }
    return s.orderRepo.UpdateStatusAndContact(ctx, oid, domain.OrderStatusPaid, contactEmail)
}
