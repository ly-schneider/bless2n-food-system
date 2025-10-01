package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/response"
    "backend/internal/repository"
    "net/http"
    "strconv"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.uber.org/zap"
)

type OrderHandler struct {
    orderRepo repository.OrderRepository
    logger    *zap.Logger
}

func NewOrderHandler(orderRepo repository.OrderRepository, logger *zap.Logger) *OrderHandler {
    return &OrderHandler{orderRepo: orderRepo, logger: logger}
}

// OrderSummaryDTO is a minimal shape for listing orders
type OrderSummaryDTO struct {
    ID        string             `json:"id"`
    Status    domain.OrderStatus `json:"status"`
    CreatedAt time.Time          `json:"createdAt"`
}

// GET /v1/orders - list orders for the authenticated user
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

