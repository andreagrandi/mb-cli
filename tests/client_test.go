package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
)

func newTestClient(serverURL string) *client.Client {
	cfg := &config.Config{
		Host:   serverURL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg)
}

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Host:   "https://metabase.example.com",
		APIKey: "my-key",
	}

	c := client.NewClient(cfg)

	if c.BaseURL != "https://metabase.example.com" {
		t.Errorf("expected base URL 'https://metabase.example.com', got '%s'", c.BaseURL)
	}

	if c.APIKey != "my-key" {
		t.Errorf("expected API key 'my-key', got '%s'", c.APIKey)
	}
}

func TestDo_SetsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("x-api-key")
		if apiKey != "test-api-key" {
			t.Errorf("expected x-api-key 'test-api-key', got '%s'", apiKey)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		userAgent := r.Header.Get("User-Agent")
		if !strings.HasPrefix(userAgent, "mb-cli/") {
			t.Errorf("expected User-Agent to start with 'mb-cli/', got '%s'", userAgent)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGet_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if r.URL.Query().Get("include") != "tables" {
			t.Errorf("expected query param include=tables, got '%s'", r.URL.Query().Get("include"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	params := url.Values{}
	params.Set("include", "tables")

	resp, err := c.Get("/api/database/", params)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGet_WithoutQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got '%s'", r.URL.RawQuery)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	resp, err := c.Get("/api/database/1", nil)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestPost_WithJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if payload["database"] != float64(1) {
			t.Errorf("expected database=1, got %v", payload["database"])
		}

		if payload["type"] != "native" {
			t.Errorf("expected type=native, got %v", payload["type"])
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"rows": []}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	requestBody := map[string]any{
		"database": 1,
		"type":     "native",
		"native":   map[string]any{"query": "SELECT 1"},
	}

	resp, err := c.Post("/api/dataset/", requestBody)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDo_Error4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	_, err := c.Get("/api/database/999", nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain '400', got: %v", err)
	}

	if !strings.Contains(err.Error(), "Bad request") {
		t.Errorf("expected error to contain response body, got: %v", err)
	}
}

func TestDo_Error5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	_, err := c.Get("/api/database/", nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain '500', got: %v", err)
	}
}

func TestDecodeJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name": "test-db", "id": 42}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	resp, err := c.Get("/api/database/42", nil)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}

	var result map[string]any
	err = c.DecodeJSON(resp, &result)
	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if result["name"] != "test-db" {
		t.Errorf("expected name 'test-db', got %v", result["name"])
	}

	if result["id"] != float64(42) {
		t.Errorf("expected id 42, got %v", result["id"])
	}
}

func newTestClientWithSessionToken(serverURL string) *client.Client {
	cfg := &config.Config{
		Host:         serverURL,
		SessionToken: "test-session-token",
	}
	return client.NewClient(cfg)
}

func TestNewClient_WithSessionToken(t *testing.T) {
	cfg := &config.Config{
		Host:         "https://metabase.example.com",
		SessionToken: "my-session-token",
	}

	c := client.NewClient(cfg)

	if c.SessionToken != "my-session-token" {
		t.Errorf("expected session token 'my-session-token', got '%s'", c.SessionToken)
	}

	if c.APIKey != "" {
		t.Errorf("expected empty API key, got '%s'", c.APIKey)
	}
}

func TestDo_SetsSessionTokenHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken := r.Header.Get("X-Metabase-Session")
		if sessionToken != "test-session-token" {
			t.Errorf("expected X-Metabase-Session 'test-session-token', got '%s'", sessionToken)
		}

		if apiKey := r.Header.Get("x-api-key"); apiKey != "" {
			t.Errorf("expected no x-api-key header when using session token, got '%s'", apiKey)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	c := newTestClientWithSessionToken(server.URL)

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDo_SessionTokenOmitsAPIKeyHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Metabase-Session") == "" {
			t.Error("expected X-Metabase-Session header to be set")
		}
		if r.Header.Get("x-api-key") != "" {
			t.Errorf("x-api-key must not be sent when session token is used")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClientWithSessionToken(server.URL)
	resp, err := c.Post("/api/dataset/", map[string]any{"query": "SELECT 1"})
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestVerboseMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	c.Verbose = true

	resp, err := c.Get("/api/database/", nil)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
}
