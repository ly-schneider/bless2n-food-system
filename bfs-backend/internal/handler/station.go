package handler

import (
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// StationCreateRequestPayload represents a station verification request.
type StationCreateRequestPayload struct {
	Name      string `json:"name" validate:"required,min=2"`
	Model     string `json:"model"`
	OS        string `json:"os"`
	DeviceKey string `json:"deviceKey" validate:"required,min=8"`
}

// StationVerifyPayload carries a QR code to verify.
type StationVerifyPayload struct {
	Code string `json:"code" validate:"required"`
}

// StationRedeemPayload carries a QR code to redeem.
type StationRedeemPayload struct {
	Code string `json:"code" validate:"required"`
}

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

// CreateRequest godoc
// @Summary Request station verification
// @Tags stations
// @Accept json
// @Produce json
// @Param payload body StationCreateRequestPayload true "Request payload"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/stations/requests [post]
func (h *StationHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var body StationCreateRequestPayload
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

// Me godoc
// @Summary Get current station by key
// @Tags stations
// @Produce json
// @Param X-Station-Key header string true "Station key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/stations/me [get]
func (h *StationHandler) Me(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-Station-Key")
	if key == "" {
		response.WriteError(w, http.StatusBadRequest, "missing station key")
		return
	}
	st, err := h.stationService.GetStationByKey(r.Context(), key)
	if err != nil {
		response.WriteJSON(w, http.StatusOK, map[string]any{"exists": false, "approved": false})
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"exists": true, "approved": st.Approved, "name": st.Name})
}

// VerifyQR godoc
// @Summary Verify pickup QR code
// @Tags stations
// @Accept json
// @Produce json
// @Param X-Station-Key header string true "Station key"
// @Param payload body StationVerifyPayload true "QR payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/stations/verify-qr [post]
func (h *StationHandler) VerifyQR(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.allow("verify:"+r.RemoteAddr, 10, 10) {
		response.WriteError(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	key := r.Header.Get("X-Station-Key")
	if key == "" {
		response.WriteError(w, http.StatusBadRequest, "missing station key")
		return
	}
	st, err := h.stationService.GetStationByKey(r.Context(), key)
	if err != nil || st == nil || !st.Approved {
		response.WriteError(w, http.StatusForbidden, "station not approved")
		return
	}
	var body StationVerifyPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.validator.Struct(body); err != nil {
		response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
		return
	}
	oid, err := h.stationService.VerifyQR(r.Context(), body.Code)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Return assigned items preview (normalized + image)
	items, err := h.stationService.AssignedItemsForOrder(r.Context(), st.ID, oid)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to load items")
		return
	}
	// product images
	pidSet := map[bson.ObjectID]struct{}{}
	for _, it := range items {
		pidSet[it.ProductID] = struct{}{}
	}
	ids := make([]bson.ObjectID, 0, len(pidSet))
	for id := range pidSet {
		ids = append(ids, id)
	}
	products, _ := h.productRepo.GetByIDs(r.Context(), ids)
	imgByID := map[bson.ObjectID]*string{}
	for _, p := range products {
		if p.Image != nil && *p.Image != "" {
			v := *p.Image
			imgByID[p.ID] = &v
		} else {
			imgByID[p.ID] = nil
		}
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
			if ex.PricePerUnitCents == 0 && it.PricePerUnitCents > 0 {
				ex.PricePerUnitCents = it.PricePerUnitCents
			}
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
	for _, v := range byPID {
		out = append(out, v)
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"orderId": oid.Hex(), "items": out})
}

// Redeem godoc
// @Summary Redeem assigned items
// @Description Idempotent. Requires station key.
// @Tags stations
// @Accept json
// @Produce json
// @Param X-Station-Key header string true "Station key"
// @Param Idempotency-Key header string false "Idempotency key"
// @Param payload body StationRedeemPayload true "QR payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/stations/redeem [post]
func (h *StationHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.allow("redeem:"+r.RemoteAddr, 5, 10) {
		response.WriteError(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	key := r.Header.Get("X-Station-Key")
	if key == "" {
		response.WriteError(w, http.StatusBadRequest, "missing station key")
		return
	}
	st, err := h.stationService.GetStationByKey(r.Context(), key)
	if err != nil || st == nil || !st.Approved {
		response.WriteError(w, http.StatusForbidden, "station not approved")
		return
	}
	var body StationRedeemPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.validator.Struct(body); err != nil {
		response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
		return
	}
	oid, err := h.stationService.VerifyQR(r.Context(), body.Code)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	idem := r.Header.Get("Idempotency-Key")
	resp, err := h.stationService.RedeemAssigned(r.Context(), st.ID, oid, idem)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "redeem failed")
		return
	}
	// If nothing redeemed, return conflict with details
	if v, ok := resp["redeemed"].(int64); ok && v == 0 {
		response.WriteJSON(w, http.StatusConflict, map[string]any{"message": "no items to redeem", "details": resp})
		return
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

// simple sliding window limiter reused from auth service
type rateLimiter struct{ buckets map[string][]int64 }

func newRateLimiter() *rateLimiter { return &rateLimiter{buckets: make(map[string][]int64)} }

// allow max events within window seconds
func (r *rateLimiter) allow(key string, max int, windowSec int) bool {
	now := time.Now().Unix()
	xs := r.buckets[key]
	// drop older than window
	cutoff := now - int64(windowSec)
	i := 0
	for ; i < len(xs); i++ {
		if xs[i] > cutoff {
			break
		}
	}
	xs = xs[i:]
	if len(xs) >= max {
		r.buckets[key] = xs
		return false
	}
	xs = append(xs, now)
	r.buckets[key] = xs
	return true
}

// GetPickupQR godoc
// @Summary Get pickup QR for order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/orders/{id}/pickup-qr [get]
func (h *StationHandler) GetPickupQR(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.WriteError(w, http.StatusBadRequest, "missing id")
		return
	}
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	code, err := h.stationService.MakePickupQR(r.Context(), oid)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate qr")
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"code": code})
}
