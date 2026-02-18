package payrexx

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// TransactionStatus represents transaction statuses in Payrexx.
const (
	TransactionStatusWaiting    = "waiting"
	TransactionStatusConfirmed  = "confirmed"
	TransactionStatusAuthorized = "authorized"
	TransactionStatusReserved   = "reserved"
	TransactionStatusRefunded   = "refunded"
	TransactionStatusPartRefund = "partially-refunded"
	TransactionStatusCancelled  = "cancelled"
	TransactionStatusDeclined   = "declined"
	TransactionStatusError      = "error"
	TransactionStatusUncaptured = "uncaptured"
	TransactionStatusChargeback = "chargeback"
)

// WebhookEvent represents a parsed Payrexx webhook event.
type WebhookEvent struct {
	// Transaction details
	Transaction WebhookTransaction `json:"transaction"`
}

// WebhookTransaction contains transaction data from the webhook.
type WebhookTransaction struct {
	ID            int    `json:"id"`
	UUID          string `json:"uuid,omitempty"`
	Status        string `json:"status"`
	Amount        int    `json:"amount"` // in cents
	Currency      string `json:"currency"`
	ReferenceID   string `json:"referenceId"`
	PaymentMethod string `json:"psp"`
	GatewayID     int    `json:"invoice,omitempty"` // Gateway/Invoice ID

	// Customer info
	Contact WebhookContact `json:"contact,omitempty"`
}

// WebhookContact contains customer contact information.
type WebhookContact struct {
	Email string `json:"email,omitempty"`
}

// ParseWebhookEvent parses a Payrexx webhook body into a WebhookEvent.
// The webhook is sent as form-urlencoded data with a transaction array.
func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("payrexx: failed to parse webhook body: %w", err)
	}

	event := &WebhookEvent{}

	// Parse transaction data
	// Payrexx sends: transaction[id], transaction[status], transaction[amount], etc.
	txn := &event.Transaction

	if v := values.Get("transaction[id]"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			txn.ID = id
		}
	}
	if v := values.Get("transaction[uuid]"); v != "" {
		txn.UUID = v
	}
	if v := values.Get("transaction[status]"); v != "" {
		txn.Status = v
	}
	if v := values.Get("transaction[amount]"); v != "" {
		if amt, err := strconv.Atoi(v); err == nil {
			txn.Amount = amt
		}
	}
	if v := values.Get("transaction[currency]"); v != "" {
		txn.Currency = v
	}
	if v := values.Get("transaction[referenceId]"); v != "" {
		txn.ReferenceID = v
	}
	if v := values.Get("transaction[psp]"); v != "" {
		txn.PaymentMethod = v
	}
	if v := values.Get("transaction[invoice][paymentRequestId]"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			txn.GatewayID = id
		}
	}
	// Alternative: transaction[invoice] as direct ID
	if txn.GatewayID == 0 {
		if v := values.Get("transaction[invoice]"); v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				txn.GatewayID = id
			}
		}
	}

	// Contact info
	if v := values.Get("transaction[contact][email]"); v != "" {
		txn.Contact.Email = v
	}

	return event, nil
}

// ParseWebhookEventJSON parses a JSON webhook body (alternative format).
func ParseWebhookEventJSON(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("payrexx: failed to parse JSON webhook: %w", err)
	}
	return &event, nil
}

// IsSuccessStatus returns true if the transaction status indicates successful payment.
func (t *WebhookTransaction) IsSuccessStatus() bool {
	return t.Status == TransactionStatusConfirmed ||
		t.Status == TransactionStatusAuthorized ||
		t.Status == TransactionStatusReserved
}

// IsFailedStatus returns true if the transaction status indicates a failed payment.
func (t *WebhookTransaction) IsFailedStatus() bool {
	return t.Status == TransactionStatusDeclined ||
		t.Status == TransactionStatusError ||
		t.Status == TransactionStatusCancelled
}

// IsRefundStatus returns true if the transaction was refunded.
func (t *WebhookTransaction) IsRefundStatus() bool {
	return t.Status == TransactionStatusRefunded ||
		t.Status == TransactionStatusPartRefund ||
		t.Status == TransactionStatusChargeback
}
