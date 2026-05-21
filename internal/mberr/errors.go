// Package mberr defines typed errors shared across mb-cli so that failures can
// be classified from their type rather than by matching message substrings.
package mberr

import (
	"fmt"
	"net/http"
)

// ResourceKind identifies the kind of Metabase resource a failure relates to.
type ResourceKind string

const (
	ResourceDatabase           ResourceKind = "database"
	ResourceTable              ResourceKind = "table"
	ResourceField              ResourceKind = "field"
	ResourceDashboard          ResourceKind = "dashboard"
	ResourceCard               ResourceKind = "card"
	ResourceCollection         ResourceKind = "collection"
	ResourceDashboardParameter ResourceKind = "dashboard parameter"
)

// ConfigError indicates missing or invalid configuration, such as required
// environment variables not being set.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string { return e.Message }

// APIError represents a non-2xx response from the Metabase API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API request failed with status %d: %s", e.StatusCode, e.Body)
}

// IsAuth reports whether the response indicates an authentication or
// authorization failure.
func (e *APIError) IsAuth() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// TimeoutError indicates an API request that exceeded its deadline.
type TimeoutError struct {
	Err error
}

func (e *TimeoutError) Error() string { return fmt.Sprintf("request timed out: %v", e.Err) }
func (e *TimeoutError) Unwrap() error { return e.Err }

// CanceledError indicates an API request canceled before it completed.
type CanceledError struct {
	Err error
}

func (e *CanceledError) Error() string { return fmt.Sprintf("request canceled: %v", e.Err) }
func (e *CanceledError) Unwrap() error { return e.Err }

// RequestError wraps a failed request for a specific Metabase resource,
// preserving which resource was targeted so the failure can be given
// resource-specific guidance.
type RequestError struct {
	Resource ResourceKind
	Op       string
	Err      error
}

func (e *RequestError) Error() string { return e.Op + ": " + e.Err.Error() }
func (e *RequestError) Unwrap() error { return e.Err }

// ResolutionError indicates a failure to resolve a name to a resource ID,
// because no resource matched or the name was ambiguous.
type ResolutionError struct {
	Kind    ResourceKind
	Message string
}

func (e *ResolutionError) Error() string { return e.Message }

// ParameterizedQueryError indicates a parameterized query was rejected by the
// API, typically because parameter keys or values were invalid.
type ParameterizedQueryError struct {
	Err error
}

func (e *ParameterizedQueryError) Error() string {
	return fmt.Sprintf("parameterized query failed: check parameter keys and values (%v)", e.Err)
}

func (e *ParameterizedQueryError) Unwrap() error { return e.Err }
