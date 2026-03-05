package payrexx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBaseURL = "https://api.payrexx.com/v1.0/"

type Client struct {
	instanceName string
	apiSecret    string
	httpClient   *http.Client
}

func NewClient(instanceName, apiSecret string) *Client {
	return &Client{
		instanceName: instanceName,
		apiSecret:    apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, params map[string]any) ([]byte, error) {
	reqURL := apiBaseURL + endpoint + "?instance=" + url.QueryEscape(c.instanceName)

	var req *http.Request
	var err error

	if method == http.MethodGet || method == http.MethodDelete {
		if len(params) > 0 {
			v := url.Values{}
			for k, val := range params {
				v.Set(k, fmt.Sprintf("%v", val))
			}
			reqURL += "&" + v.Encode()
		}
		req, err = http.NewRequest(method, reqURL, nil)
	} else {
		body, jsonErr := json.Marshal(params)
		if jsonErr != nil {
			return nil, fmt.Errorf("payrexx: failed to encode request: %w", jsonErr)
		}
		req, err = http.NewRequest(method, reqURL, bytes.NewReader(body))
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("payrexx: failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payrexx: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("payrexx: failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("payrexx: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
