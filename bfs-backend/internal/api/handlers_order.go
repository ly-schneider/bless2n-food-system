package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent/order"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

const (
	idempotencyScopeOrder   = "order_create"
	idempotencyScopePayment = "order_payment"
	idempotencyTTL          = 24 * time.Hour
)

// ListOrders returns orders with optional filtering.
// Admins see all orders; regular users see only their own.
// (GET /orders)
func (h *Handlers) ListOrders(w http.ResponseWriter, r *http.Request, params generated.ListOrdersParams) {
	ctx := r.Context()

	role, _ := auth.GetUserRole(ctx)
	scopeMine := params.Scope != nil && *params.Scope == generated.Mine

	if role != string(auth.RoleAdmin) || scopeMine {
		uid, ok := auth.GetUserID(ctx)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}
		rows, _, err := h.orders.ListByCustomerID(ctx, uid, 50, 0)
		if err != nil {
			writeEntError(w, err)
			return
		}
		items := make([]generated.Order, 0, len(rows))
		for _, o := range rows {
			items = append(items, toAPIOrder(o))
		}
		response.WriteJSON(w, http.StatusOK, generated.OrderList{Items: items})
		return
	}

	listParams := service.OrderListParams{
		Limit:  50,
		Offset: 0,
	}
	if params.Status != nil {
		s := order.Status(*params.Status)
		listParams.Status = &s
	}
	if params.DateFrom != nil {
		fromStr := params.DateFrom.Format("2006-01-02T15:04:05Z07:00")
		listParams.From = &fromStr
	}
	if params.DateTo != nil {
		toStr := params.DateTo.Format("2006-01-02T15:04:05Z07:00")
		listParams.To = &toStr
	}

	orders, _, err := h.orders.ListAdmin(ctx, listParams)
	if err != nil {
		writeEntError(w, err)
		return
	}

	items := make([]generated.Order, 0, len(orders))
	for _, o := range orders {
		items = append(items, toAPIOrder(o))
	}
	response.WriteJSON(w, http.StatusOK, generated.OrderList{Items: items})
}

// CreateOrder creates a new order from cart items.
// (POST /orders)
func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request, params generated.CreateOrderParams) {
	ctx := r.Context()

	var idempotencyKey string
	if params.IdempotencyKey != nil {
		idempotencyKey = *params.IdempotencyKey
	}
	if idempotencyKey != "" {
		if cached, err := h.idempotency.Get(ctx, idempotencyScopeOrder, idempotencyKey); err == nil && cached != nil {
			respMap, _ := repository.GetResponseMap(cached)
			if respMap != nil {
				response.WriteJSON(w, http.StatusOK, respMap)
				return
			}
		}
	}

	var body generated.OrderCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if len(body.Items) == 0 {
		writeError(w, http.StatusBadRequest, "no_items", "At least one item is required")
		return
	}

	checkoutItems := make([]service.CheckoutItemInput, 0, len(body.Items))
	for _, item := range body.Items {
		ci := service.CheckoutItemInput{
			ProductID: item.ProductId.String(),
			Quantity:  item.Quantity,
		}
		if item.MenuSelections != nil {
			ci.Configuration = make(map[string]string, len(*item.MenuSelections))
			for _, sel := range *item.MenuSelections {
				ci.Configuration[sel.SlotId.String()] = sel.ProductId.String()
			}
		}
		checkoutItems = append(checkoutItems, ci)
	}

	var customerEmail *string
	if body.ContactEmail != nil {
		email := string(*body.ContactEmail)
		customerEmail = &email
	}

	origin := order.OriginShop
	if _, isDevice := auth.GetDeviceID(ctx); isDevice {
		origin = order.OriginPos
	}

	input := service.CreateCheckoutInput{
		Items:         checkoutItems,
		CustomerEmail: customerEmail,
		Origin:        origin,
	}

	var userID *string
	if origin != order.OriginPos {
		if uid, ok := auth.GetUserID(ctx); ok {
			userID = &uid
		}
	}

	prep, err := h.payments.PrepareAndCreateOrder(ctx, input, userID, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "order_failed", err.Error())
		return
	}

	createdOrder, err := h.orders.GetByID(ctx, prep.OrderID)
	if err != nil {
		writeEntError(w, err)
		return
	}

	apiOrder := toAPIOrder(createdOrder)

	if idempotencyKey != "" {
		respMap := map[string]any{
			"id":         apiOrder.Id.String(),
			"status":     string(apiOrder.Status),
			"totalCents": apiOrder.TotalCents,
			"createdAt":  apiOrder.CreatedAt.Format(time.RFC3339),
			"origin":     string(apiOrder.Origin),
		}
		if _, err := h.idempotency.SaveIfAbsent(ctx, idempotencyScopeOrder, idempotencyKey, respMap, idempotencyTTL); err != nil {
			h.logger.Warn("failed to cache idempotent order response", zap.Error(err))
		}
	}

	response.WriteJSON(w, http.StatusCreated, apiOrder)
}

