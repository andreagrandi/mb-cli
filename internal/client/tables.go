package client

import (
	"context"
	"fmt"
)

// ListTables retrieves all tables.
func (c *Client) ListTables(ctx context.Context) ([]Table, error) {
	resp, err := c.Get(ctx, "/api/table/", nil)
	if err != nil {
		return nil, err
	}

	var tables []Table
	if err := c.DecodeJSON(resp, &tables); err != nil {
		return nil, err
	}

	return tables, nil
}

// GetTable retrieves a single table by ID.
func (c *Client) GetTable(ctx context.Context, id int) (*Table, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/table/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var table Table
	if err := c.DecodeJSON(resp, &table); err != nil {
		return nil, err
	}

	return &table, nil
}

// GetTableMetadata retrieves table metadata with field details.
func (c *Client) GetTableMetadata(ctx context.Context, id int) (*TableMetadata, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/table/%d/query_metadata", id), nil)
	if err != nil {
		return nil, err
	}

	var meta TableMetadata
	if err := c.DecodeJSON(resp, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// GetTableFKs retrieves foreign key relationships for a table.
func (c *Client) GetTableFKs(ctx context.Context, id int) ([]ForeignKey, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/table/%d/fks", id), nil)
	if err != nil {
		return nil, err
	}

	var fks []ForeignKey
	if err := c.DecodeJSON(resp, &fks); err != nil {
		return nil, err
	}

	return fks, nil
}

// GetTableData retrieves raw data for a table.
func (c *Client) GetTableData(ctx context.Context, id int) (*QueryResult, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/table/%d/data", id), nil)
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
