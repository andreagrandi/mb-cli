package client

import (
	"context"
	"net/url"
	"strings"
)

// Search searches across all Metabase items.
func (c *Client) Search(ctx context.Context, query string, models []string) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	for _, m := range models {
		params.Add("models", m)
	}

	resp, err := c.Get(ctx, "/api/search/", params)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Data []SearchResult `json:"data"`
	}
	if err := c.DecodeJSON(resp, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Data, nil
}

// ParseModels splits a comma-separated models string into a slice.
func ParseModels(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
