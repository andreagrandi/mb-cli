package cli

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/andreagrandi/mb-cli/internal/mberr"
	"github.com/spf13/cobra"
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Saved question commands",
}

var cardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved questions",
	Args:  cobra.NoArgs,
	RunE:  runCardList,
}

var cardGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get card details",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardGet,
}

var cardRunCmd = &cobra.Command{
	Use:   "run <id>",
	Short: "Execute a saved question",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardRun,
}

type cardSummary struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	DatabaseID   int    `json:"database_id"`
	Display      string `json:"display"`
	QueryType    string `json:"query_type,omitempty"`
	CollectionID *int   `json:"collection_id,omitempty"`
	Archived     bool   `json:"archived"`
}

func init() {
	rootCmd.AddCommand(cardCmd)

	cardCmd.AddCommand(cardListCmd)
	cardCmd.AddCommand(cardGetCmd)
	cardCmd.AddCommand(cardRunCmd)

	cardGetCmd.Flags().Bool("full", false, "Include the full query definition and card metadata")
	cardRunCmd.Flags().String("fields", "", "Comma-separated list of columns to include in output")
	cardRunCmd.Flags().StringSlice("param", nil, "Parameter in key=value format (repeatable)")
}

func runCardList(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	cards, err := c.ListCards(ctx)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, cards)
}

func runCardGet(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	card, err := c.GetCard(ctx, id)
	if err != nil {
		return err
	}

	full, _ := cmd.Flags().GetBool("full")
	if full {
		return formatter.Output(cmd, card)
	}

	return formatter.Output(cmd, summarizeCard(card))
}

func runCardRun(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	params, err := parseNamedParams(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	result, err := c.RunCardWithParams(ctx, id, params)
	if err != nil {
		return wrapParameterizedRunError(err)
	}

	return formatQueryResultOutput(cmd, result)
}

func summarizeCard(card *client.Card) cardSummary {
	return cardSummary{
		ID:           card.ID,
		Name:         card.Name,
		Description:  card.Description,
		DatabaseID:   card.DatabaseID,
		Display:      card.Display,
		QueryType:    card.QueryType,
		CollectionID: card.CollectionID,
		Archived:     card.Archived,
	}
}

func parseNamedParams(cmd *cobra.Command) (map[string]string, error) {
	values, err := cmd.Flags().GetStringSlice("param")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}

	params := make(map[string]string, len(values))
	for _, value := range values {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf("invalid parameter %q: expected key=value", value)
		}
		params[strings.TrimSpace(parts[0])] = parts[1]
	}

	return params, nil
}

func formatQueryResultOutput(cmd *cobra.Command, result *client.QueryResult) error {
	format, _ := cmd.Flags().GetString("format")
	fields, _ := cmd.Flags().GetString("fields")

	columns := make([]string, len(result.Data.Columns))
	for i, col := range result.Data.Columns {
		columns[i] = col.Name
	}

	columns, rows := formatter.FilterColumns(columns, result.Data.Rows, fields)
	return formatter.FormatQueryResults(format, columns, rows, os.Stdout)
}

func wrapParameterizedRunError(err error) error {
	var apiErr *mberr.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusBadRequest:
			return &mberr.ParameterizedQueryError{Err: err}
		case http.StatusNotFound:
			return fmt.Errorf("query target was not found (%w)", err)
		}
	}
	return err
}
