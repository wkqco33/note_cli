package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"note_cli/config"
	"strings"
)

// Login 사용자 인증 후 설정에 토큰 저장
func (c *Client) Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	tr, err := decodeJSON[TokenResponse](c.post("/auth/login", "application/x-www-form-urlencoded", strings.NewReader(data.Encode())))
	if err != nil {
		return err
	}

	c.Config.AccessToken = tr.AccessToken
	c.Config.RefreshToken = tr.RefreshToken
	return config.Save(c.Config)
}

// Refresh 리프레시 토큰을 이용해 새 액세스 토큰 발급 시도
func (c *Client) Refresh() error {
	payload, err := json.Marshal(map[string]string{"refresh_token": c.Config.RefreshToken})
	if err != nil {
		return err
	}

	// 토큰 갱신 자체에서 401 오류 발생 시 재시도 방지
	req, err := http.NewRequest("POST", c.BaseURL+"/auth/refresh", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	tr, err := decodeJSON[TokenResponse](c.doRequest(req, false))
	if err != nil {
		return err
	}

	c.Config.AccessToken = tr.AccessToken
	c.Config.RefreshToken = tr.RefreshToken
	return config.Save(c.Config)
}

// Register 신규 사용자 계정 생성
func (c *Client) Register(user UserCreate) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = c.post("/users", "application/json", bytes.NewReader(payload))
	return err
}
