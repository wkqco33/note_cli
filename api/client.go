package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"note_cli/config"
	"note_cli/utils"
	"strings"
	"sync"
	"time"
)

// httpTimeout 대용량 스트리밍 업로드/다운로드(최대 500MB)가 정상 동작하도록
// 헤더/응답 시작까지의 대기 시간만 제한하고 본문 전송에는 제약을 두지 않는
// 커스텀 RoundTripper. ResponseHeaderTimeout은 서버가 응답을 시작하기까지
// 대기하는 최대 시간이며 본문 스트리밍에는 영향을 주지 않는다.
var httpTransport = &http.Transport{
	ResponseHeaderTimeout: 30 * time.Second,
	IdleConnTimeout:       90 * time.Second,
}

// requestTimeout 일반(비스트림) 요청의 전체 제한 시간.
// 응답 본문 수신이 멈춰도 CLI가 무한 대기하지 않도록 한다.
const requestTimeout = 60 * time.Second

// Client Note App API 통신용 HTTP 클라이언트 래퍼
type Client struct {
	BaseURL string
	// HTTPClient 일반 JSON 요청용 (전체 타임아웃 적용)
	HTTPClient *http.Client
	// StreamClient 대용량 업로드/다운로드용 (본문 전송 시간 무제한)
	StreamClient *http.Client
	Config       *config.Config

	secretsOnce sync.Once
	secrets     config.RuntimeSecrets
	secretsErr  error

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

	httpClient := c.HTTPClient
	if stream {
		httpClient = c.StreamClient
	}
	resp, err := httpClient.Do(req)
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
	// 시크릿은 프로세스 수명 동안 불변하므로 최초 1회만 로드하여 캐싱.
	// 매 요청마다 .env 디스크 읽기를 반복하지 않도록 sync.Once로 보호.
	c.secretsOnce.Do(func() {
		c.secrets, c.secretsErr = config.LoadRuntimeSecrets()
	})
	if c.secretsErr != nil {
		return c.secretsErr
	}

	req.Header.Set("Secret-Key", c.secrets.SecretKey)
	req.Header.Set("Api-Key", c.secrets.APIKey)
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
	// 자동 로그인 요청 자체가 401을 받아 다시 자동 로그인을 시도하는 재귀 방지
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

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		if msg := friendlyAuthMessage(statusCode, envelope.Error.Message); msg != "" {
			return &APIError{Detail: msg}
		}
		return &APIError{Detail: envelope.Error.Message}
	}

	if msg := friendlyAuthMessage(statusCode, string(body)); msg != "" {
		return &APIError{Detail: msg}
	}

	return fmt.Errorf("API error: status %d - %s", statusCode, string(body))
}

// friendlyAuthMessage 인증 관련 HTTP 상태 코드를 사용자 안내 문구로 변환.
// 인증 오류가 아니면 빈 문자열을 반환해 서버 메시지를 그대로 노출한다.
func friendlyAuthMessage(statusCode int, serverMessage string) string {
	switch statusCode {
	case http.StatusUnauthorized:
		if isAPIKeyError(serverMessage) {
			return "API 인증 키가 유효하지 않거나 만료되었습니다. NOTE_CLI_API_KEY / NOTE_CLI_SECRET_KEY 설정을 확인하거나 최신 버전으로 업데이트하세요"
		}
		return "로그인이 만료되었거나 인증에 실패했습니다. 'login' 명령으로 다시 로그인하세요"
	case http.StatusForbidden:
		return "이 작업을 수행할 권한이 없습니다"
	}

	return ""
}

// isAPIKeyError Secret-Key/Api-Key 헤더 문제(키 누락·만료)인지 서버 메시지로 판별
func isAPIKeyError(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "secret-key") || strings.Contains(lower, "secret key") ||
		strings.Contains(lower, "api-key") || strings.Contains(lower, "api key")
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
