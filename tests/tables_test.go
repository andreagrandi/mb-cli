package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
)

func setupTableTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestListTables(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/" {
			t.Errorf("expected path '/api/table/', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 10, "name": "users", "display_name": "Users", "schema": "public", "db_id": 1},
			{"id": 11, "name": "orders", "display_name": "Orders", "schema": "public", "db_id": 1},
			{"id": 12, "name": "products", "display_name": "Products", "schema": "public", "db_id": 1},
		})
	})
	defer server.Close()

	tables, err := c.ListTables(context.Background())
	if err != nil {
		t.Fatalf("ListTables failed: %v", err)
	}

	if len(tables) != 3 {
		t.Fatalf("expected 3 tables, got %d", len(tables))
	}

	if tables[0].ID != 10 {
		t.Errorf("expected ID 10, got %d", tables[0].ID)
	}
	if tables[0].Name != "users" {
		t.Errorf("expected name 'users', got %s", tables[0].Name)
	}
	if tables[0].Schema != "public" {
		t.Errorf("expected schema 'public', got %s", tables[0].Schema)
	}
	if tables[1].DisplayName != "Orders" {
		t.Errorf("expected display name 'Orders', got %s", tables[1].DisplayName)
	}
	if tables[2].DBId != 1 {
		t.Errorf("expected db_id 1, got %d", tables[2].DBId)
	}
}

func TestGetTable(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/10" {
			t.Errorf("expected path '/api/table/10', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 10, "name": "users", "display_name": "Users", "schema": "public", "db_id": 1, "entity_type": "entity/UserTable",
		})
	})
	defer server.Close()

	table, err := c.GetTable(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetTable failed: %v", err)
	}

	if table.ID != 10 {
		t.Errorf("expected ID 10, got %d", table.ID)
	}
	if table.Name != "users" {
		t.Errorf("expected name 'users', got %s", table.Name)
	}
	if table.EntityType != "entity/UserTable" {
		t.Errorf("expected entity type 'entity/UserTable', got %s", table.EntityType)
	}
}

func TestGetTableMetadata(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/10/query_metadata" {
			t.Errorf("expected path '/api/table/10/query_metadata', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 10, "name": "users", "display_name": "Users", "schema": "public", "db_id": 1,
			"fields": []map[string]any{
				{"id": 100, "name": "id", "display_name": "ID", "base_type": "type/Integer", "database_type": "int4", "table_id": 10},
				{"id": 101, "name": "email", "display_name": "Email", "base_type": "type/Text", "database_type": "varchar", "table_id": 10},
				{"id": 102, "name": "created_at", "display_name": "Created At", "base_type": "type/DateTime", "database_type": "timestamp", "table_id": 10},
			},
		})
	})
	defer server.Close()

	meta, err := c.GetTableMetadata(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetTableMetadata failed: %v", err)
	}

	if meta.ID != 10 {
		t.Errorf("expected ID 10, got %d", meta.ID)
	}
	if meta.Name != "users" {
		t.Errorf("expected name 'users', got %s", meta.Name)
	}
	if len(meta.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(meta.Fields))
	}
	if meta.Fields[0].Name != "id" {
		t.Errorf("expected field name 'id', got %s", meta.Fields[0].Name)
	}
	if meta.Fields[1].BaseType != "type/Text" {
		t.Errorf("expected base type 'type/Text', got %s", meta.Fields[1].BaseType)
	}
	if meta.Fields[2].DisplayName != "Created At" {
		t.Errorf("expected display name 'Created At', got %s", meta.Fields[2].DisplayName)
	}
}

func TestGetTableFKs(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/11/fks" {
			t.Errorf("expected path '/api/table/11/fks', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"relationship": "Mt1",
				"origin": map[string]any{
					"id":   200,
					"name": "user_id",
					"table": map[string]any{
						"id":   11,
						"name": "orders",
					},
				},
				"destination": map[string]any{
					"id":   100,
					"name": "id",
					"table": map[string]any{
						"id":   10,
						"name": "users",
					},
				},
			},
		})
	})
	defer server.Close()

	fks, err := c.GetTableFKs(context.Background(), 11)
	if err != nil {
		t.Fatalf("GetTableFKs failed: %v", err)
	}

	if len(fks) != 1 {
		t.Fatalf("expected 1 foreign key, got %d", len(fks))
	}

	if fks[0].Relationship != "Mt1" {
		t.Errorf("expected relationship 'Mt1', got %s", fks[0].Relationship)
	}
	if fks[0].Origin.Name != "user_id" {
		t.Errorf("expected origin name 'user_id', got %s", fks[0].Origin.Name)
	}
	if fks[0].Origin.Table.Name != "orders" {
		t.Errorf("expected origin table 'orders', got %s", fks[0].Origin.Table.Name)
	}
	if fks[0].Destination.Name != "id" {
		t.Errorf("expected destination name 'id', got %s", fks[0].Destination.Name)
	}
	if fks[0].Destination.Table.Name != "users" {
		t.Errorf("expected destination table 'users', got %s", fks[0].Destination.Table.Name)
	}
}

func TestGetTableData(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/table/10/data" {
			t.Errorf("expected path '/api/table/10/data', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"cols": []map[string]any{
					{"name": "id", "display_name": "ID", "base_type": "type/Integer"},
					{"name": "email", "display_name": "Email", "base_type": "type/Text"},
				},
				"rows": [][]any{
					{1, "alice@example.com"},
					{2, "bob@example.com"},
				},
			},
		})
	})
	defer server.Close()

	result, err := c.GetTableData(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetTableData failed: %v", err)
	}

	if len(result.Data.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result.Data.Columns))
	}
	if result.Data.Columns[0].Name != "id" {
		t.Errorf("expected column name 'id', got %s", result.Data.Columns[0].Name)
	}
	if result.Data.Columns[1].Name != "email" {
		t.Errorf("expected column name 'email', got %s", result.Data.Columns[1].Name)
	}

	if len(result.Data.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Data.Rows))
	}
}

func TestGetTableFKsEmpty(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{})
	})
	defer server.Close()

	fks, err := c.GetTableFKs(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetTableFKs failed: %v", err)
	}

	if len(fks) != 0 {
		t.Errorf("expected 0 foreign keys, got %d", len(fks))
	}
}

func TestGetTableNotFound(t *testing.T) {
	c, server := setupTableTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	})
	defer server.Close()

	_, err := c.GetTable(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
