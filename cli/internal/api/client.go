// Package api provides the HTTP client for the GatewayOps API.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://api.gatewayops.com"
	DefaultTimeout = 30 * time.Second
)

// Client is the GatewayOps API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// APIError represents an error response from the API.
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Request makes an HTTP request to the API.
func (c *Client) Request(method, path string, body interface{}, result interface{}) error {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gwo-cli/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error APIError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return &errResp.Error
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Get makes a GET request.
func (c *Client) Get(path string, result interface{}) error {
	return c.Request(http.MethodGet, path, nil, result)
}

// Post makes a POST request.
func (c *Client) Post(path string, body, result interface{}) error {
	return c.Request(http.MethodPost, path, body, result)
}

// Put makes a PUT request.
func (c *Client) Put(path string, body, result interface{}) error {
	return c.Request(http.MethodPut, path, body, result)
}

// Delete makes a DELETE request.
func (c *Client) Delete(path string, result interface{}) error {
	return c.Request(http.MethodDelete, path, nil, result)
}
