package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"note_cli/config"
	"strings"
)

// Login authenticates the user and updates the config tokens
func (c *Client) Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	body, err := c.post("/auth/login", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return err
	}

	c.Config.AccessToken = tr.AccessToken
	c.Config.RefreshToken = tr.RefreshToken
	return config.Save(c.Config)
}

// Refresh attempts to fetch a new access token using the refresh token
func (c *Client) Refresh() error {
	endpoint := "/auth/refresh?refresh_token=" + url.QueryEscape(c.Config.RefreshToken)
	
	// Ensure we don't retry on 401 within the refresh itself
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, nil)
	if err != nil {
		return err
	}
	
	body, err := c.doRequest(req, false)
	if err != nil {
		return err
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return err
	}

	c.Config.AccessToken = tr.AccessToken
	c.Config.RefreshToken = tr.RefreshToken
	return config.Save(c.Config)
}

// Register creates a new user account
func (c *Client) Register(user UserCreate) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = c.post("/users", "application/json", bytes.NewReader(payload))
	return err
}
