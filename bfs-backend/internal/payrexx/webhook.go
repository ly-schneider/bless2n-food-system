package payrexx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

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

type WebhookEvent struct {
	Transaction WebhookTransaction `json:"transaction"`
}

type WebhookTransaction struct {
	ID            int            `json:"id"`
	UUID          string         `json:"uuid,omitempty"`
	Status        string         `json:"status"`
	Amount        int            `json:"amount"`
	Currency      string         `json:"currency"`
	ReferenceID   string         `json:"referenceId"`
	PaymentMethod string         `json:"psp"`
	Invoice       WebhookInvoice `json:"invoice,omitempty"`
	Contact       WebhookContact `json:"contact,omitempty"`
}

type WebhookInvoice struct {
	PaymentRequestID *int   `json:"paymentRequestId,omitempty"`
	Number           string `json:"number,omitempty"`
	Currency         string `json:"currency,omitempty"`
	ReferenceID      string `json:"referenceId,omitempty"`
}

type WebhookContact struct {
	Email string `json:"email,omitempty"`
}

func VerifyWebhookSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("payrexx: failed to parse webhook JSON: %w", err)
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
