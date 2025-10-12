package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/repository"
    "backend/internal/response"
    "encoding/csv"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminOrderHandler struct {
    orders     repository.OrderRepository
    orderItems repository.OrderItemRepository
    products   repository.ProductRepository
    audit      repository.AuditRepository
}

func NewAdminOrderHandler(orders repository.OrderRepository, orderItems repository.OrderItemRepository, products repository.ProductRepository, audit repository.AuditRepository) *AdminOrderHandler {
    return &AdminOrderHandler{ orders: orders, orderItems: orderItems, products: products, audit: audit }
}

func (h *AdminOrderHandler) List(w http.ResponseWriter, r *http.Request) {
    // filters
    var status *domain.OrderStatus
    if s := r.URL.Query().Get("status"); s != "" {
        ss := domain.OrderStatus(s)
        status = &ss
    }
    var from, to *time.Time
    if v := r.URL.Query().Get("date_from"); v != "" {
        if t, err := time.Parse(time.RFC3339, v); err == nil { from = &t }
    }
    if v := r.URL.Query().Get("date_to"); v != "" {
        if t, err := time.Parse(time.RFC3339, v); err == nil { to = &t }
    }
    var q *string
    if v := r.URL.Query().Get("q"); v != "" { q = &v }
    limit := 50; offset := 0
    if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 { limit = n } }
    if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n } }

    items, total, err := h.orders.ListAdmin(r.Context(), status, from, to, q, limit, offset)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list orders"); return }
    // Expanded DTO for admin list view
    type OrderDTO struct {
        ID               string             `json:"id"`
        Status           domain.OrderStatus `json:"status"`
        TotalCents       int64              `json:"totalCents"`
        CreatedAt        time.Time          `json:"createdAt"`
        UpdatedAt        time.Time          `json:"updatedAt"`
        ContactEmail     *string            `json:"contactEmail,omitempty"`
        CustomerID       *string            `json:"customerId,omitempty"`
        PaymentIntentID  *string            `json:"paymentIntentId,omitempty"`
        StripeChargeID   *string            `json:"stripeChargeId,omitempty"`
        PaymentAttemptID *string            `json:"paymentAttemptId,omitempty"`
        Origin           *domain.OrderOrigin `json:"origin,omitempty"`
        PosPayment       *struct {
            Method              string  `json:"method"`
            AmountReceivedCents *int64  `json:"amountReceivedCents,omitempty"`
            ChangeCents         *int64  `json:"changeCents,omitempty"`
        } `json:"posPayment,omitempty"`
    }
    out := make([]OrderDTO, 0, len(items))
    var revenue int64
    for _, o := range items {
        var custID *string
        if o.CustomerID != nil {
            s := o.CustomerID.Hex()
            custID = &s
        }
        dto := OrderDTO{
            ID:               o.ID.Hex(),
            Status:           o.Status,
            TotalCents:       int64(o.TotalCents),
            CreatedAt:        o.CreatedAt,
            UpdatedAt:        o.UpdatedAt,
            ContactEmail:     o.ContactEmail,
            CustomerID:       custID,
            PaymentIntentID:  o.StripePaymentIntentID,
            StripeChargeID:   o.StripeChargeID,
            PaymentAttemptID: o.PaymentAttemptID,
            Origin:           o.Origin,
        }
        if o.PosPayment != nil {
            var receivedPtr, changePtr *int64
            if o.PosPayment.AmountReceivedCents != nil {
                v := int64(*o.PosPayment.AmountReceivedCents)
                receivedPtr = &v
            }
            if o.PosPayment.ChangeCents != nil {
                v := int64(*o.PosPayment.ChangeCents)
                changePtr = &v
            }
            dto.PosPayment = &struct {
                Method              string  `json:"method"`
                AmountReceivedCents *int64  `json:"amountReceivedCents,omitempty"`
                ChangeCents         *int64  `json:"changeCents,omitempty"`
            }{ Method: o.PosPayment.Method, AmountReceivedCents: receivedPtr, ChangeCents: changePtr }
        }
        out = append(out, dto)
        if o.Status == domain.OrderStatusPaid { revenue += int64(o.TotalCents) }
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total, "totals": map[string]any{"revenueCents": revenue}})
}

func (h *AdminOrderHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
    // reuse filters
    var status *domain.OrderStatus
    if s := r.URL.Query().Get("status"); s != "" { ss := domain.OrderStatus(s); status = &ss }
    var from, to *time.Time
    if v := r.URL.Query().Get("date_from"); v != "" { if t, err := time.Parse(time.RFC3339, v); err == nil { from = &t } }
    if v := r.URL.Query().Get("date_to"); v != "" { if t, err := time.Parse(time.RFC3339, v); err == nil { to = &t } }
    var q *string
    if v := r.URL.Query().Get("q"); v != "" { q = &v }
    // No pagination for export (but cap with sane limit)
    items, _, err := h.orders.ListAdmin(r.Context(), status, from, to, q, 5000, 0)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to export"); return }
    w.Header().Set("Content-Type", "text/csv; charset=utf-8")
    w.Header().Set("Content-Disposition", "attachment; filename=orders.csv")
    cw := csv.NewWriter(w)
    _ = cw.Write([]string{"id","status","total_cents","created_at"})
    for _, o := range items { _ = cw.Write([]string{o.ID.Hex(), string(o.Status), strconv.FormatInt(int64(o.TotalCents), 10), o.CreatedAt.Format(time.RFC3339)}) }
    cw.Flush()
}

