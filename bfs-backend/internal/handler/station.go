package handler

import (
    "backend/internal/service"
    "backend/internal/response"
    "backend/internal/repository"
    "encoding/json"
    "net/http"
    "time"
    "github.com/go-chi/chi/v5"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "github.com/go-playground/validator/v10"
    "go.uber.org/zap"
)

type StationHandler struct {
    stationService service.StationService
    productRepo    repository.ProductRepository
    validator      *validator.Validate
    logger         *zap.Logger
    limiter        *rateLimiter
}

func NewStationHandler(stationService service.StationService, productRepo repository.ProductRepository, logger *zap.Logger) *StationHandler {
    return &StationHandler{
        stationService: stationService,
        productRepo:    productRepo,
        validator:      validator.New(),
        logger:         logger,
        limiter:        newRateLimiter(),
    }
}

// POST /v1/stations/requests - anonymous station verification request (inline)
// Body: { name: string, model?: string, os?: string, deviceKey: string }
func (h *StationHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Name      string `json:"name" validate:"required,min=2"`
        Model     string `json:"model"`
        OS        string `json:"os"`
        DeviceKey string `json:"deviceKey" validate:"required,min=8"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        response.WriteError(w, http.StatusBadRequest, "invalid json")
        return
    }
    if err := h.validator.Struct(body); err != nil {
        response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
        return
    }
    st, err := h.stationService.RequestVerification(r.Context(), body.Name, body.Model, body.OS, body.DeviceKey)
    if err != nil {
        h.logger.Error("station request failed", zap.Error(err))
        response.WriteError(w, http.StatusInternalServerError, "failed to submit request")
        return
    }
    response.WriteJSON(w, http.StatusAccepted, map[string]any{"status": "pending", "station": map[string]any{"approved": st.Approved}})
}

// GET /v1/stations/me - public; identify station by header X-Station-Key
func (h *StationHandler) Me(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("X-Station-Key")
    if key == "" { response.WriteError(w, http.StatusBadRequest, "missing station key"); return }
    st, err := h.stationService.GetStationByKey(r.Context(), key)
    if err != nil {
        response.WriteJSON(w, http.StatusOK, map[string]any{"exists": false, "approved": false})
        return
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"exists": true, "approved": st.Approved, "name": st.Name})
}

// POST /v1/stations/verify-qr - validate QR payload and return orderId if valid
// Body: { code: string }
func (h *StationHandler) VerifyQR(w http.ResponseWriter, r *http.Request) {
    if !h.limiter.allow("verify:"+r.RemoteAddr, 10, 10) { response.WriteError(w, http.StatusTooManyRequests, "rate_limited"); return }
    key := r.Header.Get("X-Station-Key")
    if key == "" { response.WriteError(w, http.StatusBadRequest, "missing station key"); return }
    st, err := h.stationService.GetStationByKey(r.Context(), key)
    if err != nil || st == nil || !st.Approved { response.WriteError(w, http.StatusForbidden, "station not approved"); return }
    var body struct{ Code string `json:"code" validate:"required"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    oid, err := h.stationService.VerifyQR(r.Context(), body.Code)
    if err != nil { response.WriteError(w, http.StatusBadRequest, err.Error()); return }
    // Return assigned items preview (normalized + image)
    items, err := h.stationService.AssignedItemsForOrder(r.Context(), st.ID, oid)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to load items"); return }
    // product images
    pidSet := map[primitive.ObjectID]struct{}{}
    for _, it := range items { pidSet[it.ProductID] = struct{}{} }
    ids := make([]primitive.ObjectID, 0, len(pidSet))
    for id := range pidSet { ids = append(ids, id) }
    products, _ := h.productRepo.GetByIDs(r.Context(), ids)
    imgByID := map[primitive.ObjectID]*string{}
    for _, p := range products {
        if p.Image != nil && *p.Image != "" { v := *p.Image; imgByID[p.ID] = &v } else { imgByID[p.ID] = nil }
    }
    // DTO
    type StationItemDTO struct {
        ID                string     `json:"id"`
        OrderID           string     `json:"orderId"`
        ProductID         string     `json:"productId"`
        Title             string     `json:"title"`
        Quantity          int        `json:"quantity"`
        PricePerUnitCents int64      `json:"pricePerUnitCents"`
        ProductImage      *string    `json:"productImage,omitempty"`
        IsRedeemed        bool       `json:"isRedeemed"`
        RedeemedAt        *time.Time `json:"redeemedAt,omitempty"`
    }
    // Build initial flat list
    flat := make([]StationItemDTO, 0, len(items))
    for _, it := range items {
        img := imgByID[it.ProductID]
        flat = append(flat, StationItemDTO{
            ID: it.ID.Hex(), OrderID: it.OrderID.Hex(), ProductID: it.ProductID.Hex(), Title: it.Title,
            Quantity: it.Quantity, PricePerUnitCents: int64(it.PricePerUnitCents), ProductImage: img, IsRedeemed: it.IsRedeemed, RedeemedAt: it.RedeemedAt,
        })
    }
    // Group by productId for presentation: sum quantity, keep image/title, aggregate redeemed=false if any false
    byPID := map[string]StationItemDTO{}
    for _, it := range flat {
        k := it.ProductID
        if ex, ok := byPID[k]; ok {
            ex.Quantity += it.Quantity
            // prefer non-zero price per unit if available
            if ex.PricePerUnitCents == 0 && it.PricePerUnitCents > 0 { ex.PricePerUnitCents = it.PricePerUnitCents }
            // any unredeemed makes aggregated item unredeemed
            ex.IsRedeemed = ex.IsRedeemed && it.IsRedeemed
            byPID[k] = ex
        } else {
            // use productId as row id for stable keys on client
            it.ID = it.ProductID
            byPID[k] = it
        }
    }
    out := make([]StationItemDTO, 0, len(byPID))
    for _, v := range byPID { out = append(out, v) }
    response.WriteJSON(w, http.StatusOK, map[string]any{"orderId": oid.Hex(), "items": out})
}

