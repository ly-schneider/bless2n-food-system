package handler

import (
    "backend/internal/domain"
    "backend/internal/response"
    "backend/internal/service"
    "encoding/json"
    "net/http"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/go-chi/chi/v5"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// POSHandler exposes minimal POS device and checkout endpoints.
type POSHandler struct {
    pos    service.POSService
    validator *validator.Validate
}

func NewPOSHandler(pos service.POSService) *POSHandler {
    return &POSHandler{ pos: pos, validator: validator.New() }
}

// POSCreateRequestPayload represents a POS access request.
type POSCreateRequestPayload struct {
    Name        string `json:"name" validate:"required,min=2"`
    Model       string `json:"model"`
    OS          string `json:"os"`
    DeviceToken string `json:"deviceToken" validate:"required,min=8"`
}

// POSCreateOrderPayload is the request body for creating a POS order.
type POSCreateOrderPayload struct {
    Items         []service.CreateCheckoutInputItem `json:"items" validate:"required,min=1,dive"`
    CustomerEmail *string              `json:"customerEmail,omitempty"`
}

// POSPayCashPayload is the request body for cash payment.
type POSPayCashPayload struct {
    AmountReceivedCents int64 `json:"amountReceivedCents" validate:"required,gte=0"`
}

// POSPayCardPayload is the request body for card payment result.
type POSPayCardPayload struct {
    Processor     string  `json:"processor" validate:"required"`
    TransactionID *string `json:"transactionId,omitempty"`
    Status        string  `json:"status" validate:"required,oneof=succeeded failed canceled"`
}

// CreateRequest godoc
// @Summary Request POS device access
// @Tags pos
// @Accept json
// @Produce json
// @Param payload body POSCreateRequestPayload true "Request payload"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/pos/requests [post]
func (h *POSHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
    var body POSCreateRequestPayload
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        response.WriteError(w, http.StatusBadRequest, "invalid json")
        return
    }
    if err := h.validator.Struct(body); err != nil {
        response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path))
        return
    }
    if err := h.pos.RequestAccess(r.Context(), body.Name, body.Model, body.OS, body.DeviceToken); err != nil {
        response.WriteError(w, http.StatusInternalServerError, "failed to submit request")
        return
    }
    response.WriteJSON(w, http.StatusAccepted, map[string]any{"status": "pending"})
}

// Me godoc
// @Summary Get current POS device by token
// @Tags pos
// @Produce json
// @Param X-Pos-Token header string true "POS token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/pos/me [get]
func (h *POSHandler) Me(w http.ResponseWriter, r *http.Request) {
    token := strings.TrimSpace(r.Header.Get("X-Pos-Token"))
    if token == "" { response.WriteError(w, http.StatusBadRequest, "missing pos token"); return }
    dev, err := h.pos.GetDeviceByToken(r.Context(), token)
    if err != nil || dev == nil {
        response.WriteJSON(w, http.StatusOK, map[string]any{"exists": false, "approved": false})
        return
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"exists": true, "approved": dev.Approved, "name": dev.Name, "cardCapable": dev.CardCapable})
}

// CreateOrder godoc
// @Summary Create POS order
// @Tags pos
// @Accept json
// @Produce json
// @Param X-Pos-Token header string true "POS token"
// @Param payload body POSCreateOrderPayload true "Order payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/pos/orders [post]
func (h *POSHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    token := strings.TrimSpace(r.Header.Get("X-Pos-Token"))
    if token == "" { response.WriteError(w, http.StatusBadRequest, "missing pos token"); return }
    dev, err := h.pos.GetDeviceByToken(r.Context(), token)
    if err != nil || dev == nil || !dev.Approved { response.WriteError(w, http.StatusForbidden, "pos device not approved"); return }
    var body POSCreateOrderPayload
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    orderID, err := h.pos.CreateOrder(r.Context(), body.Items, body.CustomerEmail)
    if err != nil { response.WriteError(w, http.StatusBadRequest, err.Error()); return }
    response.WriteJSON(w, http.StatusOK, map[string]string{"orderId": orderID.Hex()})
}

// PayCash godoc
// @Summary Pay POS order in cash
// @Tags pos
// @Accept json
// @Produce json
// @Param X-Pos-Token header string true "POS token"
// @Param id path string true "Order ID"
// @Param payload body POSPayCashPayload true "Payment payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/pos/orders/{id}/pay-cash [post]
func (h *POSHandler) PayCash(w http.ResponseWriter, r *http.Request) {
    token := strings.TrimSpace(r.Header.Get("X-Pos-Token"))
    if token == "" { response.WriteError(w, http.StatusBadRequest, "missing pos token"); return }
    dev, err := h.pos.GetDeviceByToken(r.Context(), token)
    if err != nil || dev == nil || !dev.Approved { response.WriteError(w, http.StatusForbidden, "pos device not approved"); return }
    var body POSPayCashPayload
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    // Parse order id from path
    idStr := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    change, err := h.pos.PayCash(r.Context(), oid, domain.Cents(body.AmountReceivedCents))
    if err != nil {
        if strings.Contains(err.Error(), "insufficient_cash") { response.WriteError(w, http.StatusBadRequest, "insufficient_cash"); return }
        response.WriteError(w, http.StatusBadRequest, err.Error()); return
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"status": "paid", "changeCents": int64(change)})
}

// PayCard godoc
// @Summary Attach POS card payment result
// @Tags pos
// @Accept json
// @Produce json
// @Param X-Pos-Token header string true "POS token"
// @Param id path string true "Order ID"
// @Param payload body POSPayCardPayload true "Payment payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/pos/orders/{id}/pay-card [post]
func (h *POSHandler) PayCard(w http.ResponseWriter, r *http.Request) {
    token := strings.TrimSpace(r.Header.Get("X-Pos-Token"))
    if token == "" { response.WriteError(w, http.StatusBadRequest, "missing pos token"); return }
    dev, err := h.pos.GetDeviceByToken(r.Context(), token)
    if err != nil || dev == nil || !dev.Approved { response.WriteError(w, http.StatusForbidden, "pos device not approved"); return }
    var body POSPayCardPayload
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil { response.WriteError(w, http.StatusBadRequest, "invalid json"); return }
    if err := h.validator.Struct(body); err != nil { response.WriteProblem(w, response.NewValidationProblem(response.ConvertValidationErrors(err.(validator.ValidationErrors)), r.URL.Path)); return }
    // Parse id
    idStr := chi.URLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if err := h.pos.PayCard(r.Context(), oid, body.Processor, body.TransactionID, body.Status); err != nil {
        response.WriteError(w, http.StatusBadRequest, err.Error()); return
    }
    out := map[string]any{"status": body.Status}
    if body.Status == "succeeded" { out["orderStatus"] = "paid" }
    response.WriteJSON(w, http.StatusOK, out)
}
