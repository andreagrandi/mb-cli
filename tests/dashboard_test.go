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

func setupDashboardTestClient(handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	cfg := &config.Config{
		Host:   server.URL,
		APIKey: "test-api-key",
	}
	return client.NewClient(cfg), server
}

func TestListDashboards(t *testing.T) {
	c, server := setupDashboardTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/" {
			t.Errorf("expected path '/api/dashboard/', got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Overview", "description": "Top-level KPIs", "archived": false},
			{"id": 2, "name": "Retention", "description": "Monthly retention", "archived": true},
		})
	})
	defer server.Close()

	dashboards, err := c.ListDashboards(context.Background())
	if err != nil {
		t.Fatalf("ListDashboards failed: %v", err)
	}

	if len(dashboards) != 2 {
		t.Fatalf("expected 2 dashboards, got %d", len(dashboards))
	}
	if dashboards[0].Name != "Overview" {
		t.Errorf("expected first dashboard name Overview, got %s", dashboards[0].Name)
	}
	if !dashboards[1].Archived {
		t.Error("expected second dashboard to be archived")
	}
}

func TestGetDashboard(t *testing.T) {
	c, server := setupDashboardTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/1" {
			t.Errorf("expected path '/api/dashboard/1', got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":          1,
			"name":        "Merchant Retention",
			"description": "30-day dashboard",
			"archived":    false,
			"parameters": []map[string]any{
				{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="},
			},
			"tabs": []map[string]any{
				{"id": 10, "name": "Overview"},
			},
			"dashcards": []map[string]any{
				{
					"id":               100,
					"card_id":          50,
					"dashboard_tab_id": 10,
					"parameter_mappings": []map[string]any{
						{"parameter_id": "param-merchant", "card_id": 50, "target": []any{"variable", []any{"template-tag", "merchant_name"}}},
					},
					"card": map[string]any{
						"id":          50,
						"name":        "Retention by Merchant",
						"database_id": 1,
						"display":     "table",
						"query_type":  "native",
						"archived":    false,
					},
				},
			},
		})
	})
	defer server.Close()

	dashboard, err := c.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dashboard.Name != "Merchant Retention" {
		t.Errorf("expected dashboard name Merchant Retention, got %s", dashboard.Name)
	}
	if len(dashboard.Parameters) != 1 || dashboard.Parameters[0].Slug != "merchant_name" {
		t.Fatalf("expected parsed dashboard parameter, got %+v", dashboard.Parameters)
	}
	if len(dashboard.Tabs) != 1 || dashboard.Tabs[0].Name != "Overview" {
		t.Fatalf("expected parsed dashboard tab, got %+v", dashboard.Tabs)
	}
	if len(dashboard.DashCards) != 1 || dashboard.DashCards[0].Card == nil {
		t.Fatalf("expected parsed dashboard card, got %+v", dashboard.DashCards)
	}
	if len(dashboard.DashCards[0].ParameterMappings) != 1 {
		t.Fatalf("expected parameter mappings to be parsed, got %+v", dashboard.DashCards[0].ParameterMappings)
	}
}

func TestGetDashboardNotFound(t *testing.T) {
	c, server := setupDashboardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
	})
	defer server.Close()

	_, err := c.GetDashboard(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing dashboard")
	}
}

func TestGetDashboardCards(t *testing.T) {
	c, server := setupDashboardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":       1,
			"name":     "Card Listing",
			"archived": false,
			"dashcards": []map[string]any{
				{
					"id":      101,
					"card_id": 77,
					"card": map[string]any{
						"id":          77,
						"name":        "Users by Day",
						"database_id": 1,
						"display":     "line",
						"query_type":  "query",
						"archived":    false,
					},
				},
			},
		})
	})
	defer server.Close()

	dashboard, err := c.GetDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if len(dashboard.DashCards) != 1 {
		t.Fatalf("expected 1 dashcard, got %d", len(dashboard.DashCards))
	}
	if dashboard.DashCards[0].CardID == nil || *dashboard.DashCards[0].CardID != 77 {
		t.Fatalf("expected card_id 77, got %+v", dashboard.DashCards[0].CardID)
	}
	if dashboard.DashCards[0].Card.Name != "Users by Day" {
		t.Errorf("expected nested card name Users by Day, got %s", dashboard.DashCards[0].Card.Name)
	}
}

