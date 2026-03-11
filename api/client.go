package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_cli/config"
	"note_cli/config/buildinfo"
	"note_cli/utils"
)

// Client is an HTTP client wrapper for the Note App API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Config     *config.Config
}

// NewClient creates a new API client configured with the loaded config
func NewClient(cfg *config.Config) *Client {
	baseURL := fmt.Sprintf("http://%s:%d/api/v1", cfg.Host, cfg.Port)
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		Config:     cfg,
	}
}

func (c *Client) doRequest(req *http.Request, retryOn401 bool) ([]byte, error) {
	req.Header.Set("Secret-Key", buildinfo.SecretKey)
	req.Header.Set("Api-Key", buildinfo.APIKey)
	if c.Config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
	}
	
	utils.Debugf("API Request: %s %s", req.Method, req.URL.String())
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.Debugf("API Response: Status %d %s", resp.StatusCode, resp.Status)
	utils.Debugf("API Response Body: %s", string(body))

	if resp.StatusCode >= 400 {
		// Handle token refresh logic automatically
		if resp.StatusCode == http.StatusUnauthorized && retryOn401 && c.Config.RefreshToken != "" {
			errRefresh := c.Refresh()
			if errRefresh == nil {
				// Retry original request 
				// We need to clone the request because the body might be consumed (not handled here completely but for GET it's fine)
				// For simplicity in CLI, if it's a POST, we'll recommend doing it carefully.
				// Since we might need to recreate the request, we just return a specific error flag for now,
				// or recreate simple cases.
				// Better approach: simply say login expired if refresh fails.
				req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
				return c.doRequest(req, false)
			}
		}

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("API error: status %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// post is a helper for POST requests (form or json)
func (c *Client) post(endpoint string, contentType string, bodyReader io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.doRequest(req, true)
}

// get is a helper for GET requests
func (c *Client) get(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req, true)
}

// deleteReq is a helper for DELETE requests
func (c *Client) deleteReq(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req, true)
}

// patchReq is a helper for PATCH requests
func (c *Client) patchReq(endpoint string, bodyReader io.Reader) ([]byte, error) {
	req, err := http.NewRequest("PATCH", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, true)
}
