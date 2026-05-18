package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
)

func TestIsPIISemanticType(t *testing.T) {
	tests := []struct {
		semanticType string
		expected     bool
	}{
		{"type/Email", true},
		{"type/Name", true},
		{"type/Phone", true},
		{"type/Address", true},
		{"type/City", true},
		{"type/State", true},
		{"type/ZipCode", true},
		{"type/Country", true},
		{"type/Latitude", true},
		{"type/Longitude", true},
		{"type/Birthdate", true},
		{"type/AvatarURL", true},
		{"type/URL", true},
		{"type/ImageURL", true},
		{"type/Company", true},
		{"type/FK", false},
		{"type/PK", false},
		{"type/Category", false},
		{"type/Number", false},
		{"type/Description", false},
		{"", false},
		{"type/Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.semanticType, func(t *testing.T) {
			result := client.IsPIISemanticType(tt.semanticType)
			if result != tt.expected {
				t.Errorf("IsPIISemanticType(%q) = %v, want %v", tt.semanticType, result, tt.expected)
			}
		})
	}
}

func TestRedactQueryResult(t *testing.T) {
	tests := []struct {
		name     string
		input    client.QueryResult
		expected [][]any
	}{
		{
			name: "mixed PII and non-PII columns",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "id", BaseType: "type/Integer"},
						{Name: "email", BaseType: "type/Text", SemanticType: "type/Email"},
						{Name: "name", BaseType: "type/Text", SemanticType: "type/Name"},
					},
					Rows: [][]any{
						{1, "alice@example.com", "Alice"},
						{2, "bob@example.com", "Bob"},
					},
				},
			},
			expected: [][]any{
				{1, client.RedactedValue, client.RedactedValue},
				{2, client.RedactedValue, client.RedactedValue},
			},
		},
		{
			name: "no PII columns",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "id", BaseType: "type/Integer"},
						{Name: "count", BaseType: "type/Integer", SemanticType: "type/Number"},
					},
					Rows: [][]any{
						{1, 42},
						{2, 99},
					},
				},
			},
			expected: [][]any{
				{1, 42},
				{2, 99},
			},
		},
		{
			name: "all PII columns",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "email", BaseType: "type/Text", SemanticType: "type/Email"},
						{Name: "phone", BaseType: "type/Text", SemanticType: "type/Phone"},
					},
					Rows: [][]any{
						{"alice@example.com", "+1234567890"},
					},
				},
			},
			expected: [][]any{
				{client.RedactedValue, client.RedactedValue},
			},
		},
		{
			name: "empty rows",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "email", BaseType: "type/Text", SemanticType: "type/Email"},
					},
					Rows: [][]any{},
				},
			},
			expected: [][]any{},
		},
		{
			name: "nil values in PII columns",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "id", BaseType: "type/Integer"},
						{Name: "email", BaseType: "type/Text", SemanticType: "type/Email"},
					},
					Rows: [][]any{
						{1, nil},
					},
				},
			},
			expected: [][]any{
				{1, client.RedactedValue},
			},
		},
		{
			name: "columns with empty semantic type",
			input: client.QueryResult{
				Data: client.QueryResultData{
					Columns: []client.ResultColumn{
						{Name: "data", BaseType: "type/Text"},
					},
					Rows: [][]any{
						{"some data"},
					},
				},
			},
			expected: [][]any{
				{"some data"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.RedactQueryResult(&tt.input)
			if len(tt.input.Data.Rows) != len(tt.expected) {
				t.Fatalf("row count = %d, want %d", len(tt.input.Data.Rows), len(tt.expected))
			}
			for i, row := range tt.input.Data.Rows {
				for j, val := range row {
					if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", tt.expected[i][j]) {
						t.Errorf("row[%d][%d] = %v, want %v", i, j, val, tt.expected[i][j])
					}
				}
			}
		})
	}
}

func TestRedactFieldValues(t *testing.T) {
	tests := []struct {
		name     string
		input    client.FieldValues
		expected [][]any
	}{
		{
			name: "basic redaction",
			input: client.FieldValues{
				FieldID: 1,
				Values:  [][]any{{"alice@example.com"}, {"bob@example.com"}},
			},
			expected: [][]any{{client.RedactedValue}, {client.RedactedValue}},
		},
		{
			name: "empty values",
			input: client.FieldValues{
				FieldID: 1,
				Values:  [][]any{},
			},
			expected: [][]any{},
		},
		{
			name: "multi-element value arrays",
			input: client.FieldValues{
				FieldID: 1,
				Values:  [][]any{{"alice@example.com", "Alice"}, {"bob@example.com", "Bob"}},
			},
			expected: [][]any{{client.RedactedValue, client.RedactedValue}, {client.RedactedValue, client.RedactedValue}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.RedactFieldValues(&tt.input)
			if len(tt.input.Values) != len(tt.expected) {
				t.Fatalf("values count = %d, want %d", len(tt.input.Values), len(tt.expected))
			}
			for i, vals := range tt.input.Values {
				for j, val := range vals {
					if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", tt.expected[i][j]) {
						t.Errorf("values[%d][%d] = %v, want %v", i, j, val, tt.expected[i][j])
					}
				}
			}
		})
	}
}