type patchOrderStatus struct { Status domain.OrderStatus `json:"status"` }

func (h *AdminOrderHandler) PatchStatus(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    before, _ := h.orders.FindByID(r.Context(), oid)
    if before == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    var body patchOrderStatus
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.orders.UpdateStatusAndContact(r.Context(), oid, body.Status, nil); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    after, _ := h.orders.FindByID(r.Context(), oid)
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "order", EntityID: id, Before: before, After: after, RequestID: getRequestIDPtr(r), ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role)) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// AdminOrderDetailsDTO returns detailed order including items for admin view
type AdminOrderDetailsDTO struct {
    ID               string             `json:"id"`
    Status           domain.OrderStatus `json:"status"`
    TotalCents       int64              `json:"totalCents"`
    CreatedAt        time.Time          `json:"createdAt"`
    UpdatedAt        time.Time          `json:"updatedAt"`
    ContactEmail     *string            `json:"contactEmail,omitempty"`
    CustomerID       *string            `json:"customerId,omitempty"`
    PaymentIntentID  *string            `json:"paymentIntentId,omitempty"`
    StripeChargeID   *string            `json:"stripeChargeId,omitempty"`
    PaymentAttemptID *string            `json:"paymentAttemptId,omitempty"`
    Items            []PublicOrderItemDTO `json:"items"`
    Origin           *domain.OrderOrigin `json:"origin,omitempty"`
    PosPayment       *struct {
        Method              string  `json:"method"`
        AmountReceivedCents *int64  `json:"amountReceivedCents,omitempty"`
        ChangeCents         *int64  `json:"changeCents,omitempty"`
    } `json:"posPayment,omitempty"`
}

// GET /v1/admin/orders/{id}
func (h *AdminOrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }

    ord, err := h.orders.FindByID(r.Context(), oid)
    if err != nil || ord == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }

    // Load items
    items, err := h.orderItems.FindByOrderID(r.Context(), oid)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to load items"); return }

    // Enrich items with product images (non-sensitive)
    pidSet := map[primitive.ObjectID]struct{}{}
    for _, it := range items { pidSet[it.ProductID] = struct{}{} }
    ids := make([]primitive.ObjectID, 0, len(pidSet))
    for k := range pidSet { ids = append(ids, k) }
    products, _ := h.products.GetByIDs(r.Context(), ids)
    imgByID := map[primitive.ObjectID]*string{}
    for _, p := range products {
        if p.Image != nil && *p.Image != "" { v := *p.Image; imgByID[p.ID] = &v } else { imgByID[p.ID] = nil }
    }

    dtoItems := make([]PublicOrderItemDTO, 0, len(items))
    for _, it := range items {
        var parentID *string
        if it.ParentItemID != nil { s := it.ParentItemID.Hex(); parentID = &s }
        var msID *string
        if it.MenuSlotID != nil { s := it.MenuSlotID.Hex(); msID = &s }
        img := imgByID[it.ProductID]
        dtoItems = append(dtoItems, PublicOrderItemDTO{
            ID: it.ID.Hex(), OrderID: it.OrderID.Hex(), ProductID: it.ProductID.Hex(), Title: it.Title,
            Quantity: it.Quantity, PricePerUnitCents: int64(it.PricePerUnitCents), ParentItemID: parentID,
            MenuSlotID: msID, MenuSlotName: it.MenuSlotName, ProductImage: img,
        })
    }

    var custID *string
    if ord.CustomerID != nil { s := ord.CustomerID.Hex(); custID = &s }
    dto := AdminOrderDetailsDTO{
        ID: ord.ID.Hex(), Status: ord.Status, TotalCents: int64(ord.TotalCents),
        CreatedAt: ord.CreatedAt, UpdatedAt: ord.UpdatedAt,
        ContactEmail: ord.ContactEmail, CustomerID: custID,
        PaymentIntentID: ord.StripePaymentIntentID, StripeChargeID: ord.StripeChargeID, PaymentAttemptID: ord.PaymentAttemptID,
        Items: dtoItems,
        Origin: ord.Origin,
    }
    if ord.PosPayment != nil {
        var receivedPtr, changePtr *int64
        if ord.PosPayment.AmountReceivedCents != nil {
            v := int64(*ord.PosPayment.AmountReceivedCents)
            receivedPtr = &v
        }
        if ord.PosPayment.ChangeCents != nil {
            v := int64(*ord.PosPayment.ChangeCents)
            changePtr = &v
        }
        dto.PosPayment = &struct {
            Method              string  `json:"method"`
            AmountReceivedCents *int64  `json:"amountReceivedCents,omitempty"`
            ChangeCents         *int64  `json:"changeCents,omitempty"`
        }{ Method: ord.PosPayment.Method, AmountReceivedCents: receivedPtr, ChangeCents: changePtr }
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"order": dto})
}
