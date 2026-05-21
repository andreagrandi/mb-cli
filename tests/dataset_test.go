package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/cli"
	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
)

func setupDatasetTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestRunNativeQuery(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dataset/" {
			t.Errorf("expected path '/api/dataset/', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req client.DatasetQuery
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Database != 1 {
			t.Errorf("expected database 1, got %d", req.Database)
		}
		if req.Type != "native" {
			t.Errorf("expected type 'native', got %s", req.Type)
		}
		if req.Native == nil || req.Native.Query != "SELECT 1" {
			t.Errorf("expected query 'SELECT 1', got %v", req.Native)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"cols": []map[string]any{
					{"name": "id", "display_name": "ID", "base_type": "type/Integer"},
					{"name": "name", "display_name": "Name", "base_type": "type/Text"},
				},
				"rows": [][]any{
					{1, "Alice"},
					{2, "Bob"},
				},
			},
		})
	})
	defer server.Close()

	result, err := c.RunNativeQuery(context.Background(), 1, "SELECT 1")
	if err != nil {
		t.Fatalf("RunNativeQuery failed: %v", err)
	}

	if len(result.Data.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result.Data.Columns))
	}
	if result.Data.Columns[0].Name != "id" {
		t.Errorf("expected column name 'id', got %s", result.Data.Columns[0].Name)
	}
	if result.Data.Columns[1].Name != "name" {
		t.Errorf("expected column name 'name', got %s", result.Data.Columns[1].Name)
	}
	if len(result.Data.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Data.Rows))
	}
}