func TestRedactPIIIntegration(t *testing.T) {
	queryResponse := client.QueryResult{
		Data: client.QueryResultData{
			Columns: []client.ResultColumn{
				{Name: "id", DisplayName: "ID", BaseType: "type/Integer"},
				{Name: "email", DisplayName: "Email", BaseType: "type/Text", SemanticType: "type/Email"},
				{Name: "name", DisplayName: "Name", BaseType: "type/Text", SemanticType: "type/Name"},
			},
			Rows: [][]any{
				{float64(1), "alice@example.com", "Alice"},
				{float64(2), "bob@example.com", "Bob"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(queryResponse)
	}))
	defer server.Close()

	t.Run("redaction enabled", func(t *testing.T) {
		c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
		c.RedactPII = true

		result, err := c.RunNativeQuery(1, "SELECT id, email, name FROM users")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, row := range result.Data.Rows {
			if row[1] != client.RedactedValue {
				t.Errorf("email not redacted: got %v", row[1])
			}
			if row[2] != client.RedactedValue {
				t.Errorf("name not redacted: got %v", row[2])
			}
		}

		if result.Data.Rows[0][0] != float64(1) {
			t.Errorf("id should not be redacted: got %v", result.Data.Rows[0][0])
		}
	})

	t.Run("redaction disabled", func(t *testing.T) {
		c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
		c.RedactPII = false

		result, err := c.RunNativeQuery(1, "SELECT id, email, name FROM users")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Data.Rows[0][1] != "alice@example.com" {
			t.Errorf("email should not be redacted: got %v", result.Data.Rows[0][1])
		}
	})

	t.Run("export blocked when redaction enabled", func(t *testing.T) {
		c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
		c.RedactPII = true

		_, err := c.ExportNativeQuery(1, "SELECT 1", "csv")
		if err == nil {
			t.Fatal("expected error for export with redaction enabled")
		}

		expectedMsg := "export is not supported when PII redaction is enabled"
		if err.Error() != expectedMsg+" (use JSON or table format instead)" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("export allowed when redaction disabled", func(t *testing.T) {
		exportServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte("id,email\n1,alice@example.com\n"))
		}))
		defer exportServer.Close()

		c := client.NewClient(&config.Config{Host: exportServer.URL, APIKey: "test"})
		c.RedactPII = false

		data, err := c.ExportNativeQuery(1, "SELECT 1", "csv")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(data) == 0 {
			t.Error("expected non-empty export data")
		}
	})
}

// piiQueryResponse builds a /api/dataset/ style result with mixed PII and non-PII
// columns so we can assert redaction across the various code paths that produce
// QueryResult values.
func piiQueryResponse() client.QueryResult {
	return client.QueryResult{
		Data: client.QueryResultData{
			Columns: []client.ResultColumn{
				{Name: "id", DisplayName: "ID", BaseType: "type/Integer", SemanticType: "type/PK"},
				{Name: "email", DisplayName: "Email", BaseType: "type/Text", SemanticType: "type/Email"},
				{Name: "name", DisplayName: "Name", BaseType: "type/Text", SemanticType: "type/Name"},
			},
			Rows: [][]any{
				{float64(1), "alice@example.com", "Alice"},
				{float64(2), "bob@example.com", "Bob"},
			},
		},
	}
}

func assertPIIRedacted(t *testing.T, result *client.QueryResult) {
	t.Helper()
	if len(result.Data.Rows) == 0 {
		t.Fatal("expected at least one row")
	}
	for r, row := range result.Data.Rows {
		if row[1] != client.RedactedValue {
			t.Errorf("row[%d] email not redacted: got %v", r, row[1])
		}
		if row[2] != client.RedactedValue {
			t.Errorf("row[%d] name not redacted: got %v", r, row[2])
		}
		if row[0] == client.RedactedValue {
			t.Errorf("row[%d] id should not be redacted: got %v", r, row[0])
		}
	}
}

func TestRunStructuredQueryRedaction(t *testing.T) {
	resp := piiQueryResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dataset/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.RunStructuredQuery(1, 10, [][]any{{"=", []any{"field", 100, nil}, "x"}}, 0)
	if err != nil {
		t.Fatalf("RunStructuredQuery failed: %v", err)
	}
	assertPIIRedacted(t, result)
}

