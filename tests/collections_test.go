package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
)

func setupCollectionTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestListCollections(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collection/" {
			t.Errorf("expected path '/api/collection/', got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Finance", "description": "Revenue reports", "location": "/", "archived": false},
			{"id": 2, "name": "Marketing", "location": "/", "archived": true},
		})
	})
	defer server.Close()

	collections, err := c.ListCollections(context.Background())
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}
	if collections[0].Name != "Finance" {
		t.Errorf("expected first collection name Finance, got %s", collections[0].Name)
	}
	if !collections[1].Archived {
		t.Error("expected second collection to be archived")
	}
}

func TestGetCollection(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collection/7" {
			t.Errorf("expected path '/api/collection/7', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":          7,
			"name":        "Finance",
			"description": "Revenue reports",
			"slug":        "finance",
			"location":    "/",
			"archived":    false,
		})
	})
	defer server.Close()

	collection, err := c.GetCollection(context.Background(), "7")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if collection.Name != "Finance" {
		t.Errorf("expected collection name Finance, got %s", collection.Name)
	}
	if collection.Slug != "finance" {
		t.Errorf("expected slug finance, got %s", collection.Slug)
	}
}

func TestGetCollectionRoot(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collection/root" {
			t.Errorf("expected path '/api/collection/root', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":   "root",
			"name": "Our analytics",
		})
	})
	defer server.Close()

	collection, err := c.GetCollection(context.Background(), "root")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if collection.ID != "root" {
		t.Errorf("expected root collection ID 'root', got %v", collection.ID)
	}
	if collection.Name != "Our analytics" {
		t.Errorf("expected name 'Our analytics', got %s", collection.Name)
	}
}

func TestGetCollectionNotFound(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
	})
	defer server.Close()

	_, err := c.GetCollection(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error for missing collection")
	}
}

func TestGetCollectionItems(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collection/7/items" {
			t.Errorf("expected path '/api/collection/7/items', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total": 3,
			"data": []map[string]any{
				{"id": 10, "name": "Revenue by Month", "model": "card", "description": "Monthly totals"},
				{"id": 20, "name": "Finance Overview", "model": "dashboard"},
				{"id": 30, "name": "Archived Reports", "model": "collection"},
			},
		})
	})
	defer server.Close()

	items, err := c.GetCollectionItems(context.Background(), "7", nil)
	if err != nil {
		t.Fatalf("GetCollectionItems failed: %v", err)
	}

	if items.Total != 3 {
		t.Errorf("expected total 3, got %d", items.Total)
	}
	if len(items.Data) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items.Data))
	}
	if items.Data[0].Model != "card" || items.Data[0].Name != "Revenue by Month" {
		t.Errorf("unexpected first item: %+v", items.Data[0])
	}
	if items.Data[1].Model != "dashboard" {
		t.Errorf("expected second item model dashboard, got %s", items.Data[1].Model)
	}
}

func TestGetCollectionItemsWithModels(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		models := r.URL.Query()["models"]
		if len(models) != 2 || models[0] != "card" || models[1] != "dashboard" {
			t.Errorf("expected models=[card,dashboard], got %v", models)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total": 1,
			"data": []map[string]any{
				{"id": 10, "name": "Revenue by Month", "model": "card"},
			},
		})
	})
	defer server.Close()

	items, err := c.GetCollectionItems(context.Background(), "7", []string{"card", "dashboard"})
	if err != nil {
		t.Fatalf("GetCollectionItems failed: %v", err)
	}

	if len(items.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items.Data))
	}
}

func TestGetCollectionItemsRoot(t *testing.T) {
	c, server := setupCollectionTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collection/root/items" {
			t.Errorf("expected path '/api/collection/root/items', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"total": 0, "data": []map[string]any{}})
	})
	defer server.Close()

	items, err := c.GetCollectionItems(context.Background(), "root", nil)
	if err != nil {
		t.Fatalf("GetCollectionItems failed: %v", err)
	}

	if len(items.Data) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items.Data))
	}
}

func TestCollectionListCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Finance", "location": "/", "archived": false},
			{"id": 2, "name": "Marketing", "location": "/", "archived": false},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "collection", "list", "-f", "json")
	if err != nil {
		t.Fatalf("collection list failed: %v\nstderr: %s", err, stderr)
	}

	var collections []map[string]any
	if err := json.Unmarshal([]byte(stdout), &collections); err != nil {
		t.Fatalf("failed to decode collection list output: %v\noutput: %s", err, stdout)
	}
	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}
	if collections[0]["name"] != "Finance" {
		t.Errorf("expected first collection Finance, got %v", collections[0]["name"])
	}
}

func TestCollectionItemsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total": 2,
			"data": []map[string]any{
				{"id": 10, "name": "Revenue by Month", "model": "card"},
				{"id": 20, "name": "Finance Overview", "model": "dashboard"},
			},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "collection", "items", "7", "-f", "table")
	if err != nil {
		t.Fatalf("collection items failed: %v\nstderr: %s", err, stderr)
	}

	for _, want := range []string{"model", "Revenue by Month", "dashboard"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected items table output to contain %q, got:\n%s", want, stdout)
		}
	}
}
