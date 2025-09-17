package handler

import (
	"net/http"
	"strconv"

	_ "backend/internal/domain"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
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

// ListProducts godoc
// @Summary List products
// @Description List all products with optional category filtering
// @Tags products
// @Produce json
// @Param category_id query string false "Filter by category ID"
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} domain.ListResponse[domain.ProductDTO]
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := r.URL.Query().Get("category_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

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

	var categoryID *string
	if categoryIDStr != "" {
		categoryID = &categoryIDStr
	}

	svcResp, err := h.productService.ListProducts(r.Context(), categoryID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}