func TestExportStructuredQueryBlockedWhenRedacting(t *testing.T) {
	c := client.NewClient(&config.Config{Host: "http://example.invalid", APIKey: "test"})
	c.RedactPII = true

	_, err := c.ExportStructuredQuery(1, 10, [][]any{{"=", []any{"field", 100, nil}, "x"}}, 0, "csv")
	if err == nil {
		t.Fatal("expected error for export with redaction enabled")
	}
	if !strings.Contains(err.Error(), "export is not supported when PII redaction is enabled") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGetTableDataRedaction(t *testing.T) {
	resp := piiQueryResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/42/data" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.GetTableData(42)
	if err != nil {
		t.Fatalf("GetTableData failed: %v", err)
	}
	assertPIIRedacted(t, result)
}

func TestRunCardRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1, "display": "table", "query_type": "native", "archived": false,
			})
		case "/api/card/7/query":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(piiQueryResponse())
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.RunCard(7)
	if err != nil {
		t.Fatalf("RunCard failed: %v", err)
	}
	assertPIIRedacted(t, result)
}

func TestRunCardEnrichesSemanticTypesWhenMissing(t *testing.T) {
	// Native saved cards sometimes return result cols without semantic_type.
	// The client should fall back to database field metadata for enrichment.
	naked := client.QueryResult{
		Data: client.QueryResultData{
			Columns: []client.ResultColumn{
				{Name: "id", BaseType: "type/Integer"},
				{Name: "email", BaseType: "type/Text"},
				{Name: "name", BaseType: "type/Text"},
			},
			Rows: [][]any{{float64(1), "alice@example.com", "Alice"}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1, "display": "table", "query_type": "native", "archived": false,
			})
		case "/api/card/7/query":
			json.NewEncoder(w).Encode(naked)
		case "/api/database/1/fields":
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": 100, "name": "id", "base_type": "type/Integer"},
				{"id": 101, "name": "email", "base_type": "type/Text", "semantic_type": "type/Email"},
				{"id": 102, "name": "name", "base_type": "type/Text", "semantic_type": "type/Name"},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.RunCard(7)
	if err != nil {
		t.Fatalf("RunCard failed: %v", err)
	}
	if result.Data.Rows[0][1] != client.RedactedValue {
		t.Errorf("email should be redacted after semantic-type enrichment, got %v", result.Data.Rows[0][1])
	}
	if result.Data.Rows[0][2] != client.RedactedValue {
		t.Errorf("name should be redacted after semantic-type enrichment, got %v", result.Data.Rows[0][2])
	}
	if result.Data.Rows[0][0] != float64(1) {
		t.Errorf("id should not be redacted, got %v", result.Data.Rows[0][0])
	}
}

func TestRunCardWithParamsRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1, "display": "table", "query_type": "native", "archived": false,
				"dataset_query": map[string]any{
					"database": 1, "type": "native",
					"native": map[string]any{
						"query": "select * from users where region = {{region}}",
						"template-tags": map[string]any{
							"region": map[string]any{"id": "region", "name": "region", "type": "text"},
						},
					},
				},
			})
		case "/api/card/7/query":
			json.NewEncoder(w).Encode(piiQueryResponse())
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.RunCardWithParams(7, map[string]string{"region": "EU"})
	if err != nil {
		t.Fatalf("RunCardWithParams failed: %v", err)
	}
	assertPIIRedacted(t, result)
}

func TestRunDashboardCardRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "Customers", "archived": false,
				"parameters": []map[string]any{
					{"id": "p1", "name": "Region", "slug": "region", "type": "string/="},
				},
			})
		case "/api/card/7":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 7, "name": "Users", "database_id": 1, "display": "table", "query_type": "query", "archived": false,
			})
		case "/api/dashboard/1/dashcard/200/card/7/query":
			json.NewEncoder(w).Encode(piiQueryResponse())
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
	c.RedactPII = true

	result, err := c.RunDashboardCard(1, 200, 7, map[string]string{"region": "EU"})
	if err != nil {
		t.Fatalf("RunDashboardCard failed: %v", err)
	}
	assertPIIRedacted(t, result)
}

