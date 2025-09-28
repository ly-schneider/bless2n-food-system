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

type CreateCheckoutRequest = service.CreateCheckoutInput

type CreateCheckoutResponse struct {
	URL       string `json:"url"`
	SessionID string `json:"sessionId"`
}

// POST /v1/payments/checkout
func (h *PaymentHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
    var req CreateCheckoutRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("invalid checkout request", zap.Error(err))
        response.WriteError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    h.logger.Info("create checkout received", zap.Int("items", len(req.Items)))
    if len(req.Items) == 0 {
        response.WriteError(w, http.StatusBadRequest, "at least one item is required")
        return
    }

    // Prepare order (validate products/prices) and persist it as pending
    var userIDPtr *string
    if claims, ok := mw.GetUserFromContext(r.Context()); ok {
        sub := claims.Subject
        userIDPtr = &sub
    }

    prepRes, err := h.payments.PrepareAndCreateOrder(r.Context(), req, userIDPtr)
    if err != nil {
        h.logger.Error("prepare/create order failed", zap.Error(err))
        response.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }

    successURL := strings.TrimRight(h.cfg.App.PublicBaseURL, "/") + "/checkout/success?order_id=" + prepRes.OrderID.Hex() + "&session_id={CHECKOUT_SESSION_ID}"
    cancelURL := strings.TrimRight(h.cfg.App.PublicBaseURL, "/") + "/checkout/cancel"

    sess, err := h.payments.CreateStripeCheckoutForOrder(r.Context(), prepRes, successURL, cancelURL)
    if err != nil {
        h.logger.Error("failed to create checkout session", zap.Error(err))
        response.WriteError(w, http.StatusInternalServerError, "failed to create checkout session")
        return
    }

	response.WriteSuccess(w, http.StatusOK, CreateCheckoutResponse{
		URL:       sess.URL,
		SessionID: sess.ID,
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
    case "checkout.session.completed":
        var cs stripe.CheckoutSession
        if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
            h.logger.Error("failed to parse checkout.session", zap.Error(err))
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        if err := h.payments.MarkOrderPaid(r.Context(), cs.ClientReferenceID); err != nil {
            h.logger.Error("failed to mark order paid", zap.Error(err))
        } else {
            h.logger.Info("checkout session completed", zap.String("session_id", cs.ID))
        }
	case "payment_intent.succeeded":
		h.logger.Info("payment intent succeeded")
	case "payment_intent.payment_failed":
		h.logger.Warn("payment intent failed")
	case "charge.refunded":
		h.logger.Info("charge refunded")
	default:
		// Unexpected event type
		h.logger.Info("unhandled event type", zap.String("type", string(event.Type)))
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
