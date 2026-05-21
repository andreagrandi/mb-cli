package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/client"
)

// cliEnv returns the environment a mocked CLI run needs to reach a test server.
func cliEnv(serverURL string) map[string]string {
	return map[string]string{
		"MB_HOST":    serverURL,
		"MB_API_KEY": "test-api-key",
	}
}

// --- Card execution -------------------------------------------------------

func TestCLICardRunJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "User Count", "database_id": 1,
				"display": "scalar", "query_type": "native", "archived": false,
			})
		case "/api/card/1/query":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cols": []map[string]any{
						{"name": "count", "display_name": "Count", "base_type": "type/Integer", "semantic_type": "type/Quantity"},
					},
					"rows": [][]any{{42}},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "card", "run", "1", "-f", "json")
	if err != nil {
		t.Fatalf("card run failed: %v\nstderr: %s", err, stderr)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("failed to decode card run output: %v\noutput: %s", err, stdout)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["count"] != float64(42) {
		t.Fatalf("expected count 42, got %v", rows[0]["count"])
	}
}

func TestCLICardRunTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "User Count", "database_id": 1,
				"display": "scalar", "query_type": "native", "archived": false,
			})
		case "/api/card/1/query":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cols": []map[string]any{
						{"name": "count", "display_name": "Count", "base_type": "type/Integer", "semantic_type": "type/Quantity"},
					},
					"rows": [][]any{{42}},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "card", "run", "1", "-f", "table")
	if err != nil {
		t.Fatalf("card run failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "count") {
		t.Fatalf("expected column header 'count' in table output, got %s", stdout)
	}
	if !strings.Contains(stdout, "42") {
		t.Fatalf("expected value '42' in table output, got %s", stdout)
	}
}

func TestCLICardRunWithParamJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "Retention Card", "database_id": 1,
				"display": "table", "query_type": "native", "archived": false,
				"dataset_query": map[string]any{
					"database": 1, "type": "native",
					"native": map[string]any{
						"query": "select * from retention where region = {{region}}",
						"template-tags": map[string]any{
							"region": map[string]any{"id": "region", "name": "region", "type": "text"},
						},
					},
				},
			})
		case "/api/card/1/query":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			parameters, ok := body["parameters"].([]any)
			if !ok || len(parameters) != 1 {
				t.Fatalf("expected one parameter in request, got %v", body["parameters"])
			}
			parameter := parameters[0].(map[string]any)
			if parameter["value"] != "EU" {
				t.Fatalf("expected parameter value EU, got %v", parameter["value"])
			}
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cols": []map[string]any{
						{"name": "count", "display_name": "Count", "base_type": "type/Integer", "semantic_type": "type/Quantity"},
					},
					"rows": [][]any{{7}},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "card", "run", "1", "--param", "region=EU", "-f", "json")
	if err != nil {
		t.Fatalf("card run with param failed: %v\nstderr: %s", err, stderr)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("failed to decode card run output: %v\noutput: %s", err, stdout)
	}
	if len(rows) != 1 || rows[0]["count"] != float64(7) {
		t.Fatalf("unexpected parameterized card run result: %v", rows)
	}
}

func TestCLICardRunRedactsPII(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1,
				"display": "table", "query_type": "native", "archived": false,
			})
		case "/api/card/7/query":
			json.NewEncoder(w).Encode(piiQueryResponse())
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "card", "run", "7", "-f", "json")
	if err != nil {
		t.Fatalf("card run failed: %v\nstderr: %s", err, stderr)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("failed to decode card run output: %v\noutput: %s", err, stdout)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one row")
	}
	for i, row := range rows {
		if row["email"] != client.RedactedValue {
			t.Errorf("row[%d] email not redacted: got %v", i, row["email"])
		}
		if row["name"] != client.RedactedValue {
			t.Errorf("row[%d] name not redacted: got %v", i, row["name"])
		}
		if row["id"] == client.RedactedValue {
			t.Errorf("row[%d] id should not be redacted", i)
		}
	}
}

func TestCLICardRunRedactionDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1,
				"display": "table", "query_type": "native", "archived": false,
			})
		case "/api/card/7/query":
			json.NewEncoder(w).Encode(piiQueryResponse())
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "card", "run", "7", "--redact-pii=false", "-f", "json")
	if err != nil {
		t.Fatalf("card run failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stderr, "PII redaction is disabled") {
		t.Errorf("expected a redaction-disabled warning on stderr, got %s", stderr)
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
		t.Fatalf("failed to decode card run output: %v\noutput: %s", err, stdout)
	}
	if len(rows) == 0 || rows[0]["email"] != "alice@example.com" {
		t.Fatalf("expected unredacted email when redaction is disabled, got %v", rows)
	}
}

// --- Schema discovery -----------------------------------------------------

func TestCLIDatabaseSchemasJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/schemas" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"public", "analytics"})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "database", "schemas", "1", "-f", "json")
	if err != nil {
		t.Fatalf("database schemas failed: %v\nstderr: %s", err, stderr)
	}

	var schemas []string
	if err := json.Unmarshal([]byte(stdout), &schemas); err != nil {
		t.Fatalf("failed to decode schemas output: %v\noutput: %s", err, stdout)
	}
	if len(schemas) != 2 || schemas[0] != "public" || schemas[1] != "analytics" {
		t.Fatalf("unexpected schemas output: %v", schemas)
	}
}

func TestCLIDatabaseSchemaJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/schema/public" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 10, "name": "orders", "display_name": "Orders", "schema": "public", "db_id": 1},
			{"id": 11, "name": "customers", "display_name": "Customers", "schema": "public", "db_id": 1},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "database", "schema", "1", "public", "-f", "json")
	if err != nil {
		t.Fatalf("database schema failed: %v\nstderr: %s", err, stderr)
	}

	var tables []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal([]byte(stdout), &tables); err != nil {
		t.Fatalf("failed to decode schema output: %v\noutput: %s", err, stdout)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "orders" || tables[0].Schema != "public" {
		t.Fatalf("unexpected first table: %+v", tables[0])
	}
}

func TestCLIDatabaseMetadataJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/metadata" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1, "name": "Sales DB", "engine": "postgres",
			"tables": []map[string]any{
				{
					"id": 10, "name": "orders", "display_name": "Orders", "schema": "public", "db_id": 1,
					"fields": []map[string]any{
						{"id": 100, "name": "id", "base_type": "type/Integer"},
						{"id": 101, "name": "total", "base_type": "type/Float"},
					},
				},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "database", "metadata", "1", "-f", "json")
	if err != nil {
		t.Fatalf("database metadata failed: %v\nstderr: %s", err, stderr)
	}

	var meta struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Engine string `json:"engine"`
		Tables []struct {
			Name   string `json:"name"`
			Fields []struct {
				Name string `json:"name"`
			} `json:"fields"`
		} `json:"tables"`
	}
	if err := json.Unmarshal([]byte(stdout), &meta); err != nil {
		t.Fatalf("failed to decode metadata output: %v\noutput: %s", err, stdout)
	}
	if meta.Name != "Sales DB" || meta.Engine != "postgres" {
		t.Fatalf("unexpected database metadata: %+v", meta)
	}
	if len(meta.Tables) != 1 || meta.Tables[0].Name != "orders" {
		t.Fatalf("expected one table 'orders', got %+v", meta.Tables)
	}
	if len(meta.Tables[0].Fields) != 2 {
		t.Fatalf("expected 2 fields on the orders table, got %d", len(meta.Tables[0].Fields))
	}
}

func TestCLITableMetadataJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/10/query_metadata" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 10, "name": "customers", "display_name": "Customers", "schema": "public", "db_id": 1,
			"fields": []map[string]any{
				{"id": 100, "name": "id", "base_type": "type/Integer", "semantic_type": "type/PK"},
				{"id": 101, "name": "email", "base_type": "type/Text", "semantic_type": "type/Email"},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "table", "metadata", "10", "-f", "json")
	if err != nil {
		t.Fatalf("table metadata failed: %v\nstderr: %s", err, stderr)
	}

	var meta struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Fields []struct {
			Name         string `json:"name"`
			SemanticType string `json:"semantic_type"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(stdout), &meta); err != nil {
		t.Fatalf("failed to decode table metadata output: %v\noutput: %s", err, stdout)
	}
	if meta.Name != "customers" {
		t.Fatalf("expected table name 'customers', got %s", meta.Name)
	}
	if len(meta.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(meta.Fields))
	}
	if meta.Fields[1].Name != "email" || meta.Fields[1].SemanticType != "type/Email" {
		t.Fatalf("expected PII-typed email field, got %+v", meta.Fields[1])
	}
}

// --- Parameter lookup -----------------------------------------------------

func paramLookupDashboard() map[string]any {
	return map[string]any{
		"id": 1, "name": "Merchant Retention", "archived": false,
		"parameters": []map[string]any{
			{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="},
		},
		"dashcards": []map[string]any{
			{
				"id": 100, "card_id": 50,
				"parameter_mappings": []map[string]any{
					{"parameter_id": "param-merchant", "card_id": 50, "target": []any{"variable", []any{"template-tag", "merchant_name"}}},
				},
				"card": map[string]any{"id": 50, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "native", "archived": false},
			},
		},
	}
}

func TestCLIDashboardParamValuesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(paramLookupDashboard())
		case "/api/dashboard/1/params/param-merchant/values":
			json.NewEncoder(w).Encode(map[string]any{
				"values": []any{
					[]any{"merchant-a", "Merchant A"},
					[]any{"merchant-b", "Merchant B"},
				},
				"has_more_values": false,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "dashboard", "params", "values", "1", "merchant_name", "-f", "json")
	if err != nil {
		t.Fatalf("dashboard params values failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		DashboardID  int    `json:"dashboard_id"`
		RequestedKey string `json:"requested_key"`
		ResolvedKey  string `json:"resolved_key"`
		Values       []struct {
			Value any    `json:"value"`
			Label string `json:"label"`
		} `json:"values"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode param values output: %v\noutput: %s", err, stdout)
	}
	if result.DashboardID != 1 {
		t.Fatalf("expected dashboard_id 1, got %d", result.DashboardID)
	}
	if result.RequestedKey != "merchant_name" || result.ResolvedKey != "param-merchant" {
		t.Fatalf("expected slug to resolve to parameter id, got requested=%q resolved=%q", result.RequestedKey, result.ResolvedKey)
	}
	if len(result.Values) != 2 || result.Values[0].Label != "Merchant A" {
		t.Fatalf("unexpected parameter values: %+v", result.Values)
	}
}

func TestCLIDashboardParamValuesTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(paramLookupDashboard())
		case "/api/dashboard/1/params/param-merchant/values":
			json.NewEncoder(w).Encode(map[string]any{
				"values":          []any{[]any{"merchant-a", "Merchant A"}},
				"has_more_values": false,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "dashboard", "params", "values", "1", "merchant_name", "-f", "table")
	if err != nil {
		t.Fatalf("dashboard params values failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "merchant_name") {
		t.Fatalf("expected parameter slug in table output, got %s", stdout)
	}
	if !strings.Contains(stdout, "Merchant A") {
		t.Fatalf("expected value label 'Merchant A' in table output, got %s", stdout)
	}
}

func TestCLIDashboardParamSearchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(paramLookupDashboard())
		case "/api/dashboard/1/params/param-merchant/search/acme":
			json.NewEncoder(w).Encode(map[string]any{
				"values":          []any{[]any{"acme", "Acme Corp"}},
				"has_more_values": true,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "dashboard", "params", "search", "1", "merchant_name", "acme", "-f", "json")
	if err != nil {
		t.Fatalf("dashboard params search failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		ResolvedKey   string `json:"resolved_key"`
		Query         string `json:"query"`
		HasMoreValues bool   `json:"has_more_values"`
		Values        []struct {
			Label string `json:"label"`
		} `json:"values"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode param search output: %v\noutput: %s", err, stdout)
	}
	if result.ResolvedKey != "param-merchant" {
		t.Fatalf("expected resolved key 'param-merchant', got %q", result.ResolvedKey)
	}
	if result.Query != "acme" {
		t.Fatalf("expected query 'acme', got %q", result.Query)
	}
	if !result.HasMoreValues {
		t.Fatal("expected has_more_values to be true for the search response")
	}
	if len(result.Values) != 1 || result.Values[0].Label != "Acme Corp" {
		t.Fatalf("unexpected search values: %+v", result.Values)
	}
}

// --- Dashboard analysis ---------------------------------------------------

func TestCLIDashboardAnalyzeTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "Merchant Retention", "description": "30-day retention dashboard", "archived": false,
				"tabs":       []map[string]any{{"id": 10, "name": "Overview"}},
				"parameters": []map[string]any{{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="}},
				"dashcards": []map[string]any{
					{
						"id": 100, "card_id": 10, "dashboard_tab_id": 10,
						"parameter_mappings": []map[string]any{{"parameter_id": "param-merchant", "card_id": 10, "target": []any{"variable", []any{"template-tag", "merchant_name"}}}},
						"card":               map[string]any{"id": 10, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "query", "archived": false},
					},
				},
			})
		case "/api/card/10":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 10, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "query", "archived": false,
				"dataset_query": map[string]any{
					"database": 1, "type": "query",
					"query": map[string]any{"filter": []any{"=", []any{"field", 1, nil}, "merchant-a"}},
				},
				"visualization_settings": map[string]any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not found"}`))
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, cliEnv(server.URL), "dashboard", "analyze", "1", "-f", "table")
	if err != nil {
		t.Fatalf("dashboard analyze failed: %v\nstderr: %s", err, stderr)
	}

	for _, section := range []string{"dashboard_id", "Tabs", "Dashcards", "Parameters", "Base Cards", "Flagged Cards"} {
		if !strings.Contains(stdout, section) {
			t.Fatalf("expected %q section in analyze table output, got %s", section, stdout)
		}
	}
	if !strings.Contains(stdout, "Merchant Retention") {
		t.Fatalf("expected dashboard name in analyze table output, got %s", stdout)
	}
}