// POST /v1/stations/redeem - idempotent redemption for assigned items
// Headers: X-Station-Key, Idempotency-Key (optional)
// Body: { code: string }
func (h *StationHandler) Redeem(w http.ResponseWriter, r *http.Request) {
    if !h.limiter.allow("redeem:"+r.RemoteAddr, 5, 10) { response.WriteError(w, http.StatusTooManyRequests, "rate_limited"); return }
    key := r.Header.Get("X-Station-Key")
    if key == "" { response.WriteError(w, http.StatusBadRequest, "missing station key"); return }
    st, err := h.stationService.GetStationByKey(r.Context(), key)
    if err != nil || st == nil || !st.Approved { response.WriteError(w, http.StatusForbidden, "station not approved"); return }
    var body struct{ Code string `json:"code" validate:"required"` }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    oid, err := h.stationService.VerifyQR(r.Context(), body.Code)
    if err != nil { response.WriteError(w, http.StatusBadRequest, err.Error()); return }
    idem := r.Header.Get("Idempotency-Key")
    resp, err := h.stationService.RedeemAssigned(r.Context(), st.ID, oid, idem)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "redeem failed"); return }
    // If nothing redeemed, return conflict with details
    if v, ok := resp["redeemed"].(int64); ok && v == 0 {
        response.WriteJSON(w, http.StatusConflict, map[string]any{"message": "no items to redeem", "details": resp})
        return
    }
    response.WriteJSON(w, http.StatusOK, resp)
}

// simple sliding window limiter reused from auth service
type rateLimiter struct{ buckets map[string][]int64 }
func newRateLimiter() *rateLimiter { return &rateLimiter{ buckets: make(map[string][]int64) } }
// allow max events within window seconds
func (r *rateLimiter) allow(key string, max int, windowSec int) bool {
    now := time.Now().Unix()
    xs := r.buckets[key]
    // drop older than window
    cutoff := now - int64(windowSec)
    i := 0
    for ; i < len(xs); i++ { if xs[i] > cutoff { break } }
    xs = xs[i:]
    if len(xs) >= max { r.buckets[key] = xs; return false }
    xs = append(xs, now)
    r.buckets[key] = xs
    return true
}

// GET /v1/orders/{id}/pickup-qr - public signed QR payload
func (h *StationHandler) GetPickupQR(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    if idStr == "" { response.WriteError(w, http.StatusBadRequest, "missing id"); return }
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    code, err := h.stationService.MakePickupQR(r.Context(), oid)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to generate qr"); return }
    response.WriteJSON(w, http.StatusOK, map[string]any{"code": code})
}
