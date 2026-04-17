package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"note_cli/config"
)

func newTestClient(serverURL string) *Client {
	client := NewClient(&config.Config{
		Host:         "127.0.0.1",
		Port:         8880,
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token",
	})
	client.BaseURL = serverURL
	return client
}

func setTestSecrets(t *testing.T) {
	t.Helper()
	t.Setenv("NOTE_CLI_API_KEY", "test-api-key")
	t.Setenv("NOTE_CLI_SECRET_KEY", "test-secret-key")
}

func TestDoRequestRetriesAfterRefresh(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	setTestSecrets(t)

	var protectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			if got := r.URL.Query().Get("refresh_token"); got != "refresh-token" {
				t.Fatalf("unexpected refresh token: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/protected":
			protectedCalls++
			if protectedCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer new-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/protected", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	body, err := client.doRequest(req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
	if client.Config.AccessToken != "new-token" {
		t.Fatalf("expected refreshed access token, got %q", client.Config.AccessToken)
	}
}

func TestDoRequestStreamRetriesAfterRefresh(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	setTestSecrets(t)

	var protectedCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/stream":
			protectedCalls++
			if protectedCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer new-token" {
				t.Fatalf("unexpected authorization header: %q", got)
			}
			io.WriteString(w, "stream-ok")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/stream", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.doRequestStream(req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != "stream-ok" {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestDoRequestReturnsAPIErrorDetail(t *testing.T) {
	setTestSecrets(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodGet, server.URL+"/missing", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = client.doRequest(req, false)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Detail != "not found" {
		t.Fatalf("unexpected API error detail: %q", apiErr.Detail)
	}
}

func TestDoRequestResetsBodyOnRetry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	setTestSecrets(t)

	var payloads []string
	var postCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"new-token","refresh_token":"next-refresh","token_type":"bearer"}`)
		case "/boards":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			payloads = append(payloads, string(body))
			postCalls++
			if postCalls == 1 {
				http.Error(w, `{"detail":"expired"}`, http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, `{"id":1}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/boards", strings.NewReader(`{"title":"memo"}`))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if _, err := client.doRequest(req, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(payloads))
	}
	if payloads[0] != `{"title":"memo"}` || payloads[1] != `{"title":"memo"}` {
		t.Fatalf("request body was not preserved across retry: %#v", payloads)
	}
}
