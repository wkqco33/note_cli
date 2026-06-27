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

// Refresh 리프레시 토큰을 이용해 새 액세스 토큰 발급 시도
func (c *Client) Refresh() error {
	// NOTE: API 스펙(CLIENT_API_GUIDE.md 2.2)이 refresh_token을 query string으로
	// 받도록 정의되어 있어 따름. 일반적으로 토큰은 URL에 노출되면 로그/프록시에
	// 유출될 수 있으므로 바람직하지 않으나, 서버 계약 변경 전까지는 유지.
	endpoint := "/auth/refresh?refresh_token=" + url.QueryEscape(c.Config.RefreshToken)

	// 토큰 갱신 자체에서 401 오류 발생 시 재시도 방지
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

// Register 신규 사용자 계정 생성
func (c *Client) Register(user UserCreate) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_, err = c.post("/users", "application/json", bytes.NewReader(payload))
	return err
}