func TestGetFieldValuesRedaction(t *testing.T) {
	piiField := map[string]any{
		"id": 101, "name": "email", "base_type": "type/Text", "semantic_type": "type/Email", "table_id": 10,
	}
	plainField := map[string]any{
		"id": 102, "name": "status", "base_type": "type/Text", "table_id": 10,
	}

	tests := []struct {
		name        string
		fieldID     int
		fieldResp   map[string]any
		valuesResp  [][]any
		redactPII   bool
		wantValues  [][]any
		expectField bool
	}{
		{
			name:        "PII field redacted when enabled",
			fieldID:     101,
			fieldResp:   piiField,
			valuesResp:  [][]any{{"alice@example.com"}, {"bob@example.com"}},
			redactPII:   true,
			wantValues:  [][]any{{client.RedactedValue}, {client.RedactedValue}},
			expectField: true,
		},
		{
			name:        "non-PII field not redacted when enabled",
			fieldID:     102,
			fieldResp:   plainField,
			valuesResp:  [][]any{{"active"}, {"pending"}},
			redactPII:   true,
			wantValues:  [][]any{{"active"}, {"pending"}},
			expectField: true,
		},
		{
			name:        "PII field not redacted when disabled",
			fieldID:     101,
			fieldResp:   piiField,
			valuesResp:  [][]any{{"alice@example.com"}},
			redactPII:   false,
			wantValues:  [][]any{{"alice@example.com"}},
			expectField: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldFetched := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case fmt.Sprintf("/api/field/%d/values", tt.fieldID):
					json.NewEncoder(w).Encode(map[string]any{"field_id": tt.fieldID, "values": tt.valuesResp})
				case fmt.Sprintf("/api/field/%d", tt.fieldID):
					fieldFetched = true
					json.NewEncoder(w).Encode(tt.fieldResp)
				default:
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
			}))
			defer server.Close()

			c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})
			c.RedactPII = tt.redactPII

			values, err := c.GetFieldValues(tt.fieldID)
			if err != nil {
				t.Fatalf("GetFieldValues failed: %v", err)
			}
			if fieldFetched != tt.expectField {
				t.Errorf("field metadata fetched = %v, want %v", fieldFetched, tt.expectField)
			}
			if len(values.Values) != len(tt.wantValues) {
				t.Fatalf("got %d value rows, want %d", len(values.Values), len(tt.wantValues))
			}
			for i := range values.Values {
				for j := range values.Values[i] {
					if fmt.Sprintf("%v", values.Values[i][j]) != fmt.Sprintf("%v", tt.wantValues[i][j]) {
						t.Errorf("values[%d][%d] = %v, want %v", i, j, values.Values[i][j], tt.wantValues[i][j])
					}
				}
			}
		})
	}
}

func TestEnrichSemanticTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/fields" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "id", "base_type": "type/Integer"},
			{"id": 2, "name": "email", "base_type": "type/Text", "semantic_type": "type/Email"},
			{"id": 3, "name": "shadow_email", "base_type": "type/Text"},
			{"id": 4, "name": "shadow_email", "base_type": "type/Text", "semantic_type": "type/Email"},
		})
	}))
	defer server.Close()

	c := client.NewClient(&config.Config{Host: server.URL, APIKey: "test"})

	t.Run("fills in missing semantic types from database metadata", func(t *testing.T) {
		result := &client.QueryResult{
			Data: client.QueryResultData{
				Columns: []client.ResultColumn{
					{Name: "id", BaseType: "type/Integer"},
					{Name: "email", BaseType: "type/Text"},
				},
			},
		}
		c.EnrichSemanticTypes(result, 1)
		if result.Data.Columns[1].SemanticType != "type/Email" {
			t.Errorf("expected email semantic type to be filled in, got %q", result.Data.Columns[1].SemanticType)
		}
	})

	t.Run("preserves PII type when duplicate field names disagree", func(t *testing.T) {
		result := &client.QueryResult{
			Data: client.QueryResultData{
				Columns: []client.ResultColumn{
					{Name: "shadow_email", BaseType: "type/Text"},
				},
			},
		}
		c.EnrichSemanticTypes(result, 1)
		if result.Data.Columns[0].SemanticType != "type/Email" {
			t.Errorf("expected PII semantic type to win for duplicate field name, got %q", result.Data.Columns[0].SemanticType)
		}
	})

	t.Run("no-op when all columns already have semantic types", func(t *testing.T) {
		result := &client.QueryResult{
			Data: client.QueryResultData{
				Columns: []client.ResultColumn{
					{Name: "id", BaseType: "type/Integer", SemanticType: "type/PK"},
					{Name: "email", BaseType: "type/Text", SemanticType: "type/Email"},
				},
			},
		}
		// Use a closed server URL to prove we don't hit the network.
		closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request %s", r.URL.Path)
		}))
		closed.Close()
		nc := client.NewClient(&config.Config{Host: closed.URL, APIKey: "test"})
		nc.EnrichSemanticTypes(result, 1)
		if result.Data.Columns[0].SemanticType != "type/PK" {
			t.Errorf("expected SemanticType preserved, got %q", result.Data.Columns[0].SemanticType)
		}
	})
}
