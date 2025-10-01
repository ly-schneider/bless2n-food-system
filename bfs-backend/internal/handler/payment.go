package handler

import (
    "backend/internal/config"
    "backend/internal/response"
    "backend/internal/service"
    mw "backend/internal/middleware"
    "encoding/json"
    "io"
    "net/http"
    "strings"

    stripe "github.com/stripe/stripe-go/v82"
    "github.com/stripe/stripe-go/v82/webhook"
    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"
)

type PaymentHandler struct {
	payments service.PaymentService
	cfg      config.Config
	logger   *zap.Logger
}

func NewPaymentHandler(payments service.PaymentService, cfg config.Config, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{payments: payments, cfg: cfg, logger: logger}
}

// ---- Payment Intents (Payment Element) ----

type CreateIntentRequest struct {
    service.CreateCheckoutInput
    AttemptID *string `json:"attemptId,omitempty"`
}
type CreateIntentResponse struct {
    ClientSecret    string `json:"clientSecret"`
    PaymentIntentID string `json:"paymentIntentId"`
    OrderID         string `json:"orderId"`
}

// POST /v1/payments/create-intent
func (h *PaymentHandler) CreateIntent(w http.ResponseWriter, r *http.Request) {
    var req CreateIntentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("invalid create-intent request", zap.Error(err))
        response.WriteError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    if len(req.Items) == 0 {
        response.WriteError(w, http.StatusBadRequest, "at least one item is required")
        return
    }
    // Resolve auth user (optional)
    var userIDPtr *string
    if claims, ok := mw.GetUserFromContext(r.Context()); ok {
        sub := claims.Subject
        userIDPtr = &sub
    }
    // If attemptId provided, try reuse existing pending order/PI
    if req.AttemptID != nil && *req.AttemptID != "" {
        if existing, err := h.payments.FindPendingOrderByAttemptID(r.Context(), *req.AttemptID); err == nil && existing != nil {
            // If PI exists, return it
            if existing.StripePaymentIntentID != nil && *existing.StripePaymentIntentID != "" {
                pi, gerr := h.payments.GetPaymentIntent(r.Context(), *existing.StripePaymentIntentID)
                if gerr == nil && pi != nil {
                    response.WriteSuccess(w, http.StatusOK, CreateIntentResponse{
                        ClientSecret:    pi.ClientSecret,
                        PaymentIntentID: pi.ID,
                        OrderID:         existing.ID.Hex(),
                    })
                    return
                }
            }
            // Fallback: create PI for existing order
            pi, gerr := h.payments.CreatePaymentIntentForExistingPendingOrder(r.Context(), existing, userIDPtr, req.CustomerEmail)
            if gerr == nil && pi != nil {
                response.WriteSuccess(w, http.StatusOK, CreateIntentResponse{
                    ClientSecret:    pi.ClientSecret,
                    PaymentIntentID: pi.ID,
                    OrderID:         existing.ID.Hex(),
                })
                return
            }
            // If still failing, continue to create a brand new order below (last resort)
        }
    }

    // Prepare a new order and PI (attach attemptId atomically to avoid duplicates)
    prep, err := h.payments.PrepareAndCreateOrder(r.Context(), req.CreateCheckoutInput, userIDPtr, req.AttemptID)
    if err != nil {
        // Graceful recovery on unique attemptId conflict: fetch existing pending order and continue
        if req.AttemptID != nil && *req.AttemptID != "" && strings.Contains(err.Error(), "E11000 duplicate key error") {
            if existing, ferr := h.payments.FindPendingOrderByAttemptID(r.Context(), *req.AttemptID); ferr == nil && existing != nil {
                if existing.StripePaymentIntentID != nil && *existing.StripePaymentIntentID != "" {
                    if pi, gerr := h.payments.GetPaymentIntent(r.Context(), *existing.StripePaymentIntentID); gerr == nil && pi != nil {
                        response.WriteSuccess(w, http.StatusOK, CreateIntentResponse{ ClientSecret: pi.ClientSecret, PaymentIntentID: pi.ID, OrderID: existing.ID.Hex() })
                        return
                    }
                }
                if pi, cerr := h.payments.CreatePaymentIntentForExistingPendingOrder(r.Context(), existing, userIDPtr, req.CustomerEmail); cerr == nil && pi != nil {
                    response.WriteSuccess(w, http.StatusOK, CreateIntentResponse{ ClientSecret: pi.ClientSecret, PaymentIntentID: pi.ID, OrderID: existing.ID.Hex() })
                    return
                }
            }
        }
        h.logger.Error("prepare/create order failed", zap.Error(err))
        response.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }
    // Use provided email if any; backend will default to account email if logged-in and not provided
    pi, err := h.payments.CreatePaymentIntentForOrder(r.Context(), prep, req.CustomerEmail)
    if err != nil {
        h.logger.Error("failed to create payment intent", zap.Error(err))
        response.WriteError(w, http.StatusInternalServerError, "failed to create payment intent")
        return
    }
    // Best-effort cleanup of any other pending orders for this attempt
    if req.AttemptID != nil && *req.AttemptID != "" {
        if deleted, derr := h.payments.CleanupOtherPendingOrdersByAttemptID(r.Context(), *req.AttemptID, prep.OrderID); derr == nil && deleted > 0 {
            h.logger.Info("cleaned up duplicate pending orders for attempt", zap.String("attempt_id", *req.AttemptID), zap.Int64("deleted", deleted))
        }
    }
    response.WriteSuccess(w, http.StatusOK, CreateIntentResponse{
        ClientSecret:    pi.ClientSecret,
        PaymentIntentID: pi.ID,
        OrderID:         prep.OrderID.Hex(),
    })
}

