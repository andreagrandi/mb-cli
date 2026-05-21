package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/andreagrandi/mb-cli/internal/mberr"
)

// ListCollections retrieves all collections.
func (c *Client) ListCollections(ctx context.Context) ([]Collection, error) {
	resp, err := c.Get(ctx, "/api/collection/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var collections []Collection
	if err := c.DecodeJSON(resp, &collections); err != nil {
		return nil, err
	}

	return collections, nil
}

// GetCollection retrieves a single collection. The ID may be a numeric
// collection ID or the literal "root" for the root collection.
func (c *Client) GetCollection(ctx context.Context, id string) (*Collection, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/collection/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, &mberr.RequestError{
			Resource: mberr.ResourceCollection,
			Op:       fmt.Sprintf("failed to get collection %s", id),
			Err:      err,
		}
	}

	var collection Collection
	if err := c.DecodeJSON(resp, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

// GetCollectionItems retrieves the items contained in a collection, optionally
// filtered to specific models such as card or dashboard.
func (c *Client) GetCollectionItems(ctx context.Context, id string, models []string) (*CollectionItems, error) {
	var params url.Values
	if len(models) > 0 {
		params = url.Values{}
		for _, m := range models {
			params.Add("models", m)
		}
	}

	resp, err := c.Get(ctx, fmt.Sprintf("/api/collection/%s/items", url.PathEscape(id)), params)
	if err != nil {
		return nil, &mberr.RequestError{
			Resource: mberr.ResourceCollection,
			Op:       fmt.Sprintf("failed to get items for collection %s", id),
			Err:      err,
		}
	}

	var items CollectionItems
	if err := c.DecodeJSON(resp, &items); err != nil {
		return nil, err
	}

	return &items, nil
}
