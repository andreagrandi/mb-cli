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

func setupSearchTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestSearch(t *testing.T) {
	c, server := setupSearchTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search/" {
			t.Errorf("expected path '/api/search/', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("q") != "users" {
			t.Errorf("expected q=users, got %s", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Users Table", "model": "table", "database_id": 1},
				{"id": 5, "name": "Active Users", "model": "card", "database_id": 1},
			},
		})
	})
	defer server.Close()

	results, err := c.Search(context.Background(), "users", nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "Users Table" {
		t.Errorf("expected name 'Users Table', got %s", results[0].Name)
	}
	if results[0].Model != "table" {
		t.Errorf("expected model 'table', got %s", results[0].Model)
	}
	if results[1].Name != "Active Users" {
		t.Errorf("expected name 'Active Users', got %s", results[1].Name)
	}
	if results[1].Model != "card" {
		t.Errorf("expected model 'card', got %s", results[1].Model)
	}
}

func TestSearchWithModels(t *testing.T) {
	c, server := setupSearchTestClient(func(w http.ResponseWriter, r *http.Request) {
		models := r.URL.Query()["models"]
		if len(models) != 2 || models[0] != "table" || models[1] != "card" {
			t.Errorf("expected models=[table,card], got %v", models)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Users Table", "model": "table", "database_id": 1},
			},
		})
	})
	defer server.Close()

	results, err := c.Search(context.Background(), "users", []string{"table", "card"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchEmpty(t *testing.T) {
	c, server := setupSearchTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	})
	defer server.Close()

	results, err := c.Search(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchAPIError(t *testing.T) {
	c, server := setupSearchTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	})
	defer server.Close()

	_, err := c.Search(context.Background(), "users", nil)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestParseModels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"single", "table", []string{"table"}},
		{"multiple", "table,card,database", []string{"table", "card", "database"}},
		{"with spaces", "table, card, database", []string{"table", "card", "database"}},
		{"trailing comma", "table,card,", []string{"table", "card"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.ParseModels(tt.input)
			if tt.expected == nil && result != nil {
				t.Errorf("expected nil, got %v", result)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d models, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}
