package payrexx

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURLTemplate = "https://api.payrexx.com/v1.0/"
)

type Client struct {
	instanceName string
	apiSecret    string
	httpClient   *http.Client
	baseURL      string
}

func NewClient(instanceName, apiSecret string) *Client {
	return &Client{
		instanceName: instanceName,
		apiSecret:    apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURLTemplate,
	}
}

// buildSignature creates the HMAC-MD5 signature required by Payrexx API.
// The signature is computed over the query string parameters (sorted alphabetically by key).
func (c *Client) buildSignature(params url.Values) string {
	// Build the query string (keys sorted alphabetically)
	queryString := params.Encode()

	// Compute HMAC-MD5
	h := hmac.New(md5.New, []byte(c.apiSecret))
	h.Write([]byte(queryString))
	signature := h.Sum(nil)

	// Return base64 encoded signature
	return base64.StdEncoding.EncodeToString(signature)
}

// doRequest performs an authenticated HTTP request to the Payrexx API.
func (c *Client) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	// Add instance to params
	if params == nil {
		params = url.Values{}
	}
	params.Set("instance", c.instanceName)

	// Build signature
	signature := c.buildSignature(params)
	params.Set("ApiSignature", signature)

	// Build request URL
	reqURL := c.baseURL + endpoint

	var req *http.Request
	var err error

	if method == http.MethodGet {
		reqURL += "?" + params.Encode()
		req, err = http.NewRequest(method, reqURL, nil)
	} else {
		req, err = http.NewRequest(method, reqURL, strings.NewReader(params.Encode()))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("payrexx: failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payrexx: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payrexx: failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payrexx: API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// VerifyWebhookSignature verifies the HMAC signature of a webhook payload.
// Payrexx webhooks include the signature in the POST body as ApiSignature parameter.
func VerifyWebhookSignature(body []byte, secret string) bool {
	// Parse the body as form data
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return false
	}

	// Extract and remove the signature
	signature := values.Get("ApiSignature")
	if signature == "" {
		return false
	}
	values.Del("ApiSignature")

	// Recompute signature
	queryString := values.Encode()
	h := hmac.New(md5.New, []byte(secret))
	h.Write([]byte(queryString))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// Hash computes MD5 hash (used for some Payrexx fields).
func Hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
