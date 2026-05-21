package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/cli"
	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/mberr"
)

// firstJSONLine returns the first line that looks like a JSON object, skipping
// any surrounding noise such as the "exit status 1" line `go run` emits.
func firstJSONLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			return line
		}
	}
	return ""
}

func TestCardParametersMethod(t *testing.T) {
	card := &client.Card{
		ID: 1,
		DatasetQuery: &client.DatasetQuery{
			Type: "native",
			Native: &client.NativeQuery{
				TemplateTags: map[string]client.TemplateTag{
					"region":         {ID: "t2", Name: "region", Type: "text"},
					"timeframe_days": {ID: "t1", Name: "timeframe_days", DisplayName: "Timeframe Days", Type: "number", Required: true, Default: float64(30)},
				},
			},
		},
	}

	params := card.CardParameters()
	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}
	if params[0].Name != "region" {
		t.Errorf("expected parameters sorted by name with 'region' first, got %s", params[0].Name)
	}
	if params[1].Name != "timeframe_days" || !params[1].Required || params[1].Default != float64(30) {
		t.Errorf("unexpected timeframe_days parameter: %+v", params[1])
	}

	empty := (&client.Card{}).CardParameters()
	if empty == nil || len(empty) != 0 {
		t.Errorf("expected empty non-nil slice for a card with no query, got %+v", empty)
	}
}

func TestCardParamsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/card/1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":          1,
			"name":        "Retention Card",
			"database_id": 1,
			"display":     "table",
			"query_type":  "native",
			"archived":    false,
			"dataset_query": map[string]any{
				"database": 1,
				"type":     "native",
				"native": map[string]any{
					"query": "select * from retention where timeframe_days = {{timeframe_days}} and region = {{region}}",
					"template-tags": map[string]any{
						"timeframe_days": map[string]any{"id": "t1", "name": "timeframe_days", "display-name": "Timeframe Days", "type": "number", "required": true, "default": 30},
						"region":         map[string]any{"id": "t2", "name": "region", "display-name": "Region", "type": "text"},
					},
				},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "card", "params", "1", "-f", "json")
	if err != nil {
		t.Fatalf("card params failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		CardID     int `json:"card_id"`
		Parameters []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode card params output: %v\noutput: %s", err, stdout)
	}

	if result.CardID != 1 {
		t.Fatalf("expected card_id 1, got %d", result.CardID)
	}
	if len(result.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(result.Parameters))
	}
	if result.Parameters[0].Name != "region" {
		t.Errorf("expected first parameter 'region', got %s", result.Parameters[0].Name)
	}
	if result.Parameters[1].Name != "timeframe_days" || !result.Parameters[1].Required {
		t.Errorf("expected required 'timeframe_days' parameter, got %+v", result.Parameters[1])
	}
}

func TestCardParamsNoParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 2, "name": "Plain Card", "database_id": 1, "display": "table", "query_type": "query", "archived": false,
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "card", "params", "2", "-f", "json")
	if err != nil {
		t.Fatalf("card params failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		Parameters []any `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode card params output: %v\noutput: %s", err, stdout)
	}
	if len(result.Parameters) != 0 {
		t.Fatalf("expected no parameters for an MBQL card, got %v", result.Parameters)
	}
}

func TestCardParamsTableHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":          1,
			"name":        "Retention Card",
			"database_id": 1,
			"display":     "table",
			"query_type":  "native",
			"archived":    false,
			"dataset_query": map[string]any{
				"database": 1,
				"type":     "native",
				"native": map[string]any{
					"query": "select * from retention where timeframe_days = {{timeframe_days}}",
					"template-tags": map[string]any{
						"timeframe_days": map[string]any{"id": "t1", "name": "timeframe_days", "type": "number"},
					},
				},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "card", "params", "1", "-f", "table")
	if err != nil {
		t.Fatalf("card params failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Run with:") {
		t.Fatalf("expected a run hint in table output, got %s", stdout)
	}
	if !strings.Contains(stdout, "mb-cli card run 1 --param timeframe_days=<value>") {
		t.Fatalf("expected a ready-to-edit card run command, got %s", stdout)
	}
}

func TestDashboardParamsListCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":       1,
			"name":     "Merchant Retention",
			"archived": false,
			"parameters": []map[string]any{
				{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="},
			},
			"dashcards": []map[string]any{
				{
					"id":      100,
					"card_id": 50,
					"parameter_mappings": []map[string]any{
						{"parameter_id": "param-merchant", "card_id": 50, "target": []any{"variable", []any{"template-tag", "merchant_name"}}},
					},
					"card": map[string]any{"id": 50, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "native", "archived": false},
				},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "dashboard", "params", "list", "1", "-f", "json")
	if err != nil {
		t.Fatalf("dashboard params list failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		DashboardID int `json:"dashboard_id"`
		Parameters  []struct {
			ID          string `json:"id"`
			Slug        string `json:"slug"`
			MappedCards []struct {
				DashcardID int `json:"dashcard_id"`
			} `json:"mapped_cards"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode dashboard params list output: %v\noutput: %s", err, stdout)
	}

	if result.DashboardID != 1 {
		t.Fatalf("expected dashboard_id 1, got %d", result.DashboardID)
	}
	if len(result.Parameters) != 1 || result.Parameters[0].ID != "param-merchant" {
		t.Fatalf("expected one parameter 'param-merchant', got %+v", result.Parameters)
	}
	if len(result.Parameters[0].MappedCards) != 1 || result.Parameters[0].MappedCards[0].DashcardID != 100 {
		t.Fatalf("expected parameter mapped to dashcard 100, got %+v", result.Parameters[0].MappedCards)
	}
}

func TestCardRunInvalidParameterSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "Retention Card", "database_id": 1, "display": "table", "query_type": "native", "archived": false,
				"dataset_query": map[string]any{
					"database": 1, "type": "native",
					"native": map[string]any{
						"query": "select * from retention where timeframe_days = {{timeframe_days}}",
						"template-tags": map[string]any{
							"timeframe_days": map[string]any{"id": "t1", "name": "timeframe_days", "type": "number"},
						},
					},
				},
			})
		case "/api/card/1/query":
			// Metabase rejects invalid card parameters with a 500.
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"data":{"type":"invalid-parameter"}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	_, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "card", "run", "1", "--param", "bad=1", "--error-format", "json")
	if err == nil {
		t.Fatal("expected card run to fail for an invalid parameter")
	}

	var je struct {
		Error struct {
			Type       string `json:"type"`
			Suggestion string `json:"suggestion"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(firstJSONLine(stderr)), &je); err != nil {
		t.Fatalf("failed to decode structured error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(je.Error.Suggestion, "mb-cli card params 1") {
		t.Fatalf("expected suggestion to name 'mb-cli card params 1', got %q", je.Error.Suggestion)
	}
}

func TestDashboardRunCardInvalidParameterSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "Merchant Retention", "archived": false,
				"parameters": []map[string]any{
					{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="},
				},
			})
		case "/api/card/50":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 50, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "native", "archived": false,
			})
		case "/api/dashboard/1/dashcard/100/card/50/query":
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid parameter"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	_, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "dashboard", "run-card", "1", "100", "50", "--param", "bad=1", "--error-format", "json")
	if err == nil {
		t.Fatal("expected dashboard run-card to fail for an invalid parameter")
	}

	var je struct {
		Error struct {
			Suggestion string `json:"suggestion"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(firstJSONLine(stderr)), &je); err != nil {
		t.Fatalf("failed to decode structured error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(je.Error.Suggestion, "mb-cli dashboard params list 1") {
		t.Fatalf("expected suggestion to name 'mb-cli dashboard params list 1', got %q", je.Error.Suggestion)
	}
}

func TestParameterizedErrorSuggestionNamesInspectCommand(t *testing.T) {
	withInspect := &mberr.ParameterizedQueryError{
		Err:     &mberr.APIError{StatusCode: 400, Body: "bad request"},
		Inspect: "mb-cli card params 5",
	}
	errType, suggestion := cli.ClassifyError(withInspect)
	if errType != "API_ERROR" {
		t.Errorf("expected API_ERROR, got %s", errType)
	}
	if !strings.Contains(suggestion, "mb-cli card params 5") {
		t.Fatalf("expected suggestion to name the inspect command, got %q", suggestion)
	}

	withoutInspect := &mberr.ParameterizedQueryError{Err: &mberr.APIError{StatusCode: 400, Body: "bad request"}}
	if _, fallback := cli.ClassifyError(withoutInspect); fallback == "" {
		t.Error("expected a fallback suggestion when no inspect command is set")
	}
}
