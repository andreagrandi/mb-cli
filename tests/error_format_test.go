package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/cli"
	"github.com/andreagrandi/mb-cli/internal/mberr"
)

func TestClassifyConfigError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedType  string
		hasSuggestion bool
	}{
		{
			name:          "missing MB_HOST",
			err:           &mberr.ConfigError{Message: "MB_HOST environment variable is required"},
			expectedType:  "CONFIG_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "missing MB_API_KEY",
			err:           &mberr.ConfigError{Message: "either MB_API_KEY or MB_SESSION_TOKEN environment variable is required"},
			expectedType:  "CONFIG_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "auth 401",
			err:           &mberr.APIError{StatusCode: 401, Body: "Unauthorized"},
			expectedType:  "AUTH_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "auth 403",
			err:           &mberr.APIError{StatusCode: 403, Body: "Forbidden"},
			expectedType:  "AUTH_ERROR",
			hasSuggestion: true,
		},
		{
			name: "dashboard not found",
			err: &mberr.RequestError{
				Resource: mberr.ResourceDashboard,
				Op:       "failed to get dashboard 298",
				Err:      &mberr.APIError{StatusCode: 404, Body: "Not found"},
			},
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name: "card not found",
			err: &mberr.RequestError{
				Resource: mberr.ResourceCard,
				Op:       "failed to get card 5",
				Err:      &mberr.APIError{StatusCode: 404, Body: "Not found"},
			},
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name: "dashboard parameter not found",
			err: &mberr.RequestError{
				Resource: mberr.ResourceDashboardParameter,
				Op:       "failed to get values for dashboard 1 parameter region",
				Err:      &mberr.APIError{StatusCode: 404, Body: "Not found"},
			},
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "parameterized query failure",
			err:           &mberr.ParameterizedQueryError{Err: &mberr.APIError{StatusCode: 400, Body: "bad request"}},
			expectedType:  "API_ERROR",
			hasSuggestion: true,
		},
		{
			name:          "request timeout",
			err:           &mberr.TimeoutError{Err: context.DeadlineExceeded},
			expectedType:  "TIMEOUT_ERROR",
			hasSuggestion: true,
		},
		{
			name:         "request canceled",
			err:          &mberr.CanceledError{Err: context.Canceled},
			expectedType: "CANCELED_ERROR",
		},
		{
			name:         "api 404",
			err:          &mberr.APIError{StatusCode: 404, Body: "Not Found"},
			expectedType: "API_ERROR",
		},
		{
			name:         "api 500",
			err:          &mberr.APIError{StatusCode: 500, Body: "Internal Server Error"},
			expectedType: "API_ERROR",
		},
		{
			name: "no database match",
			err: &mberr.ResolutionError{
				Kind:    mberr.ResourceDatabase,
				Message: "no database matching 'foo' found",
			},
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name: "ambiguous database",
			err: &mberr.ResolutionError{
				Kind:    mberr.ResourceDatabase,
				Message: "ambiguous database name 'prod', matches: Production (id=1), Prod-staging (id=2). Use database ID instead",
			},
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name: "no table match",
			err: &mberr.ResolutionError{
				Kind:    mberr.ResourceTable,
				Message: "no table matching 'bar' found",
			},
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name: "no field match",
			err: &mberr.ResolutionError{
				Kind:    mberr.ResourceField,
				Message: "no field matching 'baz' found in table",
			},
			expectedType:  "RESOLUTION_ERROR",
			hasSuggestion: true,
		},
		{
			name:         "wrapped api error",
			err:          fmt.Errorf("failed to list databases: %w", &mberr.APIError{StatusCode: 500, Body: "boom"}),
			expectedType: "API_ERROR",
		},
		{
			name:          "wrapped timeout error",
			err:           fmt.Errorf("failed to list databases: %w", &mberr.TimeoutError{Err: context.DeadlineExceeded}),
			expectedType:  "TIMEOUT_ERROR",
			hasSuggestion: true,
		},
		{
			name:         "generic error",
			err:          errors.New("something went wrong"),
			expectedType: "GENERAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errType, suggestion := cli.ClassifyError(tt.err)

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
