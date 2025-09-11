package handler

import (
	"encoding/json"
	"net/http"

	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type RedemptionHandler struct {
	orderService service.OrderService
	validator    *validator.Validate
	logger       *zap.Logger
}

func NewRedemptionHandler(orderService service.OrderService, logger *zap.Logger) *RedemptionHandler {
	return &RedemptionHandler{
		orderService: orderService,
		validator:    validator.New(),
		logger:       logger,
	}
}

// GetOrderForRedemption godoc
// @Summary Get order details for redemption at a station
// @Description Get order details and items available for redemption at a specific station (scanned QR code)
// @Tags redemption
// @Produce json
// @Param order_id path string true "Order ID"
// @Param station_id query string true "Station ID"
// @Success 200 {object} service.OrderRedemptionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/redemption/orders/{order_id} [get]
func (h *RedemptionHandler) GetOrderForRedemption(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")
	if orderID == "" {
		response.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	stationID := r.URL.Query().Get("station_id")
	if stationID == "" {
		response.WriteError(w, http.StatusBadRequest, "Station ID is required")
		return
	}

	svcResp, err := h.orderService.GetOrderForRedemption(r.Context(), orderID, stationID)
	if err != nil {
		h.logger.Error("Failed to get order for redemption", zap.Error(err))
		if err.Error() == "order not found" || err.Error() == "station not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// RedeemOrderItems godoc
// @Summary Redeem order items at a station
// @Description Redeem all available order items for a specific station (device confirms redemption)
// @Tags redemption
// @Accept json
// @Produce json
// @Param request body service.RedeemOrderItemsRequest true "Redemption request"
// @Success 200 {object} service.RedeemOrderItemsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/redemption/redeem [post]
func (h *RedemptionHandler) RedeemOrderItems(w http.ResponseWriter, r *http.Request) {
	var req service.RedeemOrderItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	svcResp, err := h.orderService.RedeemOrderItems(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to redeem order items", zap.Error(err))
		if err.Error() == "order not found" || err.Error() == "device not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}