package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type ProductHandler struct {
	productService service.ProductService
	validator      *validator.Validate
	logger         *zap.Logger
}

func NewProductHandler(productService service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		validator:      validator.New(),
		logger:         logger,
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with specified category, type, name, image, and price
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateProductRequest true "Product creation payload"
// @Success 201 {object} service.CreateProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req service.CreateProductRequest
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

	svcResp, err := h.productService.CreateProduct(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Retrieve detailed information about a specific product
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} service.GetProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	svcResp, err := h.productService.GetProduct(r.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to get product", zap.Error(err))
		if err.Error() == "product not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update product details including category, name, image, and price
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body service.UpdateProductRequest true "Product update payload"
// @Success 200 {object} service.UpdateProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req service.UpdateProductRequest
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

	svcResp, err := h.productService.UpdateProduct(r.Context(), productID, req)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Error(err))
		if err.Error() == "product not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Remove a product from the system
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} service.DeleteProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	svcResp, err := h.productService.DeleteProduct(r.Context(), productID)
	if err != nil {
		h.logger.Error("Failed to delete product", zap.Error(err))
		if err.Error() == "product not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// ListProducts godoc
// @Summary List products
// @Description List all products with optional category filtering
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param category_id query string false "Filter by category ID"
// @Param active_only query boolean false "Show only active products" default(false)
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListProductsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeOnlyStr := r.URL.Query().Get("active_only")

	limit := 50
	offset := 0
	activeOnly := false

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

	if activeOnlyStr != "" {
		activeOnly = activeOnlyStr == "true"
	}

	var categoryID *string
	if categoryIDStr != "" {
		categoryID = &categoryIDStr
	}

	svcResp, err := h.productService.ListProducts(r.Context(), categoryID, activeOnly, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}


// SetProductActive godoc
// @Summary Set product active status
// @Description Activate or deactivate a product
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param active query boolean true "Active status" default(true)
// @Success 200 {object} service.SetProductActiveResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id}/status [put]
func (h *ProductHandler) SetProductActive(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	activeStr := r.URL.Query().Get("active")
	if activeStr == "" {
		response.WriteError(w, http.StatusBadRequest, "Active status is required")
		return
	}

	isActive := activeStr == "true"

	svcResp, err := h.productService.SetProductActive(r.Context(), productID, isActive)
	if err != nil {
		h.logger.Error("Failed to set product active status", zap.Error(err))
		if err.Error() == "product not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// UpdateProductStock godoc
// @Summary Update product stock
// @Description Update product inventory stock levels
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body service.UpdateProductStockRequest true "Stock update payload"
// @Success 200 {object} service.UpdateProductStockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id}/stock [put]
func (h *ProductHandler) UpdateProductStock(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var req service.UpdateProductStockRequest
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

	svcResp, err := h.productService.UpdateProductStock(r.Context(), productID, req)
	if err != nil {
		h.logger.Error("Failed to update product stock", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// CreateProductBundle godoc
// @Summary Create a new product bundle (menu)
// @Description Create a new product bundle with multiple component products
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateProductBundleRequest true "Bundle creation payload"
// @Success 201 {object} service.CreateProductBundleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/products/bundles [post]
func (h *ProductHandler) CreateProductBundle(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req service.CreateProductBundleRequest
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

	svcResp, err := h.productService.CreateProductBundle(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create product bundle", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

// UpdateProductBundle godoc
// @Summary Update product bundle
// @Description Update an existing product bundle and its components
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bundle ID"
// @Param request body service.UpdateProductBundleRequest true "Bundle update payload"
// @Success 200 {object} service.UpdateProductBundleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/bundles/{id} [put]
func (h *ProductHandler) UpdateProductBundle(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	bundleID := chi.URLParam(r, "id")
	if bundleID == "" {
		response.WriteError(w, http.StatusBadRequest, "Bundle ID is required")
		return
	}

	var req service.UpdateProductBundleRequest
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

	svcResp, err := h.productService.UpdateProductBundle(r.Context(), bundleID, req)
	if err != nil {
		h.logger.Error("Failed to update product bundle", zap.Error(err))
		if err.Error() == "bundle not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}


// AssignProductToStations godoc
// @Summary Assign product to stations
// @Description Assign a product to multiple stations for availability
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param stationIds body []string true "Station IDs to assign the product to"
// @Success 200 {object} service.AssignProductToStationsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/products/{id}/stations [post]
func (h *ProductHandler) AssignProductToStations(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	productID := chi.URLParam(r, "id")
	if productID == "" {
		response.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var stationIDStrs []string
	if err := json.NewDecoder(r.Body).Decode(&stationIDStrs); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	stationIDs := make([]primitive.ObjectID, len(stationIDStrs))
	for i, idStr := range stationIDStrs {
		stationID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			h.logger.Error("Invalid station ID", zap.String("stationID", idStr), zap.Error(err))
			response.WriteError(w, http.StatusBadRequest, "Invalid station ID: "+idStr)
			return
		}
		stationIDs[i] = stationID
	}

	svcResp, err := h.productService.AssignProductToStations(r.Context(), productID, stationIDs)
	if err != nil {
		h.logger.Error("Failed to assign product to stations", zap.Error(err))
		if err.Error() == "product not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}