func TestExportNativeQuery(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dataset/csv" {
			t.Errorf("expected path '/api/dataset/csv', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var wrapper struct {
			Query client.DatasetQuery `json:"query"`
		}
		if err := json.Unmarshal(body, &wrapper); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if wrapper.Query.Database != 1 {
			t.Errorf("expected database 1, got %d", wrapper.Query.Database)
		}
		if wrapper.Query.Native == nil || wrapper.Query.Native.Query != "SELECT id, name FROM users" {
			t.Errorf("expected query 'SELECT id, name FROM users', got %v", wrapper.Query.Native)
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte("id,name\n1,Alice\n2,Bob\n"))
	})
	defer server.Close()

	data, err := c.ExportNativeQuery(context.Background(), 1, "SELECT id, name FROM users", "csv")
	if err != nil {
		t.Fatalf("ExportNativeQuery failed: %v", err)
	}

	expected := "id,name\n1,Alice\n2,Bob\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestRunNativeQueryError(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid SQL query"}`))
	})
	defer server.Close()

	_, err := c.RunNativeQuery(context.Background(), 1, "INVALID SQL")
	if err == nil {
		t.Fatal("expected error for bad query")
	}
}

func TestMatchDatabaseByNameExactSubstring(t *testing.T) {
	databases := []client.Database{
		{ID: 1, Name: "Production"},
		{ID: 2, Name: "Analytics"},
		{ID: 3, Name: "Staging"},
	}

	id, err := cli.MatchDatabaseByName(databases, "prod")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected 1, got %d", id)
	}
}

func TestMatchDatabaseByNameCaseInsensitive(t *testing.T) {
	databases := []client.Database{
		{ID: 1, Name: "Production"},
		{ID: 2, Name: "Analytics"},
	}

	id, err := cli.MatchDatabaseByName(databases, "ANALYTICS")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 2 {
		t.Errorf("expected 2, got %d", id)
	}
}

func TestMatchDatabaseByNameNoMatch(t *testing.T) {
	databases := []client.Database{
		{ID: 1, Name: "Production"},
	}

	_, err := cli.MatchDatabaseByName(databases, "nonexistent")
	if err == nil {
		t.Fatal("expected error for no match")
	}
	expected := "no database matching 'nonexistent' found"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestMatchDatabaseByNameAmbiguous(t *testing.T) {
	databases := []client.Database{
		{ID: 1, Name: "Production DB"},
		{ID: 2, Name: "Production Replica"},
	}

	_, err := cli.MatchDatabaseByName(databases, "production")
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
	if !strings.Contains(err.Error(), "ambiguous database name") {
		t.Errorf("expected ambiguous error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Production DB (id=1)") {
		t.Errorf("expected match list in error, got %q", err.Error())
	}
}

func TestRunStructuredQuery(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dataset/" {
			t.Errorf("expected path '/api/dataset/', got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if raw["type"] != "query" {
			t.Errorf("expected type 'query', got %v", raw["type"])
		}
		if raw["database"].(float64) != 1 {
			t.Errorf("expected database 1, got %v", raw["database"])
		}

		query := raw["query"].(map[string]any)
		if query["source-table"].(float64) != 42 {
			t.Errorf("expected source-table 42, got %v", query["source-table"])
		}

		filter := query["filter"].([]any)
		if filter[0] != "=" {
			t.Errorf("expected filter operator '=', got %v", filter[0])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"cols": []map[string]any{
					{"name": "id", "display_name": "ID", "base_type": "type/Integer"},
				},
				"rows": [][]any{
					{1},
				},
			},
		})
	})
	defer server.Close()

	filters := [][]any{
		{"=", []any{"field", 100, nil}, "prod_1234"},
	}

	result, err := c.RunStructuredQuery(context.Background(), 1, 42, filters, 0)
	if err != nil {
		t.Fatalf("RunStructuredQuery failed: %v", err)
	}

	if len(result.Data.Columns) != 1 {
		t.Fatalf("expected 1 column, got %d", len(result.Data.Columns))
	}
	if len(result.Data.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Data.Rows))
	}
}

func TestRunStructuredQueryMultipleFilters(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]any
		json.Unmarshal(body, &raw)

		query := raw["query"].(map[string]any)
		filter := query["filter"].([]any)

		if filter[0] != "and" {
			t.Errorf("expected 'and' combinator, got %v", filter[0])
		}
		if len(filter) != 3 {
			t.Errorf("expected 3 elements (and + 2 filters), got %d", len(filter))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"cols": []map[string]any{},
				"rows": [][]any{},
			},
		})
	})
	defer server.Close()

	filters := [][]any{
		{"=", []any{"field", 100, nil}, "alice"},
		{"=", []any{"field", 101, nil}, "true"},
	}

	_, err := c.RunStructuredQuery(context.Background(), 1, 42, filters, 0)
	if err != nil {
		t.Fatalf("RunStructuredQuery failed: %v", err)
	}
}

func TestRunStructuredQueryWithLimit(t *testing.T) {
	c, server := setupDatasetTestClient(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]any
		json.Unmarshal(body, &raw)

		query := raw["query"].(map[string]any)
		if query["limit"].(float64) != 10 {
			t.Errorf("expected limit 10, got %v", query["limit"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"cols": []map[string]any{},
				"rows": [][]any{},
			},
		})
	})
	defer server.Close()

	filters := [][]any{
		{"=", []any{"field", 100, nil}, "pending"},
	}

	_, err := c.RunStructuredQuery(context.Background(), 1, 42, filters, 10)
	if err != nil {
		t.Fatalf("RunStructuredQuery failed: %v", err)
	}
}

func TestMatchTableByNameExactSubstring(t *testing.T) {
	tables := []client.TableMetadata{
		{ID: 1, Name: "users"},
		{ID: 2, Name: "orders"},
		{ID: 3, Name: "products"},
	}

	id, err := cli.MatchTableByName(tables, "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected 1, got %d", id)
	}
}

func TestMatchTableByNameNoMatch(t *testing.T) {
	tables := []client.TableMetadata{
		{ID: 1, Name: "users"},
	}

	_, err := cli.MatchTableByName(tables, "nonexistent")
	if err == nil {
		t.Fatal("expected error for no match")
	}
	if !strings.Contains(err.Error(), "no table matching") {
		t.Errorf("expected 'no table matching' error, got %q", err.Error())
	}
}

func TestMatchTableByNameAmbiguous(t *testing.T) {
	tables := []client.TableMetadata{
		{ID: 1, Name: "user_accounts"},
		{ID: 2, Name: "user_profiles"},
	}

	_, err := cli.MatchTableByName(tables, "user")
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
	if !strings.Contains(err.Error(), "ambiguous table name") {
		t.Errorf("expected ambiguous error, got %q", err.Error())
	}
}

func TestParseWhereClause(t *testing.T) {
	tests := []struct {
		input     string
		field     string
		value     string
		expectErr bool
	}{
		{"id=prod_1234", "id", "prod_1234", false},
		{"name=alice", "name", "alice", false},
		{"status=", "status", "", false},
		{"no_equals", "", "", true},
		{"=no_field", "", "", true},
	}

	for _, tt := range tests {
		field, value, err := cli.ParseWhereClause(tt.input)
		if tt.expectErr {
			if err == nil {
				t.Errorf("ParseWhereClause(%q): expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseWhereClause(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if field != tt.field {
			t.Errorf("ParseWhereClause(%q): field = %q, want %q", tt.input, field, tt.field)
		}
		if value != tt.value {
			t.Errorf("ParseWhereClause(%q): value = %q, want %q", tt.input, value, tt.value)
		}
	}
}
