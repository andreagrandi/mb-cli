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

func setupDatabaseTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestListDatabases(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/" {
			t.Errorf("expected path '/api/database/', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Production", "engine": "postgres"},
				{"id": 2, "name": "Analytics", "engine": "bigquery"},
			},
		})
	})
	defer server.Close()

	databases, err := c.ListDatabases(context.Background(), false)
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}

	if len(databases) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(databases))
	}

	if databases[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", databases[0].ID)
	}
	if databases[0].Name != "Production" {
		t.Errorf("expected name 'Production', got %s", databases[0].Name)
	}
	if databases[0].Engine != "postgres" {
		t.Errorf("expected engine 'postgres', got %s", databases[0].Engine)
	}
	if databases[1].Name != "Analytics" {
		t.Errorf("expected name 'Analytics', got %s", databases[1].Name)
	}
}

func TestListDatabasesWithTables(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include") != "tables" {
			t.Errorf("expected include=tables query param, got '%s'", r.URL.Query().Get("include"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id": 1, "name": "Production", "engine": "postgres",
					"tables": []map[string]any{
						{"id": 10, "name": "users", "schema": "public"},
					},
				},
			},
		})
	})
	defer server.Close()

	databases, err := c.ListDatabases(context.Background(), true)
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}

	if len(databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(databases))
	}
	if len(databases[0].Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(databases[0].Tables))
	}
	if databases[0].Tables[0].Name != "users" {
		t.Errorf("expected table name 'users', got %s", databases[0].Tables[0].Name)
	}
}

func TestGetDatabase(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1" {
			t.Errorf("expected path '/api/database/1', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1, "name": "Production", "engine": "postgres",
		})
	})
	defer server.Close()

	db, err := c.GetDatabase(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDatabase failed: %v", err)
	}

	if db.ID != 1 {
		t.Errorf("expected ID 1, got %d", db.ID)
	}
	if db.Name != "Production" {
		t.Errorf("expected name 'Production', got %s", db.Name)
	}
}

func TestGetDatabaseMetadata(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/metadata" {
			t.Errorf("expected path '/api/database/1/metadata', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1, "name": "Production", "engine": "postgres",
			"tables": []map[string]any{
				{
					"id": 10, "name": "users", "display_name": "Users", "schema": "public", "db_id": 1,
					"fields": []map[string]any{
						{"id": 100, "name": "id", "display_name": "ID", "base_type": "type/Integer", "database_type": "int4", "table_id": 10},
						{"id": 101, "name": "email", "display_name": "Email", "base_type": "type/Text", "database_type": "varchar", "table_id": 10},
					},
				},
			},
		})
	})
	defer server.Close()

	meta, err := c.GetDatabaseMetadata(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDatabaseMetadata failed: %v", err)
	}

	if meta.ID != 1 {
		t.Errorf("expected ID 1, got %d", meta.ID)
	}
	if len(meta.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(meta.Tables))
	}
	if meta.Tables[0].Name != "users" {
		t.Errorf("expected table name 'users', got %s", meta.Tables[0].Name)
	}
	if len(meta.Tables[0].Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(meta.Tables[0].Fields))
	}
	if meta.Tables[0].Fields[0].Name != "id" {
		t.Errorf("expected field name 'id', got %s", meta.Tables[0].Fields[0].Name)
	}
	if meta.Tables[0].Fields[1].BaseType != "type/Text" {
		t.Errorf("expected base type 'type/Text', got %s", meta.Tables[0].Fields[1].BaseType)
	}
}

func TestGetDatabaseFields(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/fields" {
			t.Errorf("expected path '/api/database/1/fields', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 100, "name": "id", "display_name": "ID", "base_type": "type/Integer", "database_type": "int4", "table_id": 10, "table_name": "users"},
			{"id": 101, "name": "email", "display_name": "Email", "base_type": "type/Text", "database_type": "varchar", "table_id": 10, "table_name": "users"},
		})
	})
	defer server.Close()

	fields, err := c.GetDatabaseFields(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDatabaseFields failed: %v", err)
	}

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Name != "id" {
		t.Errorf("expected field name 'id', got %s", fields[0].Name)
	}
	if fields[1].TableName != "users" {
		t.Errorf("expected table name 'users', got %s", fields[1].TableName)
	}
}

func TestListDatabaseSchemas(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/schemas" {
			t.Errorf("expected path '/api/database/1/schemas', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"public", "analytics", "internal"})
	})
	defer server.Close()

	schemas, err := c.ListDatabaseSchemas(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListDatabaseSchemas failed: %v", err)
	}

	if len(schemas) != 3 {
		t.Fatalf("expected 3 schemas, got %d", len(schemas))
	}
	if schemas[0] != "public" {
		t.Errorf("expected schema 'public', got %s", schemas[0])
	}
	if schemas[2] != "internal" {
		t.Errorf("expected schema 'internal', got %s", schemas[2])
	}
}

func TestGetDatabaseSchema(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/database/1/schema/public" {
			t.Errorf("expected path '/api/database/1/schema/public', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 10, "name": "users", "display_name": "Users", "schema": "public", "db_id": 1},
			{"id": 11, "name": "orders", "display_name": "Orders", "schema": "public", "db_id": 1},
		})
	})
	defer server.Close()

	tables, err := c.GetDatabaseSchema(context.Background(), 1, "public")
	if err != nil {
		t.Fatalf("GetDatabaseSchema failed: %v", err)
	}

	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "users" {
		t.Errorf("expected table name 'users', got %s", tables[0].Name)
	}
	if tables[1].Name != "orders" {
		t.Errorf("expected table name 'orders', got %s", tables[1].Name)
	}
}

func TestGetDatabaseNotFound(t *testing.T) {
	c, server := setupDatabaseTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	})
	defer server.Close()

	_, err := c.GetDatabase(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
