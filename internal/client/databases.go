package client

import (
	"context"
	"fmt"
	"net/url"
)

// ListDatabases retrieves all databases. If includeTables is true, tables are included.
func (c *Client) ListDatabases(ctx context.Context, includeTables bool) ([]Database, error) {
	var params url.Values
	if includeTables {
		params = url.Values{}
		params.Set("include", "tables")
	}

	resp, err := c.Get(ctx, "/api/database/", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Database `json:"data"`
	}
	if err := c.DecodeJSON(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// GetDatabase retrieves a single database by ID.
func (c *Client) GetDatabase(ctx context.Context, id int) (*Database, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/database/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var db Database
	if err := c.DecodeJSON(resp, &db); err != nil {
		return nil, err
	}

	return &db, nil
}

// GetDatabaseMetadata retrieves full metadata (tables + fields) for a database.
func (c *Client) GetDatabaseMetadata(ctx context.Context, id int) (*DatabaseMetadata, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/database/%d/metadata", id), nil)
	if err != nil {
		return nil, err
	}

	var meta DatabaseMetadata
	if err := c.DecodeJSON(resp, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// GetDatabaseFields retrieves all fields in a database.
func (c *Client) GetDatabaseFields(ctx context.Context, id int) ([]Field, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/database/%d/fields", id), nil)
	if err != nil {
		return nil, err
	}

	var fields []Field
	if err := c.DecodeJSON(resp, &fields); err != nil {
		return nil, err
	}

	return fields, nil
}

// ListDatabaseSchemas retrieves schema names for a database.
func (c *Client) ListDatabaseSchemas(ctx context.Context, id int) ([]string, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/database/%d/schemas", id), nil)
	if err != nil {
		return nil, err
	}

	var schemas []string
	if err := c.DecodeJSON(resp, &schemas); err != nil {
		return nil, err
	}

	return schemas, nil
}

// GetDatabaseSchema retrieves tables in a specific schema of a database.
func (c *Client) GetDatabaseSchema(ctx context.Context, id int, schema string) ([]Table, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/database/%d/schema/%s", id, url.PathEscape(schema)), nil)
	if err != nil {
		return nil, err
	}

	var tables []Table
	if err := c.DecodeJSON(resp, &tables); err != nil {
		return nil, err
	}

	return tables, nil
}
