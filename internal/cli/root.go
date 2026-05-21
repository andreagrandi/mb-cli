package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreagrandi/mb-cli/internal/mberr"
	"github.com/andreagrandi/mb-cli/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "mb-cli",
	Short:   "A read-only CLI for the Metabase API",
	Version: version.Version,
	Long: `mb-cli is a read-only command-line interface for querying Metabase databases.
It allows you to list databases, inspect schemas, run SQL queries, and explore
saved questions directly from your terminal.

Before using mb-cli, set your environment variables:
  export MB_HOST=https://your-metabase-instance.com
  export MB_API_KEY=your-api-key`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !cmd.Flags().Changed("format") && IsTTY() {
			cmd.Flags().Set("format", "table")
		}

	},
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		errorFormat, _ := rootCmd.PersistentFlags().GetString("error-format")
		if errorFormat == "json" {
			writeJSONError(err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

// requestContext derives a cancellable context for a command, applying the
// configured --timeout. The returned cancel func must always be called.
func requestContext(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, timeout)
}

func init() {
	rootCmd.PersistentFlags().StringP("format", "f", "json", "Output format: json, table")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show request details on stderr")
	rootCmd.PersistentFlags().String("error-format", "text", "Error output format: text, json")
	rootCmd.PersistentFlags().Bool("redact-pii", true, "Redact PII values in query results (disable with --redact-pii=false)")
	rootCmd.PersistentFlags().Duration("timeout", 30*time.Second, "Timeout for a command's API requests (0 disables)")
}

type jsonError struct {
	Error jsonErrorDetail `json:"error"`
}

type jsonErrorDetail struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	ExitCode   int    `json:"exit_code"`
}

// ClassifyError determines the error type and suggestion for structured error output. Exported for testing.
func ClassifyError(err error) (errorType, suggestion string) {
	return classifyError(err)
}

// classifyError inspects the error's type (not its message text) to determine
// the structured error type and an actionable suggestion.
func classifyError(err error) (errorType, suggestion string) {
	var configErr *mberr.ConfigError
	if errors.As(err, &configErr) {
		return "CONFIG_ERROR", "Set MB_HOST and MB_API_KEY environment variables"
	}

	var timeoutErr *mberr.TimeoutError
	if errors.As(err, &timeoutErr) {
		return "TIMEOUT_ERROR", "Increase --timeout or check connectivity to MB_HOST"
	}

	var canceledErr *mberr.CanceledError
	if errors.As(err, &canceledErr) {
		return "CANCELED_ERROR", ""
	}

	var paramErr *mberr.ParameterizedQueryError
	if errors.As(err, &paramErr) {
		return "API_ERROR", "Check parameter IDs with 'mb-cli dashboard get <id>' or 'mb-cli card get <id> --full'"
	}

	var resolutionErr *mberr.ResolutionError
	if errors.As(err, &resolutionErr) {
		return "RESOLUTION_ERROR", resolutionSuggestion(resolutionErr.Kind)
	}

	var apiErr *mberr.APIError
	if errors.As(err, &apiErr) {
		if apiErr.IsAuth() {
			return "AUTH_ERROR", "Check that MB_API_KEY is valid and can access the requested resource"
		}
		var reqErr *mberr.RequestError
		if apiErr.StatusCode == http.StatusNotFound && errors.As(err, &reqErr) {
			return "API_ERROR", notFoundSuggestion(reqErr.Resource)
		}
		return "API_ERROR", ""
	}

	return "GENERAL_ERROR", ""
}

// resolutionSuggestion returns advice for a name-to-ID resolution failure.
func resolutionSuggestion(kind mberr.ResourceKind) string {
	switch kind {
	case mberr.ResourceDatabase:
		return "Use a database ID instead of a name"
	case mberr.ResourceTable:
		return "Use a table ID instead of a name"
	case mberr.ResourceField:
		return "Check field names with 'mb-cli table metadata <id>'"
	default:
		return ""
	}
}

// notFoundSuggestion returns advice for a 404 on a specific resource request.
func notFoundSuggestion(kind mberr.ResourceKind) string {
	switch kind {
	case mberr.ResourceDashboard:
		return "Check that the dashboard ID exists and is visible to this API key"
	case mberr.ResourceCard:
		return "Check that the card ID exists and is visible to this API key"
	case mberr.ResourceDashboardParameter:
		return "Check that the dashboard parameter ID exists for this dashboard"
	default:
		return ""
	}
}

func writeJSONError(err error) {
	errorType, suggestion := classifyError(err)
	je := jsonError{
		Error: jsonErrorDetail{
			Type:       errorType,
			Message:    err.Error(),
			Suggestion: suggestion,
			ExitCode:   1,
		},
	}
	data, _ := json.Marshal(je)
	fmt.Fprintln(os.Stderr, string(data))
}
