package handler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	productRepo   repository.ProductRepository
	logger        *zap.Logger
}

func NewOrderHandler(orderRepo repository.OrderRepository, orderItemRepo repository.OrderItemRepository, productRepo repository.ProductRepository, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, orderItemRepo: orderItemRepo, productRepo: productRepo, logger: logger}
}

// OrderSummaryDTO is a minimal shape for listing orders
type OrderSummaryDTO struct {
	ID        string             `json:"id"`
	Status    domain.OrderStatus `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
}

// ListMyOrders godoc
// @Summary List my orders
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} domain.ListResponse[OrderSummaryDTO]
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/orders [get]
func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok || claims == nil || claims.Subject == "" {
		response.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	limit := 50
	offset := 0
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if os := r.URL.Query().Get("offset"); os != "" {
		if v, err := strconv.Atoi(os); err == nil && v >= 0 {
			offset = v
		}
	}

	userOID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	orders, total, err := h.orderRepo.ListByCustomerID(r.Context(), userOID, limit, offset)
	if err != nil {
		h.logger.Error("list orders failed", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	items := make([]OrderSummaryDTO, 0, len(orders))
	for _, o := range orders {
		items = append(items, OrderSummaryDTO{ID: o.ID.Hex(), Status: o.Status, CreatedAt: o.CreatedAt})
	}

	response.WriteJSON(w, http.StatusOK, domain.ListResponse[OrderSummaryDTO]{
		Items: items,
		Count: int(total),
	})
}

// ----- Public order details (no auth) -----

// PublicOrderItemDTO exposes non-sensitive order item data
type PublicOrderItemDTO struct {
	ID                string  `json:"id"`
	OrderID           string  `json:"orderId"`
	ProductID         string  `json:"productId"`
	Title             string  `json:"title"`
	Quantity          int     `json:"quantity"`
	PricePerUnitCents int64   `json:"pricePerUnitCents"`
	ParentItemID      *string `json:"parentItemId,omitempty"`
	MenuSlotID        *string `json:"menuSlotId,omitempty"`
	MenuSlotName      *string `json:"menuSlotName,omitempty"`
	ProductImage      *string `json:"productImage,omitempty"`
}

// PublicOrderDetailsDTO exposes minimal, non-sensitive order details
type PublicOrderDetailsDTO struct {
	ID         string               `json:"id"`
	Status     domain.OrderStatus   `json:"status"`
	TotalCents int64                `json:"totalCents"`
	CreatedAt  time.Time            `json:"createdAt"`
	Items      []PublicOrderItemDTO `json:"items"`
}

// GetPublicByID godoc
// @Summary Get public order details
// @Description Public, read-only. Does not return customer identifiers.
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} PublicOrderDetailsDTO
// @Failure 400 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/orders/{id} [get]
func (h *OrderHandler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	// Parse {id}
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.WriteError(w, http.StatusBadRequest, "missing order id")
		return
	}
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	ord, err := h.orderRepo.FindByID(r.Context(), oid)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "order not found")
		return
	}
	items, err := h.orderItemRepo.FindByOrderID(r.Context(), oid)
	if err != nil {
		h.logger.Error("find order items", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, "failed to load order items")
		return
	}
	// Enrich with product images (non-sensitive)
	// Collect distinct product IDs
	pidSet := map[primitive.ObjectID]struct{}{}
	for _, it := range items {
		pidSet[it.ProductID] = struct{}{}
	}
	ids := make([]primitive.ObjectID, 0, len(pidSet))
	for id := range pidSet {
		ids = append(ids, id)
	}
	products, _ := h.productRepo.GetByIDs(r.Context(), ids)
	imgByID := map[primitive.ObjectID]*string{}
	for _, p := range products {
		if p.Image != nil && *p.Image != "" {
			// Copy value to avoid aliasing
			v := *p.Image
			imgByID[p.ID] = &v
		} else {
			imgByID[p.ID] = nil
		}
	}
	dtoItems := make([]PublicOrderItemDTO, 0, len(items))
	for _, it := range items {
		var parentID *string
		if it.ParentItemID != nil {
			s := it.ParentItemID.Hex()
			parentID = &s
		}
		var msID *string
		if it.MenuSlotID != nil {
			s := it.MenuSlotID.Hex()
			msID = &s
		}
		msName := it.MenuSlotName
		img := imgByID[it.ProductID]
		dtoItems = append(dtoItems, PublicOrderItemDTO{
			ID: it.ID.Hex(), OrderID: it.OrderID.Hex(), ProductID: it.ProductID.Hex(), Title: it.Title,
			Quantity: it.Quantity, PricePerUnitCents: int64(it.PricePerUnitCents), ParentItemID: parentID,
			MenuSlotID: msID, MenuSlotName: msName, ProductImage: img,
		})
	}
	dto := PublicOrderDetailsDTO{
		ID: ord.ID.Hex(), Status: ord.Status, TotalCents: int64(ord.TotalCents), CreatedAt: ord.CreatedAt, Items: dtoItems,
	}
	response.WriteJSON(w, http.StatusOK, dto)
}
