package client

import (
	"context"
	"fmt"
	"io"
)

// RunNativeQuery executes a native SQL query against the specified database.
func (c *Client) RunNativeQuery(ctx context.Context, databaseID int, sql string) (*QueryResult, error) {
	query := DatasetQuery{
		Database: databaseID,
		Type:     "native",
		Native:   &NativeQuery{Query: sql},
	}

	resp, err := c.Post(ctx, "/api/dataset/", query)
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := c.DecodeJSON(resp, &result); err != nil {
		return nil, err
	}

	if c.RedactPII {
		c.EnrichSemanticTypes(ctx, &result, databaseID)
		RedactQueryResult(&result)
	}

	return &result, nil
}

// ExportNativeQuery executes a native SQL query and returns the result in the specified export format.
func (c *Client) ExportNativeQuery(ctx context.Context, databaseID int, sql string, format string) ([]byte, error) {
	if c.RedactPII {
		return nil, fmt.Errorf("export is not supported when PII redaction is enabled (use JSON or table format instead)")
	}

	query := DatasetQuery{
		Database: databaseID,
		Type:     "native",
		Native:   &NativeQuery{Query: sql},
	}

	return c.exportQuery(ctx, query, format)
}

// RunStructuredQuery executes an MBQL structured query with filters.
func (c *Client) RunStructuredQuery(ctx context.Context, databaseID, tableID int, filters [][]any, limit int) (*QueryResult, error) {
	sq := &StructuredQuery{
		SourceTable: tableID,
	}

	if len(filters) == 1 {
		sq.Filter = filters[0]
	} else if len(filters) > 1 {
		filter := []any{"and"}
		for _, f := range filters {
			filter = append(filter, f)
		}
		sq.Filter = filter
	}

	if limit > 0 {
		sq.Limit = limit
	}

	query := DatasetQuery{
		Database: databaseID,
		Type:     "query",
		Query:    sq,
	}

	resp, err := c.Post(ctx, "/api/dataset/", query)
	if err != nil {
		return nil, err
	}

	var result QueryResult
	if err := c.DecodeJSON(resp, &result); err != nil {
		return nil, err
	}

	if c.RedactPII {
		RedactQueryResult(&result)
	}

	return &result, nil
}

// ExportStructuredQuery executes an MBQL structured query and returns the result in the specified export format.
func (c *Client) ExportStructuredQuery(ctx context.Context, databaseID, tableID int, filters [][]any, limit int, format string) ([]byte, error) {
	if c.RedactPII {
		return nil, fmt.Errorf("export is not supported when PII redaction is enabled (use JSON or table format instead)")
	}

	sq := &StructuredQuery{
		SourceTable: tableID,
	}

	if len(filters) == 1 {
		sq.Filter = filters[0]
	} else if len(filters) > 1 {
		filter := []any{"and"}
		for _, f := range filters {
			filter = append(filter, f)
		}
		sq.Filter = filter
	}

	if limit > 0 {
		sq.Limit = limit
	}

	query := DatasetQuery{
		Database: databaseID,
		Type:     "query",
		Query:    sq,
	}

	return c.exportQuery(ctx, query, format)
}

func (c *Client) exportQuery(ctx context.Context, query DatasetQuery, format string) ([]byte, error) {
	body := map[string]any{
		"query": query,
	}

	resp, err := c.Post(ctx, fmt.Sprintf("/api/dataset/%s", format), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read export response: %w", err)
	}

	return data, nil
}