type AttachEmailRequest struct {
    PaymentIntentID string  `json:"paymentIntentId"`
    Email           *string `json:"email"`
}

// PATCH /v1/payments/attach-email
func (h *PaymentHandler) AttachEmail(w http.ResponseWriter, r *http.Request) {
    var req AttachEmailRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.WriteError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    if req.PaymentIntentID == "" {
        response.WriteError(w, http.StatusBadRequest, "paymentIntentId is required")
        return
    }
    pi, err := h.payments.UpdatePaymentIntentReceiptEmail(r.Context(), req.PaymentIntentID, req.Email)
    if err != nil {
        h.logger.Error("attach email failed", zap.Error(err))
        response.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }
    response.WriteSuccess(w, http.StatusOK, map[string]any{
        "paymentIntentId": pi.ID,
        "receiptEmail":    pi.ReceiptEmail,
        "status":          pi.Status,
    })
}

// GET /v1/payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        response.WriteError(w, http.StatusBadRequest, "missing id")
        return
    }
    pi, err := h.payments.GetPaymentIntent(r.Context(), id)
    if err != nil {
        h.logger.Error("get payment failed", zap.Error(err))
        response.WriteError(w, http.StatusNotFound, "not found")
        return
    }
    // Sanitize response for frontend
    var chargeID *string
    if pi.LatestCharge != nil && pi.LatestCharge.ID != "" {
        cid := pi.LatestCharge.ID
        chargeID = &cid
    }
    // Expose only the Customer ID (if present)
    var customerID *string
    if pi.Customer != nil && pi.Customer.ID != "" {
        cid := pi.Customer.ID
        customerID = &cid
    }
    response.WriteSuccess(w, http.StatusOK, map[string]any{
        "id":           pi.ID,
        "status":       pi.Status,
        "amount":       pi.Amount,
        "currency":     pi.Currency,
        "chargeId":     chargeID,
        "customer":     customerID,
        "receiptEmail": pi.ReceiptEmail,
        "metadata":     pi.Metadata,
    })
}

// POST /v1/payments/webhook
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("reading webhook body failed", zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

    sig := r.Header.Get("Stripe-Signature")
    event, err := webhook.ConstructEvent(payload, sig, h.cfg.Stripe.WebhookSecret)
    if err != nil {
        h.logger.Error("webhook signature verification failed", zap.Error(err), zap.Int("payload_len", len(payload)), zap.Bool("sig_present", sig != ""))
        w.WriteHeader(http.StatusBadRequest)
        return
    }

	// Handle the event
    switch event.Type {
    case "payment_intent.succeeded":
        var pi stripe.PaymentIntent
        if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
            h.logger.Error("parse payment_intent.succeeded failed", zap.Error(err))
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        // Correlate by metadata.order_id
        if oid, ok := pi.Metadata["order_id"]; ok && oid != "" {
            var chargeID *string
            if pi.LatestCharge != nil && pi.LatestCharge.ID != "" {
                cid := pi.LatestCharge.ID
                chargeID = &cid
            }
            var custID *string
            if pi.Customer != nil && pi.Customer.ID != "" {
                c := pi.Customer.ID
                custID = &c
            }
            // Best-effort: contact email from PI
            var emailPtr *string
            if pi.ReceiptEmail != "" {
                re := pi.ReceiptEmail
                emailPtr = &re
            }
            // Update order paid and persist Stripe references
            if err := h.payments.MarkOrderPaid(r.Context(), oid, emailPtr); err != nil {
                h.logger.Error("mark order paid failed", zap.Error(err), zap.String("order_id", oid))
            }
            // Persist Stripe IDs for admin/ops visibility
            if oerr := h.payments.PersistPaymentSuccessByOrderID(r.Context(), oid, pi.ID, chargeID, custID, emailPtr); oerr != nil {
                h.logger.Warn("persist payment refs failed", zap.Error(oerr), zap.String("order_id", oid))
            }
            h.logger.Info("payment intent succeeded", zap.String("pi", pi.ID), zap.String("order_id", oid))
        } else {
            h.logger.Warn("payment intent succeeded missing order_id metadata", zap.String("pi", pi.ID))
        }
    case "payment_intent.payment_failed":
        var pi stripe.PaymentIntent
        if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
            if oid, ok := pi.Metadata["order_id"]; ok && oid != "" {
                if err := h.payments.CleanupPendingOrderByID(r.Context(), oid); err != nil {
                    h.logger.Error("cleanup pending order (PI) failed", zap.Error(err), zap.String("order_id", oid), zap.String("pi", pi.ID))
                } else {
                    h.logger.Warn("payment intent failed; pending order removed", zap.String("order_id", oid), zap.String("pi", pi.ID))
                }
            } else {
                h.logger.Warn("payment intent failed without order_id metadata", zap.String("pi", pi.ID))
            }
        } else {
            h.logger.Warn("payment intent failed (parse error)")
        }
    default:
        // Ignore other events
        h.logger.Info("unhandled event type", zap.String("type", string(event.Type)))
    }

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// (legacy helper removed)
