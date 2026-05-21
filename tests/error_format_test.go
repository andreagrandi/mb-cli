package tests

import (
	"fmt"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/cli"
)

func TestClassifyConfigError(t *testing.T) {
	tests := []struct {
		name          string
		errMsg        string
		expectedType  string
		hasSuggestion bool
	}{
		{
			name:          "missing MB_HOST",
			errMsg:        "MB_HOST is required",
			expectedType:  "CONFIG_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "missing MB_API_KEY",
			errMsg:        "MB_API_KEY is required",
			expectedType:  "CONFIG_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "auth 401",
			errMsg:        "API request failed with status 401: Unauthorized",
			expectedType:  "AUTH_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "auth 403",
			errMsg:        "API request failed with status 403: Forbidden",
			expectedType:  "AUTH_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "dashboard not found",
			errMsg:        "failed to get dashboard 298: API request failed with status 404: Not found",
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "parameterized query failure",
			errMsg:        "parameterized query failed: check parameter keys and values (API request failed with status 400: bad request)",
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "request timeout",
			errMsg:        "failed to list databases: request timed out: context deadline exceeded",
			expectedType:  "TIMEOUT_ERROR",
			hasSuggestion: true,
		},
		{
			name:         "request canceled",
			errMsg:       "request canceled: context canceled",
			expectedType: "CANCELED_ERROR",
		},
		{
			name:         "api 404",
			errMsg:       "API request failed with status 404: Not Found",
			expectedType: "API_ERROR",
		},
		{
			name:         "api 500",
			errMsg:       "API request failed with status 500: Internal Server Error",
			expectedType: "API_ERROR",
		},
		{
			name:          "no database match",
			errMsg:        "no database matching 'foo' found",
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "ambiguous database",
			errMsg:        "ambiguous database name 'prod', matches: Production (id=1), Prod-staging (id=2). Use database ID instead",
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name:         "generic error",
			errMsg:       "something went wrong",
			expectedType: "GENERAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, suggestion := cli.ClassifyError(fmt.Errorf("%s", tt.errMsg))

			if errType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, errType)
			}
			if tt.hasSuggestion && suggestion == "" {
				t.Error("expected a suggestion but got empty string")
			}
			if !tt.hasSuggestion && suggestion != "" {
				t.Errorf("expected no suggestion but got: %s", suggestion)
			}
		})
	}
}
