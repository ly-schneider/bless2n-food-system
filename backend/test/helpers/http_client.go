package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient provides helper methods for making HTTP requests in tests
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		headers: make(map[string]string),
	}
}

// SetHeader sets a default header for all requests
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetAuthToken sets the Authorization header with Bearer token
func (c *HTTPClient) SetAuthToken(token string) {
	c.SetHeader("Authorization", "Bearer "+token)
}

// ClearAuth removes the Authorization header
func (c *HTTPClient) ClearAuth() {
	delete(c.headers, "Authorization")
}

// POST makes a POST request with JSON payload
func (c *HTTPClient) POST(ctx context.Context, endpoint string, payload interface{}) (*HTTPResponse, error) {
	return c.makeRequest(ctx, "POST", endpoint, payload)
}

// GET makes a GET request
func (c *HTTPClient) GET(ctx context.Context, endpoint string) (*HTTPResponse, error) {
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

// PUT makes a PUT request with JSON payload
func (c *HTTPClient) PUT(ctx context.Context, endpoint string, payload interface{}) (*HTTPResponse, error) {
	return c.makeRequest(ctx, "PUT", endpoint, payload)
}

// DELETE makes a DELETE request
func (c *HTTPClient) DELETE(ctx context.Context, endpoint string) (*HTTPResponse, error) {
	return c.makeRequest(ctx, "DELETE", endpoint, nil)
}

// makeRequest is the internal method for making HTTP requests
func (c *HTTPClient) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) (*HTTPResponse, error) {
	var body io.Reader
	
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for requests with body
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add default headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
		RawResponse: resp,
	}, nil
}

// HTTPResponse represents an HTTP response for testing
type HTTPResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	RawResponse *http.Response
}

// JSON unmarshals the response body into the provided interface
func (r *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// IsSuccess checks if the response has a 2xx status code
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError checks if the response has a 4xx status code  
func (r *HTTPResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError checks if the response has a 5xx status code
func (r *HTTPResponse) IsServerError() bool {
	return r.StatusCode >= 500
}