package api

import (
	"io"
	"net/http"

	"backend/internal/payrexx"
	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

func (h *Handlers) HandlePayrexxWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Failed to read webhook payload")
		return
	}

	event, err := payrexx.ParseWebhookEvent(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid webhook payload")
		return
	}

	ctx := r.Context()
	txn := &event.Transaction

	if !txn.IsSuccessStatus() {
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	if txn.ReferenceID == "" {
		h.logger.Warn("payrexx webhook: missing referenceId")
		writeError(w, http.StatusBadRequest, "invalid_payload", "Missing referenceId")
		return
	}

	orderID, err := uuid.Parse(txn.ReferenceID)
	if err != nil {
		h.logger.Warn("payrexx webhook: invalid referenceId", zap.String("referenceId", txn.ReferenceID))
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid referenceId")
		return
	}

	gatewayID := 0
	if txn.Invoice.PaymentRequestID != nil {
		gatewayID = *txn.Invoice.PaymentRequestID
	}

	var contactEmail *string
	if txn.Contact.Email != "" {
		contactEmail = &txn.Contact.Email
	}

	if err := h.payments.MarkOrderPaidByPayrexx(ctx, orderID, gatewayID, txn.ID, contactEmail); err != nil {
		h.logger.Error("payrexx webhook: failed to mark order paid",
			zap.Error(err),
			zap.String("orderId", orderID.String()),
		)
		writeError(w, http.StatusInternalServerError, "processing_error", "Failed to process payment")
		return
	}

	h.logger.Info("payrexx webhook: order marked as paid",
		zap.String("orderId", orderID.String()),
		zap.Int("gatewayId", gatewayID),
		zap.Int("transactionId", txn.ID),
	)

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetPayment returns a payment by ID.
// (GET /payments/{paymentId})
func (h *Handlers) GetPayment(w http.ResponseWriter, r *http.Request, paymentId openapi_types.UUID) {
	// Payment retrieval is handled by looking up the order payment.
	// The paymentId here corresponds to an OrderPayment UUID.
	// For now, delegate to the order service to fetch the order that contains this payment.
	ctx := r.Context()

	// Payments are stored as OrderPayment entities linked to orders.
	// We need a service method to look up a payment by its own ID.
	// For the interim, return a not-implemented error until the service is updated.
	_ = ctx
	writeError(w, http.StatusNotImplemented, "not_implemented", "Payment lookup by ID will be available after service refactoring")
}

