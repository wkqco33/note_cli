package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"note_cli/config"
	"note_cli/utils"
	"strings"
	"time"
)

// httpTimeout 스트리밍 전송을 위해 응답 시작 대기 시간만 제한하고 본문 전송은 무제한으로 둔다.
var httpTransport = &http.Transport{
	ResponseHeaderTimeout: 30 * time.Second,
	IdleConnTimeout:       90 * time.Second,
}

// requestTimeout 일반(비스트림) 요청의 전체 제한 시간.
const requestTimeout = 60 * time.Second

// Client Note App API 통신용 HTTP 클라이언트 래퍼
type Client struct {
	BaseURL string
	// HTTPClient 일반 JSON 요청용 (전체 타임아웃 적용)
	HTTPClient *http.Client
	// StreamClient 대용량 업로드/다운로드용 (본문 전송 시간 무제한)
	StreamClient *http.Client
	Config       *config.Config

	// autoLoggingIn 자동 로그인 진행 중 여부 (재귀 자동 로그인 방지)
	autoLoggingIn bool
}

// NewClient 로드된 설정으로 새 API 클라이언트 생성
func NewClient(cfg *config.Config) *Client {
	baseURL := fmt.Sprintf("http://%s:%d/api/v1", cfg.Host, cfg.Port)
	return &Client{
		BaseURL:      baseURL,
		HTTPClient:   &http.Client{Transport: httpTransport, Timeout: requestTimeout},
		StreamClient: &http.Client{Transport: httpTransport},
		Config:       cfg,
	}
}

// Close 원격 클라이언트는 별도 자원 정리가 필요 없으므로 호환을 위해 제공한다.
func (c *Client) Close() error {
	return nil
}

func (c *Client) doRequest(req *http.Request, retryOn401 bool) ([]byte, error) {
	resp, err := c.sendRequest(req, retryOn401, false)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 본문을 읽지 못했습니다: %w", err)
	}

	utils.Debugf("API Response Body: %d bytes", len(body))
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

// doRequestStream 스트리밍용 Response를 반환한다. 호출자가 응답 본문을 닫아야 한다.
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

// postStream 스트리밍 본문 POST 요청. 복원 불가능한 본문이므로 401 재시도를 끈다.
func (c *Client) postStream(endpoint string, contentType string, bodyReader io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.doRequestStream(req, false)
}

// ensureValidToken 스트리밍 업로드 전에 토큰을 사전 검증한다.
func (c *Client) ensureValidToken() error {
	// 유효한 JWT면 로컬에서 만료 여부만 판단해 서버 호출을 줄인다.
	if tokenStillValid(c.Config.AccessToken) {
		return nil
	}
	if _, err := c.get("/boards/me"); err != nil {
		return fmt.Errorf("인증 토큰 확인에 실패했습니다: %w", err)
	}
	return nil
}

func (c *Client) sendRequest(req *http.Request, retryOn401 bool, stream bool) (*http.Response, error) {
	if err := c.applyAuthHeaders(req); err != nil {
		return nil, err
	}
	c.logRequest(req, stream)

	httpClient := c.HTTPClient
	if stream {
		httpClient = c.StreamClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 실패: %w", err)
	}

	c.logResponse(resp, stream)
	if resp.StatusCode < http.StatusBadRequest {
		return resp, nil
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("응답 본문을 읽지 못했습니다: %w", err)
	}

	utils.Debugf("API Error Body: %d bytes", len(body))
	if resp.StatusCode == http.StatusUnauthorized && retryOn401 {
		if c.Config.RefreshToken != "" {
			if err := c.retryWithRefresh(req); err == nil {
				return c.sendRequest(req, false, stream)
			}
		}
		// 토큰 갱신이 실패했거나 불가능하면 저장된 계정으로 자동 로그인 시도
		if err := c.tryAutoLogin(req); err == nil {
			return c.sendRequest(req, false, stream)
		}
	}

	return nil, parseAPIError(resp.StatusCode, body)
}

func (c *Client) applyAuthHeaders(req *http.Request) error {
	// 인증은 로그인/리프레시로 발급된 JWT만 사용한다.
	if c.Config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
	}

	return nil
}

// errAutoLoginUnavailable 자동 로그인이 꺼져 있거나 계정 정보가 없어 시도할 수 없음
var errAutoLoginUnavailable = errors.New("자동 로그인을 사용할 수 없습니다")

// tryAutoLogin 자동 로그인이 켜져 있으면 저장된 계정으로 재로그인 후 요청 본문 복원
func (c *Client) tryAutoLogin(req *http.Request) error {
	if !c.Config.AutoLogin || c.Config.Username == "" || c.Config.Password == "" {
		return errAutoLoginUnavailable
	}
	// 자동 로그인 재귀 방지
	if c.autoLoggingIn {
		return errAutoLoginUnavailable
	}
	c.autoLoggingIn = true
	defer func() { c.autoLoggingIn = false }()

	utils.Debugln("인증이 만료되어 저장된 계정으로 자동 로그인을 시도합니다.")
	if err := c.Login(c.Config.Username, c.Config.Password); err != nil {
		utils.Debugf("자동 로그인 실패: %v", err)
		return err
	}

	return resetRequestBody(req)
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
		return fmt.Errorf("요청 본문을 재구성하지 못했습니다: %w", err)
	}
	req.Body = body

	return nil
}

func parseAPIError(statusCode int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail != "" {
		return &apiErr
	}

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		if msg := friendlyAuthMessage(statusCode); msg != "" {
			return &APIError{Detail: msg}
		}
		return &APIError{Detail: envelope.Error.Message}
	}

	if msg := friendlyAuthMessage(statusCode); msg != "" {
		return &APIError{Detail: msg}
	}

	return fmt.Errorf("API 오류: status %d - %s", statusCode, string(body))
}

// friendlyAuthMessage 인증 오류 상태 코드를 사용자 안내 문구로 변환한다.
func friendlyAuthMessage(statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return "로그인이 만료되었거나 인증에 실패했습니다. 'login' 명령으로 다시 로그인하세요"
	case http.StatusForbidden:
		return "이 작업을 수행할 권한이 없습니다"
	}

	return ""
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

// tokenStillValid JWT의 exp 클레임으로 유효 여부를 판단한다. 파싱 실패 시 false.
func tokenStillValid(token string) bool {
	payload := decodeJWTClaims(token)
	if payload == "" {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal([]byte(payload), &claims); err != nil || claims.Exp == 0 {
		return false
	}
	return time.Now().Unix() < claims.Exp
}

// decodeJWTClaims JWT 페이로드를 Base64 디코딩한다. 형식이 아니면 빈 문자열.
func decodeJWTClaims(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	return string(raw)
}
