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

func setupFieldTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestGetField(t *testing.T) {
	c, server := setupFieldTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/field/100" {
			t.Errorf("expected path '/api/field/100', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":            100,
			"name":          "email",
			"display_name":  "Email",
			"base_type":     "type/Text",
			"database_type": "varchar",
			"semantic_type": "type/Email",
			"table_id":      10,
			"table_name":    "users",
		})
	})
	defer server.Close()

	field, err := c.GetField(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetField failed: %v", err)
	}

	if field.ID != 100 {
		t.Errorf("expected ID 100, got %d", field.ID)
	}
	if field.Name != "email" {
		t.Errorf("expected name 'email', got %s", field.Name)
	}
	if field.DisplayName != "Email" {
		t.Errorf("expected display name 'Email', got %s", field.DisplayName)
	}
	if field.BaseType != "type/Text" {
		t.Errorf("expected base type 'type/Text', got %s", field.BaseType)
	}
	if field.DatabaseType != "varchar" {
		t.Errorf("expected database type 'varchar', got %s", field.DatabaseType)
	}
	if field.SemanticType != "type/Email" {
		t.Errorf("expected semantic type 'type/Email', got %s", field.SemanticType)
	}
	if field.TableID != 10 {
		t.Errorf("expected table ID 10, got %d", field.TableID)
	}
	if field.TableName != "users" {
		t.Errorf("expected table name 'users', got %s", field.TableName)
	}
}

func TestGetFieldSummary(t *testing.T) {
	c, server := setupFieldTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/field/100/summary" {
			t.Errorf("expected path '/api/field/100/summary', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		// API returns pairs: [["count",1500],["distincts",1450]]
		json.NewEncoder(w).Encode([][]any{
			{"count", 1500},
			{"distincts", 1450},
		})
	})
	defer server.Close()

	summary, err := c.GetFieldSummary(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetFieldSummary failed: %v", err)
	}

	if len(summary) != 2 {
		t.Fatalf("expected 2 summary entries, got %d", len(summary))
	}
	if summary[0].Type != "count" {
		t.Errorf("expected type 'count', got %s", summary[0].Type)
	}
	if summary[1].Type != "distincts" {
		t.Errorf("expected type 'distincts', got %s", summary[1].Type)
	}
}

func TestGetFieldValues(t *testing.T) {
	c, server := setupFieldTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/field/101/values" {
			t.Errorf("expected path '/api/field/101/values', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"field_id": 101,
			"values":   [][]any{{"active"}, {"inactive"}, {"pending"}},
		})
	})
	defer server.Close()

	values, err := c.GetFieldValues(context.Background(), 101)
	if err != nil {
		t.Fatalf("GetFieldValues failed: %v", err)
	}

	if values.FieldID != 101 {
		t.Errorf("expected field ID 101, got %d", values.FieldID)
	}
	if len(values.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values.Values))
	}
}

func TestGetFieldValuesEmpty(t *testing.T) {
	c, server := setupFieldTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"field_id": 102,
			"values":   [][]any{},
		})
	})
	defer server.Close()

	values, err := c.GetFieldValues(context.Background(), 102)
	if err != nil {
		t.Fatalf("GetFieldValues failed: %v", err)
	}

	if len(values.Values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values.Values))
	}
}

func TestGetFieldNotFound(t *testing.T) {
	c, server := setupFieldTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	})
	defer server.Close()

	_, err := c.GetField(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
