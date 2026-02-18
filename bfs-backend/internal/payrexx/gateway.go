package payrexx

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Gateway represents a Payrexx payment gateway (payment session).
type Gateway struct {
	ID           int       `json:"id"`
	Status       string    `json:"status"`
	Hash         string    `json:"hash"`
	Link         string    `json:"link"`
	Amount       int       `json:"amount"` // in cents
	Currency     string    `json:"currency"`
	ReferenceID  string    `json:"referenceId"`
	PaymentMeans []string  `json:"pm"`
	CreatedAt    time.Time `json:"-"`
	InvoiceItems []InvoiceItem
	SuccessURL   string
	FailedURL    string
	CancelURL    string
}

// InvoiceItem represents a line item in a Payrexx invoice.
type InvoiceItem struct {
	Name     string
	Quantity int
	Amount   int // in cents
}

// GatewayResponse wraps the Payrexx API response for Gateway operations.
type GatewayResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"`
	Data    []Gateway `json:"data"`
}

// CreateGatewayParams contains parameters for creating a new payment gateway.
type CreateGatewayParams struct {
	// Amount in cents (e.g., 1500 for CHF 15.00)
	Amount int
	// Currency code (e.g., "CHF")
	Currency string
	// ReferenceID is your order reference (stored and returned in webhooks)
	ReferenceID string
	// SuccessRedirectURL is where the customer is redirected after successful payment
	SuccessRedirectURL string
	// FailedRedirectURL is where the customer is redirected after failed payment
	FailedRedirectURL string
	// CancelRedirectURL is where the customer is redirected if they cancel
	CancelRedirectURL string
	// PaymentMeans (optional) limits which payment methods are available
	// e.g., ["twint", "visa", "mastercard"]
	PaymentMeans []string
	// InvoiceItems (optional) line items to display on payment page
	InvoiceItems []InvoiceItem
	// CustomerEmail (optional) pre-fills customer email
	CustomerEmail string
	// Purpose (optional) description shown on payment page
	Purpose string
	// Validity in minutes (optional, default is 15 minutes)
	ValidityMinutes int
	// LookAndFeelProfile (optional) custom styling profile ID
	LookAndFeelProfile string
}

// CreateGateway creates a new payment gateway (payment session).
func (c *Client) CreateGateway(params CreateGatewayParams) (*Gateway, error) {
	v := url.Values{}

	// Required fields
	v.Set("amount", strconv.Itoa(params.Amount))
	v.Set("currency", params.Currency)

	// Optional fields
	if params.ReferenceID != "" {
		v.Set("referenceId", params.ReferenceID)
	}
	if params.SuccessRedirectURL != "" {
		v.Set("successRedirectUrl", params.SuccessRedirectURL)
	}
	if params.FailedRedirectURL != "" {
		v.Set("failedRedirectUrl", params.FailedRedirectURL)
	}
	if params.CancelRedirectURL != "" {
		v.Set("cancelRedirectUrl", params.CancelRedirectURL)
	}
	if params.CustomerEmail != "" {
		v.Set("fields[email][value]", params.CustomerEmail)
	}
	if params.Purpose != "" {
		v.Set("purpose", params.Purpose)
	}
	if params.ValidityMinutes > 0 {
		v.Set("validity", strconv.Itoa(params.ValidityMinutes))
	}
	if params.LookAndFeelProfile != "" {
		v.Set("lookAndFeelProfile", params.LookAndFeelProfile)
	}

	// Payment means (e.g., twint, visa)
	for i, pm := range params.PaymentMeans {
		v.Set(fmt.Sprintf("pm[%d]", i), pm)
	}

	// Invoice items — name must be a language-keyed array per Payrexx API
	for i, item := range params.InvoiceItems {
		v.Set(fmt.Sprintf("basket[%d][name][1]", i), item.Name)
		v.Set(fmt.Sprintf("basket[%d][quantity]", i), strconv.Itoa(item.Quantity))
		v.Set(fmt.Sprintf("basket[%d][amount]", i), strconv.Itoa(item.Amount))
	}

	body, err := c.doRequest("POST", "Gateway/", v)
	if err != nil {
		return nil, err
	}

	var resp GatewayResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("payrexx: failed to parse response: %w", err)
	}

	if resp.Status != "success" || len(resp.Data) == 0 {
		return nil, fmt.Errorf("payrexx: failed to create gateway: %s", resp.Message)
	}

	return &resp.Data[0], nil
}

// GetGateway retrieves a gateway by ID.
func (c *Client) GetGateway(id int) (*Gateway, error) {
	v := url.Values{}

	body, err := c.doRequest("GET", fmt.Sprintf("Gateway/%d/", id), v)
	if err != nil {
		return nil, err
	}

	var resp GatewayResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("payrexx: failed to parse response: %w", err)
	}

	if resp.Status != "success" || len(resp.Data) == 0 {
		return nil, fmt.Errorf("payrexx: gateway not found: %s", resp.Message)
	}

	return &resp.Data[0], nil
}

// DeleteGateway deletes a gateway by ID.
func (c *Client) DeleteGateway(id int) error {
	v := url.Values{}

	_, err := c.doRequest("DELETE", fmt.Sprintf("Gateway/%d/", id), v)
	return err
}

// GatewayStatus constants
const (
	GatewayStatusWaiting   = "waiting"
	GatewayStatusConfirmed = "confirmed"
	GatewayStatusCancelled = "cancelled"
	GatewayStatusDeclined  = "declined"
	GatewayStatusError     = "error"
	GatewayStatusExpired   = "expired"
)
