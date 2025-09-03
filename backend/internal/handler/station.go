package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type StationHandler struct {
	stationService service.StationService
	validator      *validator.Validate
	logger         *zap.Logger
}

func NewStationHandler(stationService service.StationService, logger *zap.Logger) *StationHandler {
	return &StationHandler{
		stationService: stationService,
		validator:      validator.New(),
		logger:         logger,
	}
}

// RequestStation godoc
// @Summary Request to become a station (Public)
// @Description Allow anyone to request to become a station without authentication
// @Tags stations
// @Accept json
// @Produce json
// @Param request body domain.StationRequest true "Station request"
// @Success 201 {object} service.StationResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/stations/request [post]
func (h *StationHandler) RequestStation(w http.ResponseWriter, r *http.Request) {
	var req domain.StationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	stationResponse, err := h.stationService.CreateStation(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create station request", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, stationResponse)
}

// CreateStation godoc
// @Summary Create a new station (Admin only)
// @Description Allow admins to create a new station
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.StationRequest true "Station request"
// @Success 201 {object} service.StationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/stations [post]
func (h *StationHandler) CreateStation(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Only admins can create stations
	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req domain.StationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	stationResponse, err := h.stationService.CreateStation(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create station", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, stationResponse)
}

// ListStations godoc
// @Summary List all stations
// @Description List all stations with optional filtering
// @Tags stations
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(pending, approved, rejected)
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListStationsResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/stations [get]
func (h *StationHandler) ListStations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var stationsResponse *service.ListStationsResponse
	var err error

	if statusStr != "" {
		status := domain.StationStatus(statusStr)
		// Validate status
		switch status {
		case domain.StationStatusPending, domain.StationStatusApproved, domain.StationStatusRejected:
			stationsResponse, err = h.stationService.ListStationsByStatus(r.Context(), status, limit, offset)
		default:
			response.WriteError(w, http.StatusBadRequest, "Invalid status parameter")
			return
		}
	} else {
		stationsResponse, err = h.stationService.ListStations(r.Context(), limit, offset)
	}

	if err != nil {
		h.logger.Error("Failed to list stations", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to list stations")
		return
	}

	response.WriteJSON(w, http.StatusOK, stationsResponse)
}


// GetStation godoc
// @Summary Get station by ID
// @Description Get detailed information about a specific station
// @Tags stations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Success 200 {object} service.StationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/stations/{id} [get]
func (h *StationHandler) GetStation(w http.ResponseWriter, r *http.Request) {
	stationIDStr := chi.URLParam(r, "id")
	stationID, err := primitive.ObjectIDFromHex(stationIDStr)
	if err != nil {
		h.logger.Error("Invalid station ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid station ID")
		return
	}

	stationResponse, err := h.stationService.GetStation(r.Context(), stationID)
	if err != nil {
		h.logger.Error("Failed to get station", zap.Error(err))
		response.WriteError(w, http.StatusNotFound, "Station not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, stationResponse)
}

// ApproveStation godoc
// @Summary Approve or reject a station request (Admin only)
// @Description Allow admins to approve or reject station requests
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Param request body domain.StationApprovalRequest true "Approval request"
// @Success 200 {object} service.ApprovalResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/stations/{id}/approve [put]
func (h *StationHandler) ApproveStation(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Only admins can approve/reject stations
	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	adminID, err := primitive.ObjectIDFromHex(userClaim.Subject)
	if err != nil {
		h.logger.Error("Invalid admin ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid admin ID")
		return
	}

	stationIDStr := chi.URLParam(r, "id")
	stationID, err := primitive.ObjectIDFromHex(stationIDStr)
	if err != nil {
		h.logger.Error("Invalid station ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid station ID")
		return
	}

	var req domain.StationApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var approvalResponse *service.ApprovalResponse
	if req.Approve {
		approvalResponse, err = h.stationService.ApproveStation(r.Context(), stationID, adminID)
	} else {
		reason := ""
		if req.Reason != nil {
			reason = *req.Reason
		}
		approvalResponse, err = h.stationService.RejectStation(r.Context(), stationID, adminID, reason)
	}

	if err != nil {
		h.logger.Error("Failed to process station approval", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to process approval")
		return
	}

	status := http.StatusOK
	if !approvalResponse.Success {
		status = http.StatusBadRequest
	}

	response.WriteJSON(w, status, approvalResponse)
}

// AssignProductsToStation godoc
// @Summary Assign products to a station (Admin only)
// @Description Allow admins to assign products to approved stations
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Param productIds body []string true "Product IDs to assign"
// @Success 200 {object} service.AssignProductsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/stations/{id}/products [post]
func (h *StationHandler) AssignProductsToStation(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Only admins can assign products
	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	stationIDStr := chi.URLParam(r, "id")
	stationID, err := primitive.ObjectIDFromHex(stationIDStr)
	if err != nil {
		h.logger.Error("Invalid station ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid station ID")
		return
	}

	var productIDStrs []string
	if err := json.NewDecoder(r.Body).Decode(&productIDStrs); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert string IDs to ObjectIDs
	productIDs := make([]primitive.ObjectID, len(productIDStrs))
	for i, idStr := range productIDStrs {
		productID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			h.logger.Error("Invalid product ID", zap.String("productID", idStr), zap.Error(err))
			response.WriteError(w, http.StatusBadRequest, "Invalid product ID: "+idStr)
			return
		}
		productIDs[i] = productID
	}

	assignResponse, err := h.stationService.AssignProductsToStation(r.Context(), stationID, productIDs)
	if err != nil {
		h.logger.Error("Failed to assign products to station", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to assign products")
		return
	}

	status := http.StatusOK
	if !assignResponse.Success {
		status = http.StatusBadRequest
	}

	response.WriteJSON(w, status, assignResponse)
}

// GetStationProducts godoc
// @Summary Get products assigned to a station
// @Description Get list of products assigned to a specific station
// @Tags stations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Success 200 {object} service.StationProductsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/stations/{id}/products [get]
func (h *StationHandler) GetStationProducts(w http.ResponseWriter, r *http.Request) {
	stationIDStr := chi.URLParam(r, "id")
	stationID, err := primitive.ObjectIDFromHex(stationIDStr)
	if err != nil {
		h.logger.Error("Invalid station ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid station ID")
		return
	}

	productsResponse, err := h.stationService.GetStationProducts(r.Context(), stationID)
	if err != nil {
		h.logger.Error("Failed to get station products", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to get station products")
		return
	}

	response.WriteJSON(w, http.StatusOK, productsResponse)
}

// RemoveProductFromStation godoc
// @Summary Remove product from station (Admin only)
// @Description Allow admins to remove a product from a station
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Station ID"
// @Param productId path string true "Product ID"
// @Success 200 {object} service.AssignProductsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/stations/{id}/products/{productId} [delete]
func (h *StationHandler) RemoveProductFromStation(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Only admins can remove products
	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	stationIDStr := chi.URLParam(r, "id")
	stationID, err := primitive.ObjectIDFromHex(stationIDStr)
	if err != nil {
		h.logger.Error("Invalid station ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid station ID")
		return
	}

	productIDStr := chi.URLParam(r, "productId")
	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		h.logger.Error("Invalid product ID", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	removeResponse, err := h.stationService.RemoveProductFromStation(r.Context(), stationID, productID)
	if err != nil {
		h.logger.Error("Failed to remove product from station", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "Failed to remove product")
		return
	}

	response.WriteJSON(w, http.StatusOK, removeResponse)
}