// GetOrder returns a single order by ID with lines, payments, and redemptions.
// (GET /orders/{orderId})
func (h *Handlers) GetOrder(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID) {
	o, err := h.orders.GetByIDWithRelations(r.Context(), uuid.UUID(orderId))
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIOrder(o))
}

// UpdateOrderStatus updates the status of an order.
// (PATCH /orders/{orderId})
func (h *Handlers) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID) {
	var body generated.OrderStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	ctx := r.Context()
	id := uuid.UUID(orderId)

	// After Task 6, UpdateStatus takes order.Status from the ent package.
	if err := h.orders.UpdateStatus(ctx, id, order.Status(body.Status)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_transition", err.Error())
		return
	}

	updated, err := h.orders.GetByID(ctx, id)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPIOrder(updated))
}

// GetOrderPayment returns the payment details for an order.
// (GET /orders/{orderId}/payment)
func (h *Handlers) GetOrderPayment(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID) {
	ctx := r.Context()

	o, err := h.orders.GetByID(ctx, uuid.UUID(orderId))
	if err != nil {
		writeEntError(w, err)
		return
	}

	if o.PayrexxGatewayID == nil {
		if o.Status == order.StatusPaid {
			response.WriteJSON(w, http.StatusOK, map[string]any{
				"orderId": o.ID.String(),
				"status":  "paid",
			})
			return
		}
		writeError(w, http.StatusNotFound, "no_payment", "No payment gateway associated with this order")
		return
	}

	gw, err := h.payments.GetPayrexxGateway(ctx, *o.PayrexxGatewayID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "gateway_error", "Could not fetch payment details")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"orderId":     o.ID.String(),
		"gatewayId":   gw.ID,
		"status":      gw.Status,
		"link":        gw.Link,
		"referenceId": gw.ReferenceID,
	})
}

