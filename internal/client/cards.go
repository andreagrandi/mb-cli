package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// ListCards retrieves all saved questions (cards).
func (c *Client) ListCards() ([]Card, error) {
	resp, err := c.Get("/api/card/", nil)
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
func (c *Client) GetCard(id int) (*Card, error) {
	params := url.Values{}
	params.Set("legacy-mbql", "true")

	resp, err := c.Get(fmt.Sprintf("/api/card/%d", id), params)
	if err != nil {
		return nil, fmt.Errorf("failed to get card %d: %w", id, err)
	}

	var card Card
	if err := c.DecodeJSON(resp, &card); err != nil {
		return nil, err
	}

	return &card, nil
}

// RunCard executes a saved question and returns the query result.
func (c *Client) RunCard(id int) (*QueryResult, error) {
	return c.RunCardWithParams(id, nil)
}

// RunCardWithParams executes a saved question, optionally with parameter values.
// The card is always fetched so semantic types can be enriched for redaction,
// keeping behavior consistent between parameterized and non-parameterized runs.
func (c *Client) RunCardWithParams(id int, params map[string]string) (*QueryResult, error) {
	card, err := c.GetCard(id)
	if err != nil {
		return nil, err
	}

	var body any
	if len(params) > 0 {
		body = map[string]any{
			"parameters": buildCardQueryParameters(card, params),
		}
	}

	resp, err := c.Post(fmt.Sprintf("/api/card/%d/query", id), body)
	if err != nil {
		return nil, fmt.Errorf("failed to run card %d: %w", id, err)
	}

	return c.decodeCardQueryResult(resp, card.DatabaseID)
}

func (c *Client) decodeCardQueryResult(resp *http.Response, databaseID int) (*QueryResult, error) {
	var result QueryResult
	if err := c.DecodeJSON(resp, &result); err != nil {
		return nil, err
	}

	if c.RedactPII {
		if databaseID > 0 {
			c.EnrichSemanticTypes(&result, databaseID)
		}
		RedactQueryResult(&result)
	}

	return &result, nil
}
