package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_cli/config"
	"note_cli/utils"
)

// Client Note App API 통신용 HTTP 클라이언트 래퍼
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Config     *config.Config
}

// NewClient 로드된 설정으로 새 API 클라이언트 생성
func NewClient(cfg *config.Config) *Client {
	baseURL := fmt.Sprintf("http://%s:%d/api/v1", cfg.Host, cfg.Port)
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		Config:     cfg,
	}
}

func (c *Client) doRequest(req *http.Request, retryOn401 bool) ([]byte, error) {
	resp, err := c.sendRequest(req, retryOn401, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.Debugf("API Response Body: %s", string(body))
	return body, nil
}

// post POST 요청 헬퍼 (form 또는 json)
func (c *Client) post(endpoint string, contentType string, bodyReader io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.doRequest(req, true)
}

// get GET 요청 헬퍼
func (c *Client) get(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req, true)
}

// deleteReq DELETE 요청 헬퍼
func (c *Client) deleteReq(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req, true)
}

// patchReq PATCH 요청 헬퍼
func (c *Client) patchReq(endpoint string, bodyReader io.Reader) ([]byte, error) {
	req, err := http.NewRequest("PATCH", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, true)
}

// doRequestStream HTTP 요청을 실행하고 스트리밍용 Response 객체 반환
// 토큰 주입 및 기본 인증 에러 확인 처리
// 호출자가 반드시 응답 본문을 닫아야 함
func (c *Client) doRequestStream(req *http.Request, retryOn401 bool) (*http.Response, error) {
	return c.sendRequest(req, retryOn401, true)
}

// getStream 스트림을 반환하는 GET 요청 헬퍼
func (c *Client) getStream(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequestStream(req, true)
}

// postStream 스트리밍 본문을 사용하는 POST 요청 헬퍼
func (c *Client) postStream(endpoint string, contentType string, bodyReader io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.doRequestStream(req, true)
}

func (c *Client) sendRequest(req *http.Request, retryOn401 bool, stream bool) (*http.Response, error) {
	if err := c.applyAuthHeaders(req); err != nil {
		return nil, err
	}
	c.logRequest(req, stream)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.logResponse(resp, stream)
	if resp.StatusCode < http.StatusBadRequest {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	utils.Debugf("API Error Body: %s", string(body))
	if c.shouldRetryUnauthorized(resp.StatusCode, retryOn401) {
		if err := c.retryWithRefresh(req); err == nil {
			return c.sendRequest(req, false, stream)
		}
	}

	return nil, parseAPIError(resp.StatusCode, body)
}

func (c *Client) applyAuthHeaders(req *http.Request) error {
	secrets, err := config.LoadRuntimeSecrets()
	if err != nil {
		return err
	}

	req.Header.Set("Secret-Key", secrets.SecretKey)
	req.Header.Set("Api-Key", secrets.APIKey)
	if c.Config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
	}

	return nil
}

func (c *Client) shouldRetryUnauthorized(statusCode int, retryOn401 bool) bool {
	return statusCode == http.StatusUnauthorized && retryOn401 && c.Config.RefreshToken != ""
}

func (c *Client) retryWithRefresh(req *http.Request) error {
	if err := c.Refresh(); err != nil {
		return err
	}
	if err := resetRequestBody(req); err != nil {
		return err
	}

	return nil
}

func resetRequestBody(req *http.Request) error {
	if req.GetBody == nil {
		return nil
	}

	body, err := req.GetBody()
	if err != nil {
		return fmt.Errorf("failed to reset request body: %w", err)
	}
	req.Body = body

	return nil
}

func parseAPIError(statusCode int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail != "" {
		return &apiErr
	}

	return fmt.Errorf("API error: status %d - %s", statusCode, string(body))
}

func (c *Client) logRequest(req *http.Request, stream bool) {
	if stream {
		utils.Debugf("API Request (Stream): %s %s", req.Method, req.URL.String())
		return
	}

	utils.Debugf("API Request: %s %s", req.Method, req.URL.String())
}

func (c *Client) logResponse(resp *http.Response, stream bool) {
	if stream {
		utils.Debugf("API Response (Stream): Status %d %s", resp.StatusCode, resp.Status)
		return
	}

	utils.Debugf("API Response: Status %d %s", resp.StatusCode, resp.Status)
}
