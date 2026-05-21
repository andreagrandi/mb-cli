package client

import (
	"context"
	"fmt"
	"net/url"
)

// ListDashboards retrieves all dashboards.
func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	resp, err := c.Get(ctx, "/api/dashboard/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	var dashboards []Dashboard
	if err := c.DecodeJSON(resp, &dashboards); err != nil {
		return nil, err
	}

	return dashboards, nil
}

// GetDashboard retrieves a single dashboard by ID.
func (c *Client) GetDashboard(ctx context.Context, id int) (*Dashboard, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/dashboard/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard %d: %w", id, err)
	}

	var dashboard Dashboard
	if err := c.DecodeJSON(resp, &dashboard); err != nil {
		return nil, err
	}

	return &dashboard, nil
}

// GetDashboardParamValues retrieves valid values for a dashboard parameter.
func (c *Client) GetDashboardParamValues(ctx context.Context, dashboardID int, paramKey string) (*ParameterValues, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/dashboard/%d/params/%s/values", dashboardID, url.PathEscape(paramKey)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get values for dashboard %d parameter %s: %w", dashboardID, paramKey, err)
	}

	var values ParameterValues
	if err := c.DecodeJSON(resp, &values); err != nil {
		return nil, err
	}

	return &values, nil
}

// SearchDashboardParamValues searches dashboard parameter values.
func (c *Client) SearchDashboardParamValues(ctx context.Context, dashboardID int, paramKey string, query string) (*ParameterValues, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/dashboard/%d/params/%s/search/%s", dashboardID, url.PathEscape(paramKey), url.PathEscape(query)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search values for dashboard %d parameter %s: %w", dashboardID, paramKey, err)
	}

	var values ParameterValues
	if err := c.DecodeJSON(resp, &values); err != nil {
		return nil, err
	}

	return &values, nil
}

// RunDashboardCard executes a dashboard card with parameter values.
func (c *Client) RunDashboardCard(ctx context.Context, dashboardID, dashcardID, cardID int, params map[string]string) (*QueryResult, error) {
	dashboard, err := c.GetDashboard(ctx, dashboardID)
	if err != nil {
		return nil, err
	}
	card, err := c.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"parameters": buildDashboardQueryParameters(dashboard, params),
	}

	resp, err := c.Post(ctx, fmt.Sprintf("/api/dashboard/%d/dashcard/%d/card/%d/query", dashboardID, dashcardID, cardID), body)
	if err != nil {
		return nil, fmt.Errorf("failed to run dashboard %d card %d via dashcard %d: %w", dashboardID, cardID, dashcardID, err)
	}

	return c.decodeCardQueryResult(ctx, resp, card.DatabaseID)
}
