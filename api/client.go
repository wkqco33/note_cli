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
		// 자동 토큰 갱신 로직 처리
		if resp.StatusCode == http.StatusUnauthorized && retryOn401 && c.Config.RefreshToken != "" {
			errRefresh := c.Refresh()
			if errRefresh == nil {
				// 원본 요청 재시도
				// GET 요청은 본문이 없어 별도 처리 불필요, POST의 경우 주의 필요
				// 단순화를 위해 CLI 환경에서 권한 만료 시 토큰 갱신 후 재할당
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
	req.Header.Set("Secret-Key", buildinfo.SecretKey)
	req.Header.Set("Api-Key", buildinfo.APIKey)
	if c.Config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
	}
	
	utils.Debugf("API Request (Stream): %s %s", req.Method, req.URL.String())
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	utils.Debugf("API Response (Stream): Status %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode == http.StatusUnauthorized && retryOn401 && c.Config.RefreshToken != "" {
			errRefresh := c.Refresh()
			if errRefresh == nil {
				if req.GetBody != nil {
					newBody, _ := req.GetBody()
					req.Body = newBody
				}
				req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
				return c.doRequestStream(req, false)
			}
		}

		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail != "" {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("API error: status %d - %s", resp.StatusCode, string(body))
	}

	return resp, nil
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
