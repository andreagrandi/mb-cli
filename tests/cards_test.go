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

func setupCardTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestListCards(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/card/" {
			t.Errorf("expected path '/api/card/', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "User Count", "description": "Total users", "database_id": 1, "display": "scalar", "query_type": "native", "archived": false},
			{"id": 2, "name": "Revenue by Month", "description": "Monthly revenue", "database_id": 1, "display": "bar", "query_type": "query", "archived": false},
		})
	})
	defer server.Close()

	cards, err := c.ListCards(context.Background())
	if err != nil {
		t.Fatalf("ListCards failed: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", cards[0].ID)
	}
	if cards[0].Name != "User Count" {
		t.Errorf("expected name 'User Count', got %s", cards[0].Name)
	}
	if cards[0].Display != "scalar" {
		t.Errorf("expected display 'scalar', got %s", cards[0].Display)
	}
	if cards[1].Name != "Revenue by Month" {
		t.Errorf("expected name 'Revenue by Month', got %s", cards[1].Name)
	}
}

func TestGetCard(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/card/1" {
			t.Errorf("expected path '/api/card/1', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 1, "name": "User Count", "description": "Total users",
			"database_id": 1, "display": "scalar", "query_type": "native", "archived": false,
		})
	})
	defer server.Close()

	card, err := c.GetCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}

	if card.ID != 1 {
		t.Errorf("expected ID 1, got %d", card.ID)
	}
	if card.Name != "User Count" {
		t.Errorf("expected name 'User Count', got %s", card.Name)
	}
	if card.DatabaseID != 1 {
		t.Errorf("expected database_id 1, got %d", card.DatabaseID)
	}
}

func TestRunCard(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id": 1, "name": "User Count", "database_id": 1, "display": "scalar", "query_type": "native", "archived": false,
			})
		case "/api/card/1/query":
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cols": []map[string]any{
						{"name": "count", "display_name": "Count", "base_type": "type/Integer"},
					},
					"rows": [][]any{
						{42},
					},
				},
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	defer server.Close()

	result, err := c.RunCard(context.Background(), 1)
	if err != nil {
		t.Fatalf("RunCard failed: %v", err)
	}

	if len(result.Data.Columns) != 1 {
		t.Fatalf("expected 1 column, got %d", len(result.Data.Columns))
	}
	if result.Data.Columns[0].Name != "count" {
		t.Errorf("expected column name 'count', got %s", result.Data.Columns[0].Name)
	}
	if len(result.Data.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Data.Rows))
	}
}

func TestGetCardNotFound(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	})
	defer server.Close()

	_, err := c.GetCard(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestRunCardNotFound(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	})
	defer server.Close()

	_, err := c.RunCard(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestCardGetFull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/card/1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":          1,
			"name":        "Revenue by Merchant",
			"description": "Full card payload",
			"database_id": 1,
			"display":     "table",
			"query_type":  "native",
			"archived":    false,
			"dataset_query": map[string]any{
				"database": 1,
				"type":     "native",
				"native": map[string]any{
					"query": "select * from orders where merchant_id = {{merchant_id}}",
					"template-tags": map[string]any{
						"merchant_id": map[string]any{"id": "merchant_id", "name": "merchant_id", "type": "number"},
					},
				},
			},
			"visualization_settings": map[string]any{"table.columns": []string{"merchant_id", "revenue"}},
		})
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "card", "get", "1", "--full", "-f", "json")
	if err != nil {
		t.Fatalf("card get --full failed: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "dataset_query") {
		t.Fatalf("expected dataset_query in full card output, got %s", stdout)
	}
	if !strings.Contains(stdout, "merchant_id") {
		t.Fatalf("expected template tag in full card output, got %s", stdout)
	}
}

func TestParameterizedCardRun(t *testing.T) {
	c, server := setupCardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/card/1":
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
							"timeframe_days": map[string]any{"id": "timeframe_days", "name": "timeframe_days", "type": "number"},
						},
					},
				},
			})
		case "/api/card/1/query":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			parameters, ok := body["parameters"].([]any)
			if !ok || len(parameters) != 1 {
				t.Fatalf("expected one parameter, got %v", body["parameters"])
			}
			parameter := parameters[0].(map[string]any)
			if parameter["id"] != "timeframe_days" {
				t.Fatalf("expected parameter id timeframe_days, got %v", parameter["id"])
			}
			if parameter["value"] != float64(14) {
				t.Fatalf("expected parameter value 14, got %v", parameter["value"])
			}

			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cols": []map[string]any{{"name": "count", "display_name": "Count", "base_type": "type/Integer"}},
					"rows": [][]any{{12}},
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer server.Close()

	result, err := c.RunCardWithParams(context.Background(), 1, map[string]string{"timeframe_days": "14"})
	if err != nil {
		t.Fatalf("RunCardWithParams failed: %v", err)
	}

	if len(result.Data.Rows) != 1 || result.Data.Rows[0][0] != float64(12) {
		t.Fatalf("unexpected parameterized card result: %+v", result.Data.Rows)
	}
}
