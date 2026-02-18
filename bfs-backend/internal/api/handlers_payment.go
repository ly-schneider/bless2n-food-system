package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/response"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

// HandlePayrexxWebhook handles incoming Payrexx payment webhooks.
// (POST /payments/webhooks/payrexx)
func (h *Handlers) HandlePayrexxWebhook(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid webhook payload")
		return
	}

	ctx := r.Context()

	// Extract transaction data from webhook.
	transaction, _ := body["transaction"].(map[string]interface{})
	if transaction == nil {
		h.logger.Warn("payrexx webhook: missing transaction object")
		writeError(w, http.StatusBadRequest, "invalid_payload", "Missing transaction data")
		return
	}

	status, _ := transaction["status"].(string)
	if status != "confirmed" {
		// Acknowledge but ignore non-confirmed statuses.
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	// Extract reference ID (order UUID).
	referenceID, _ := transaction["referenceId"].(string)
	if referenceID == "" {
		h.logger.Warn("payrexx webhook: missing referenceId")
		writeError(w, http.StatusBadRequest, "invalid_payload", "Missing referenceId")
		return
	}

	orderID, err := uuid.Parse(referenceID)
	if err != nil {
		h.logger.Warn("payrexx webhook: invalid referenceId", zap.String("referenceId", referenceID))
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid referenceId")
		return
	}

	// Extract IDs for record keeping.
	gatewayID := extractInt(body, "id")
	transactionID := extractInt(transaction, "id")

	// Extract contact email if present.
	var contactEmail *string
	if contact, ok := transaction["contact"].(map[string]interface{}); ok {
		if email, ok := contact["email"].(string); ok && email != "" {
			contactEmail = &email
		}
	}

	if err := h.payments.MarkOrderPaidByPayrexx(ctx, orderID, gatewayID, transactionID, contactEmail); err != nil {
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
		zap.Int("transactionId", transactionID),
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

// extractInt attempts to extract an integer from a map value (handles float64 from JSON).
func extractInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}
