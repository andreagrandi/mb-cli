package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/andreagrandi/mb-cli/internal/mberr"
)

// ListCards retrieves all saved questions (cards).
func (c *Client) ListCards(ctx context.Context) ([]Card, error) {
	resp, err := c.Get(ctx, "/api/card/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}

	var cards []Card
	if err := c.DecodeJSON(resp, &cards); err != nil {
		return nil, err
	}

	return cards, nil
}

// GetCard retrieves a single card by ID.
func (c *Client) GetCard(ctx context.Context, id int) (*Card, error) {
	params := url.Values{}
	params.Set("legacy-mbql", "true")

	resp, err := c.Get(ctx, fmt.Sprintf("/api/card/%d", id), params)
	if err != nil {
		return nil, &mberr.RequestError{
			Resource: mberr.ResourceCard,
			Op:       fmt.Sprintf("failed to get card %d", id),
			Err:      err,
		}
	}

	var card Card
	if err := c.DecodeJSON(resp, &card); err != nil {
		return nil, err
	}

	return &card, nil
}

// RunCard executes a saved question and returns the query result.
func (c *Client) RunCard(ctx context.Context, id int) (*QueryResult, error) {
	return c.RunCardWithParams(ctx, id, nil)
}

// RunCardWithParams executes a saved question, optionally with parameter values.
// The card is always fetched so semantic types can be enriched for redaction,
// keeping behavior consistent between parameterized and non-parameterized runs.
func (c *Client) RunCardWithParams(ctx context.Context, id int, params map[string]string) (*QueryResult, error) {
	card, err := c.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}

	var body any
	if len(params) > 0 {
		body = map[string]any{
			"parameters": buildCardQueryParameters(card, params),
		}
	}

	resp, err := c.Post(ctx, fmt.Sprintf("/api/card/%d/query", id), body)
	if err != nil {
		return nil, fmt.Errorf("failed to run card %d: %w", id, err)
	}

	return c.decodeCardQueryResult(ctx, resp, card.DatabaseID)
}

func (c *Client) decodeCardQueryResult(ctx context.Context, resp *http.Response, databaseID int) (*QueryResult, error) {
	var result QueryResult
	if err := c.DecodeJSON(resp, &result); err != nil {
		return nil, err
	}

	if c.RedactPII {
		if databaseID > 0 {
			c.EnrichSemanticTypes(ctx, &result, databaseID)
		}
		RedactQueryResult(&result)
	}

	return &result, nil
}