func TestParameterLookup(t *testing.T) {
	c, server := setupDashboardTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1/params/param-merchant/values":
			json.NewEncoder(w).Encode(map[string]any{
				"values": []any{
					[]any{"merchant-a", "Merchant A"},
					[]any{"merchant-b"},
				},
				"has_more_values": false,
			})
		case "/api/dashboard/1/params/param-merchant/search/acme":
			json.NewEncoder(w).Encode(map[string]any{
				"values": []any{
					[]any{"acme", "Acme Corp"},
				},
				"has_more_values": true,
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})
	defer server.Close()

	values, err := c.GetDashboardParamValues(context.Background(), 1, "param-merchant")
	if err != nil {
		t.Fatalf("GetDashboardParamValues failed: %v", err)
	}
	if len(values.Values) != 2 {
		t.Fatalf("expected 2 parameter values, got %d", len(values.Values))
	}
	if values.Values[0].Value != "merchant-a" || values.Values[0].Label != "Merchant A" {
		t.Fatalf("unexpected first parameter value: %+v", values.Values[0])
	}

	searchValues, err := c.SearchDashboardParamValues(context.Background(), 1, "param-merchant", "acme")
	if err != nil {
		t.Fatalf("SearchDashboardParamValues failed: %v", err)
	}
	if !searchValues.HasMoreValues {
		t.Fatal("expected has_more_values to be true for search response")
	}
	if searchValues.Values[0].Label != "Acme Corp" {
		t.Fatalf("unexpected search label: %+v", searchValues.Values[0])
	}
}

func TestDashboardAnalyze(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/dashboard/1":
			json.NewEncoder(w).Encode(map[string]any{
				"id":          1,
				"name":        "Merchant Retention",
				"description": "30-day retention dashboard",
				"archived":    false,
				"tabs":        []map[string]any{{"id": 10, "name": "Overview"}},
				"parameters":  []map[string]any{{"id": "param-merchant", "name": "Merchant", "slug": "merchant_name", "type": "string/="}},
				"dashcards": []map[string]any{
					{
						"id":                 100,
						"card_id":            10,
						"dashboard_tab_id":   10,
						"parameter_mappings": []map[string]any{{"parameter_id": "param-merchant", "card_id": 10, "target": []any{"variable", []any{"template-tag", "merchant_name"}}}},
						"card":               map[string]any{"id": 10, "name": "Retention by Merchant", "database_id": 1, "display": "table", "query_type": "query", "archived": false},
					},
				},
			})
		case "/api/card/10":
			json.NewEncoder(w).Encode(map[string]any{
				"id":          10,
				"name":        "Retention by Merchant",
				"database_id": 1,
				"display":     "table",
				"query_type":  "query",
				"archived":    false,
				"dataset_query": map[string]any{
					"database": 1,
					"type":     "query",
					"query":    map[string]any{"source-card": 20, "filter": []any{"=", []any{"field", 1, nil}, "merchant-a"}},
				},
				"visualization_settings": map[string]any{},
			})
		case "/api/card/20":
			json.NewEncoder(w).Encode(map[string]any{
				"id":          20,
				"name":        "Merchant Base",
				"database_id": 1,
				"display":     "table",
				"query_type":  "native",
				"archived":    false,
				"dataset_query": map[string]any{
					"database": 1,
					"type":     "native",
					"native":   map[string]any{"query": "select * from merchants where plan_id = 42"},
				},
				"visualization_settings": map[string]any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Not found"}`))
		}
	}))
	defer server.Close()

	stdout, stderr, err := runMBCLI(t, map[string]string{
		"MB_HOST":    server.URL,
		"MB_API_KEY": "test-api-key",
	}, "dashboard", "analyze", "1", "-f", "json")
	if err != nil {
		t.Fatalf("dashboard analyze failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to decode dashboard analyze output: %v\noutput: %s", err, stdout)
	}

	if result["dashboard_id"] != float64(1) {
		t.Fatalf("expected dashboard_id 1, got %v", result["dashboard_id"])
	}
	baseCards, ok := result["base_cards"].([]any)
	if !ok || len(baseCards) != 1 {
		t.Fatalf("expected one base card, got %v", result["base_cards"])
	}
	flaggedCards, ok := result["flagged_cards"].([]any)
	if !ok || len(flaggedCards) == 0 {
		t.Fatalf("expected flagged cards in analysis output, got %v", result["flagged_cards"])
	}
	parameters, ok := result["parameters"].([]any)
	if !ok || len(parameters) != 1 {
		t.Fatalf("expected one analyzed parameter, got %v", result["parameters"])
	}
}
