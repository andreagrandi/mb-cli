package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// blockingServer responds only once the request context is done or the
// fallback elapses, so callers can exercise client-side timeouts without
// leaving handler goroutines sleeping after the test finishes.
func blockingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func TestGet_ContextCanceled(t *testing.T) {
	server := blockingServer()
	defer server.Close()

	c := newTestClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Get(ctx, "/api/database/", nil)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if !strings.Contains(err.Error(), "request canceled") {
		t.Errorf("expected 'request canceled' error, got: %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error to wrap context.Canceled, got: %v", err)
	}
}

func TestGet_ContextTimeout(t *testing.T) {
	server := blockingServer()
	defer server.Close()

	c := newTestClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/api/database/", nil)
	if err == nil {
		t.Fatal("expected error for timed-out context")
	}
	if !strings.Contains(err.Error(), "request timed out") {
		t.Errorf("expected 'request timed out' error, got: %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected error to wrap context.DeadlineExceeded, got: %v", err)
	}
}

func TestPost_ContextCanceled(t *testing.T) {
	server := blockingServer()
	defer server.Close()

	c := newTestClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Post(ctx, "/api/dataset/", map[string]any{"query": "SELECT 1"})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if !strings.Contains(err.Error(), "request canceled") {
		t.Errorf("expected 'request canceled' error, got: %v", err)
	}
}

// TestClientMethod_PropagatesContextCancellation proves the context is threaded
// from a high-level client method down to the HTTP request.
func TestClientMethod_PropagatesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be reached when the context is already canceled")
	}))
	defer server.Close()

	c := newTestClient(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.ListDatabases(ctx, false)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
	if !strings.Contains(err.Error(), "request canceled") {
		t.Errorf("expected 'request canceled' error, got: %v", err)
	}
}

// TestCLI_TimeoutFlagReportsTimeoutError exercises the --timeout flag end to
// end: the command context is propagated into the API call and the resulting
// timeout is reported as a structured TIMEOUT_ERROR.
func TestCLI_TimeoutFlagReportsTimeoutError(t *testing.T) {
	server := blockingServer()
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "database", "list", "--timeout", "100ms", "--error-format", "json")

	if err == nil {
		t.Fatalf("expected non-zero exit for timed-out command, stdout: %s", stdout)
	}
	if !strings.Contains(stderr, "TIMEOUT_ERROR") {
		t.Errorf("expected TIMEOUT_ERROR in stderr, got: %s", stderr)
	}
	if !strings.Contains(stderr, "request timed out") {
		t.Errorf("expected 'request timed out' message in stderr, got: %s", stderr)
	}
}
