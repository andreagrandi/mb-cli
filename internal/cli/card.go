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

var cardParamsCmd = &cobra.Command{
	Use:   "params <id>",
	Short: "List the parameters a saved question accepts",
	Long:  "Lists the parameters a saved question accepts so they can be supplied to 'card run --param key=value'.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCardParams,
}

type cardParamsResult struct {
	CardID     int                    `json:"card_id"`
	CardName   string                 `json:"card_name"`
	QueryType  string                 `json:"query_type,omitempty"`
	Parameters []client.CardParameter `json:"parameters"`
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
	cardCmd.AddCommand(cardParamsCmd)

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
		return wrapParameterizedRunError(err, len(params) > 0, fmt.Sprintf("mb-cli card params %d", id))
	}

	return formatQueryResultOutput(cmd, result)
}

func runCardParams(cmd *cobra.Command, args []string) error {
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

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return formatter.Output(cmd, cardParamsResult{
			CardID:     card.ID,
			CardName:   card.Name,
			QueryType:  card.QueryType,
			Parameters: card.CardParameters(),
		})
	}

	return formatter.FormatCardParametersTable(card, os.Stdout)
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

// wrapParameterizedRunError classifies a failed card or dashboard run. When the
// caller supplied parameters, an API rejection is surfaced as a
// ParameterizedQueryError carrying inspect, the exact command that lists the
// valid parameters for the query. Metabase rejects invalid parameters with a
// 500, so the supplied-parameters context, not the status code, drives this.
func wrapParameterizedRunError(err error, hadParams bool, inspect string) error {
	var apiErr *mberr.APIError
	if !errors.As(err, &apiErr) {
		return err
	}

	if apiErr.StatusCode == http.StatusNotFound {
		return fmt.Errorf("query target was not found (%w)", err)
	}
	if hadParams {
		return &mberr.ParameterizedQueryError{Err: err, Inspect: inspect}
	}
	return err
}