// CreateOrderPayment creates a payment for an order.
// (POST /orders/{orderId}/payment)
func (h *Handlers) CreateOrderPayment(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID, params generated.CreateOrderPaymentParams) {
	ctx := r.Context()
	id := uuid.UUID(orderId)

	var idempotencyKey string
	if params.IdempotencyKey != nil {
		idempotencyKey = *params.IdempotencyKey
	}
	if idempotencyKey != "" {
		scopeKey := idempotencyScopePayment + ":" + id.String()
		if cached, err := h.idempotency.Get(ctx, scopeKey, idempotencyKey); err == nil && cached != nil {
			respMap, _ := repository.GetResponseMap(cached)
			if respMap != nil {
				response.WriteJSON(w, http.StatusOK, respMap)
				return
			}
		}
	}

	var body generated.PaymentCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	cachePaymentResponse := func(resp map[string]any) {
		if idempotencyKey != "" {
			scopeKey := idempotencyScopePayment + ":" + id.String()
			if _, err := h.idempotency.SaveIfAbsent(ctx, scopeKey, idempotencyKey, resp, idempotencyTTL); err != nil {
				h.logger.Warn("failed to cache idempotent payment response", zap.Error(err))
			}
		}
	}

	switch body.Method {
	case generated.Cash:
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if body.Club100 != nil {
			if err := h.club100.RecordRedemption(ctx, body.Club100.ElvantoPersonId, body.Club100.ElvantoPersonName, id, body.Club100.FreeQuantity); err != nil {
				h.logger.Warn("failed to record club100 redemption for cash payment", zap.Error(err))
			}
		}
		if err := h.pos.PayCash(ctx, id, deviceID); err != nil {
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "cash"}
		if body.Club100 != nil {
			resp["club100PersonId"] = body.Club100.ElvantoPersonId
		}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.Card:
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if body.Club100 != nil {
			if err := h.club100.RecordRedemption(ctx, body.Club100.ElvantoPersonId, body.Club100.ElvantoPersonName, id, body.Club100.FreeQuantity); err != nil {
				h.logger.Warn("failed to record club100 redemption for card payment", zap.Error(err))
			}
		}
		if err := h.pos.PayCard(ctx, id, deviceID); err != nil {
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "card"}
		if body.Club100 != nil {
			resp["club100PersonId"] = body.Club100.ElvantoPersonId
		}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.Twint:
		if body.Channel != nil && *body.Channel == generated.PaymentChannelPos {
			var deviceID *uuid.UUID
			if did, ok := auth.GetDeviceID(ctx); ok {
				deviceID = &did
			}
			if body.Club100 != nil {
				if err := h.club100.RecordRedemption(ctx, body.Club100.ElvantoPersonId, body.Club100.ElvantoPersonName, id, body.Club100.FreeQuantity); err != nil {
					h.logger.Warn("failed to record club100 redemption for twint payment", zap.Error(err))
				}
			}
			if err := h.pos.PayTwint(ctx, id, deviceID); err != nil {
				writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
				return
			}
			resp := map[string]any{"orderId": id.String(), "method": "twint", "channel": "pos"}
			if body.Club100 != nil {
				resp["club100PersonId"] = body.Club100.ElvantoPersonId
			}
			cachePaymentResponse(resp)
			response.WriteJSON(w, http.StatusCreated, resp)
			return
		}

		o, err := h.orders.GetByID(ctx, id)
		if err != nil {
			writeEntError(w, err)
			return
		}

		returnURL := ""
		if body.ReturnUrl != nil {
			returnURL = *body.ReturnUrl
		}

		if !h.payments.IsPayrexxEnabled() {
			h.logger.Warn("Payrexx not configured — simulating payment for dev",
				zap.String("orderId", id.String()))
			if err := h.payments.MarkOrderPaidDev(ctx, id); err != nil {
				writeError(w, http.StatusInternalServerError, "dev_pay_failed", err.Error())
				return
			}
			redirectURL := returnURL
			if redirectURL != "" {
				redirectURL += "?order_id=" + id.String()
			}
			resp := map[string]any{"orderId": id.String(), "method": "twint", "redirectUrl": redirectURL}
			cachePaymentResponse(resp)
			response.WriteJSON(w, http.StatusCreated, resp)
			return
		}

		prep := &service.CheckoutPreparation{
			OrderID:    o.ID,
			TotalCents: o.TotalCents,
		}
		gw, err := h.payments.CreatePayrexxGateway(ctx, prep, returnURL, returnURL, returnURL)
		if err != nil {
			writeError(w, http.StatusBadGateway, "gateway_error", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "twint", "redirectUrl": gw.Link, "gatewayId": gw.ID}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.GratisGuest:
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if err := h.pos.PayGratisGuest(ctx, id, deviceID); err != nil {
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "gratis_guest"}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.GratisVip:
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if err := h.pos.PayGratisVIP(ctx, id, deviceID); err != nil {
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "gratis_vip"}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.GratisStaff:
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if err := h.pos.PayGratisStaff(ctx, id, deviceID); err != nil {
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "gratis_staff"}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	case generated.Gratis100club:
		if body.Club100 == nil {
			writeError(w, http.StatusBadRequest, "missing_club100_info", "Club100 info required for gratis_100club payment")
			return
		}
		var deviceID *uuid.UUID
		if did, ok := auth.GetDeviceID(ctx); ok {
			deviceID = &did
		}
		if err := h.pos.PayGratis100Club(ctx, id, deviceID, body.Club100.ElvantoPersonId, body.Club100.ElvantoPersonName, body.Club100.FreeQuantity); err != nil {
			if errors.Is(err, service.ErrProductNotFreeForClub100) {
				writeError(w, http.StatusBadRequest, "product_not_free", "Eines oder mehrere Produkte in dieser Bestellung sind nicht als Gratis-Produkt für 100 Club konfiguriert.")
				return
			}
			writeError(w, http.StatusBadRequest, "payment_failed", err.Error())
			return
		}
		resp := map[string]any{"orderId": id.String(), "method": "gratis_100club", "elvantoPersonId": body.Club100.ElvantoPersonId}
		cachePaymentResponse(resp)
		response.WriteJSON(w, http.StatusCreated, resp)

	default:
		writeError(w, http.StatusBadRequest, "invalid_method", "Unsupported payment method")
	}
}

// ListEvents returns months with paid orders for admin dashboard navigation.
// (GET /events)
func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.orders.ListEvents(r.Context())
	if err != nil {
		writeEntError(w, err)
		return
	}

	items := make([]generated.Event, 0, len(events))
	for _, e := range events {
		items = append(items, generated.Event{
			Year:       e.Year,
			Month:      e.Month,
			OrderCount: e.OrderCount,
		})
	}
	response.WriteJSON(w, http.StatusOK, generated.EventList{Items: items})
}
