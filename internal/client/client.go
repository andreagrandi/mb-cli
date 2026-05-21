package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/andreagrandi/mb-cli/internal/config"
	"github.com/andreagrandi/mb-cli/internal/version"
)

var UserAgent = "mb-cli/" + version.Version

// HTTPDoer is an interface for executing HTTP requests, enabling test injection.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents the Metabase API client.
type Client struct {
	BaseURL      string
	HTTPClient   HTTPDoer
	APIKey       string
	SessionToken string
	Verbose      bool
	RedactPII    bool
}

// NewClient creates a new Metabase API client from the provided config.
// Request timeouts are controlled by the context passed to each request.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		BaseURL:      cfg.Host,
		HTTPClient:   &http.Client{},
		APIKey:       cfg.APIKey,
		SessionToken: cfg.SessionToken,
	}
}

// Do executes an HTTP request with authentication headers and error handling.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.SessionToken != "" {
		req.Header.Set("X-Metabase-Session", c.SessionToken)
	} else {
		req.Header.Set("x-api-key", c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	if c.Verbose {
		fmt.Fprintf(os.Stderr, "Request: %s %s\n", req.Method, req.URL)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, requestError(req.Context(), err)
	}

	if c.Verbose {
		fmt.Fprintf(os.Stderr, "Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// requestError converts a transport-level failure into a clearer message,
// distinguishing request timeouts and cancellations from generic errors.
func requestError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded), ctx.Err() == context.DeadlineExceeded:
		return fmt.Errorf("request timed out: %w", err)
	case errors.Is(err, context.Canceled), ctx.Err() == context.Canceled:
		return fmt.Errorf("request canceled: %w", err)
	default:
		return fmt.Errorf("http request failed: %w", err)
	}
}

// Get performs a GET request to the given endpoint with optional query parameters.
func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	fullURL := c.BaseURL + endpoint
	if params != nil {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.Do(req)
}

// Post performs a POST request to the given endpoint with a JSON body.
func (c *Client) Post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.Do(req)
}

// DecodeJSON decodes a JSON response body into the provided value.
func (c *Client) DecodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}
