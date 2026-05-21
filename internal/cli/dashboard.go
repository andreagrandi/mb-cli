package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Dashboard commands",
}

var dashboardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dashboards",
	Args:  cobra.NoArgs,
	RunE:  runDashboardList,
}

var dashboardGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get dashboard details",
	Args:  cobra.ExactArgs(1),
	RunE:  runDashboardGet,
}

var dashboardCardsCmd = &cobra.Command{
	Use:   "cards <id>",
	Short: "List cards in a dashboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runDashboardCards,
}

var dashboardRunCardCmd = &cobra.Command{
	Use:   "run-card <dashboard-id> <dashcard-id> <card-id>",
	Short: "Execute a dashboard card with parameters",
	Args:  cobra.ExactArgs(3),
	RunE:  runDashboardRunCard,
}

var dashboardParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "Dashboard parameter commands",
}

var dashboardParamsListCmd = &cobra.Command{
	Use:   "list <dashboard-id>",
	Short: "List a dashboard's parameters and the cards they filter",
	Args:  cobra.ExactArgs(1),
	RunE:  runDashboardParamsList,
}

var dashboardParamsValuesCmd = &cobra.Command{
	Use:   "values <dashboard-id> <param-key>",
	Short: "List valid values for a dashboard parameter",
	Args:  cobra.ExactArgs(2),
	RunE:  runDashboardParamValues,
}

var dashboardParamsSearchCmd = &cobra.Command{
	Use:   "search <dashboard-id> <param-key> <query>",
	Short: "Search valid values for a dashboard parameter",
	Args:  cobra.ExactArgs(3),
	RunE:  runDashboardParamSearch,
}

type dashboardListRow struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Archived    bool   `json:"archived"`
}

type dashboardCardRow struct {
	DashcardID int    `json:"dashcard_id"`
	Tab        string `json:"tab"`
	CardID     string `json:"card_id,omitempty"`
	Name       string `json:"name,omitempty"`
	QueryType  string `json:"query_type,omitempty"`
	Display    string `json:"display,omitempty"`
}

type dashboardParamLookupResult struct {
	DashboardID   int                        `json:"dashboard_id"`
	RequestedKey  string                     `json:"requested_key"`
	ResolvedKey   string                     `json:"resolved_key"`
	Parameter     *client.DashParameter      `json:"parameter,omitempty"`
	MappedCards   []dashboardParamMappingRow `json:"mapped_cards,omitempty"`
	HasMoreValues bool                       `json:"has_more_values"`
	Values        []client.ParameterValue    `json:"values"`
	Query         string                     `json:"query,omitempty"`
}

type dashboardParamMappingRow struct {
	DashcardID int    `json:"dashcard_id"`
	CardID     string `json:"card_id,omitempty"`
	CardName   string `json:"card_name,omitempty"`
	Target     string `json:"target,omitempty"`
}

type dashboardParamsListResult struct {
	DashboardID   int                     `json:"dashboard_id"`
	DashboardName string                  `json:"dashboard_name"`
	Parameters    []dashboardParamListRow `json:"parameters"`
}

type dashboardParamListRow struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Slug        string                     `json:"slug"`
	Type        string                     `json:"type"`
	MappedCards []dashboardParamMappingRow `json:"mapped_cards"`
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.AddCommand(dashboardListCmd)
	dashboardCmd.AddCommand(dashboardGetCmd)
	dashboardCmd.AddCommand(dashboardCardsCmd)
	dashboardCmd.AddCommand(dashboardRunCardCmd)
	dashboardCmd.AddCommand(dashboardParamsCmd)

	dashboardParamsCmd.AddCommand(dashboardParamsListCmd)
	dashboardParamsCmd.AddCommand(dashboardParamsValuesCmd)
	dashboardParamsCmd.AddCommand(dashboardParamsSearchCmd)

	dashboardRunCardCmd.Flags().String("fields", "", "Comma-separated list of columns to include in output")
	dashboardRunCardCmd.Flags().StringSlice("param", nil, "Parameter in key=value format (repeatable)")
}

func runDashboardList(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	dashboards, err := c.ListDashboards(ctx)
	if err != nil {
		return err
	}

	rows := make([]dashboardListRow, 0, len(dashboards))
	for _, dashboard := range dashboards {
		rows = append(rows, dashboardListRow{
			ID:          dashboard.ID,
			Name:        dashboard.Name,
			Description: dashboard.Description,
			Archived:    dashboard.Archived,
		})
	}

	return formatter.Output(cmd, rows)
}

func runDashboardGet(cmd *cobra.Command, args []string) error {
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

	dashboard, err := c.GetDashboard(ctx, id)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return formatter.Output(cmd, dashboard)
	}

	return formatter.FormatDashboardTable(dashboard, os.Stdout)
}

func runDashboardCards(cmd *cobra.Command, args []string) error {
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

	dashboard, err := c.GetDashboard(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, buildDashboardCardRows(dashboard))
}

func buildDashboardCardRows(dashboard *client.Dashboard) []dashboardCardRow {
	if dashboard == nil {
		return nil
	}

	tabNames := make(map[int]string, len(dashboard.Tabs))
	for _, tab := range dashboard.Tabs {
		tabNames[tab.ID] = tab.Name
	}

	rows := make([]dashboardCardRow, 0, len(dashboard.DashCards))
	for _, dashCard := range dashboard.DashCards {
		row := dashboardCardRow{
			DashcardID: dashCard.ID,
			Tab:        "Ungrouped",
		}
		if dashCard.TabID != nil {
			if name, ok := tabNames[*dashCard.TabID]; ok {
				row.Tab = name
			}
		}
		if dashCard.CardID != nil {
			row.CardID = strconv.Itoa(*dashCard.CardID)
		}
		if dashCard.Card != nil {
			row.Name = dashCard.Card.Name
			row.QueryType = dashCard.Card.QueryType
			row.Display = dashCard.Card.Display
		}
		rows = append(rows, row)
	}

	return rows
}

func runDashboardParamsList(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	dashboard, err := c.GetDashboard(ctx, dashboardID)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		rows := make([]dashboardParamListRow, 0, len(dashboard.Parameters))
		for i := range dashboard.Parameters {
			parameter := &dashboard.Parameters[i]
			rows = append(rows, dashboardParamListRow{
				ID:          parameter.ID,
				Name:        parameter.Name,
				Slug:        parameter.Slug,
				Type:        parameter.Type,
				MappedCards: buildDashboardParamMappingRows(dashboard, parameter),
			})
		}
		return formatter.Output(cmd, dashboardParamsListResult{
			DashboardID:   dashboard.ID,
			DashboardName: dashboard.Name,
			Parameters:    rows,
		})
	}

	return formatter.FormatDashboardParametersListTable(dashboard, os.Stdout)
}

func runDashboardParamValues(cmd *cobra.Command, args []string) error {
	return runDashboardParamLookup(cmd, args[0], args[1], "", false)
}

func runDashboardParamSearch(cmd *cobra.Command, args []string) error {
	return runDashboardParamLookup(cmd, args[0], args[1], args[2], true)
}

func runDashboardRunCard(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	dashcardID, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	cardID, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	params, err := parseNamedParams(cmd)
	if err != nil {
		return err
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	result, err := c.RunDashboardCard(ctx, dashboardID, dashcardID, cardID, params)
	if err != nil {
		return wrapParameterizedRunError(err, len(params) > 0, fmt.Sprintf("mb-cli dashboard params list %d", dashboardID))
	}

	return formatQueryResultOutput(cmd, result)
}

func runDashboardParamLookup(cmd *cobra.Command, dashboardArg string, requestedKey string, query string, search bool) error {
	dashboardID, err := strconv.Atoi(dashboardArg)
	if err != nil {
		return err
	}

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	dashboard, err := c.GetDashboard(ctx, dashboardID)
	if err != nil {
		return err
	}

	parameter, resolvedKey := resolveDashboardParameter(dashboard, requestedKey)

	var values *client.ParameterValues
	if search {
		values, err = c.SearchDashboardParamValues(ctx, dashboardID, resolvedKey, query)
	} else {
		values, err = c.GetDashboardParamValues(ctx, dashboardID, resolvedKey)
	}
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return formatter.Output(cmd, dashboardParamLookupResult{
			DashboardID:   dashboardID,
			RequestedKey:  requestedKey,
			ResolvedKey:   resolvedKey,
			Parameter:     parameter,
			MappedCards:   buildDashboardParamMappingRows(dashboard, parameter),
			HasMoreValues: values.HasMoreValues,
			Values:        values.Values,
			Query:         query,
		})
	}

	return formatter.FormatDashboardParameterValuesTable(dashboard, parameter, values, os.Stdout)
}

func resolveDashboardParameter(dashboard *client.Dashboard, input string) (*client.DashParameter, string) {
	for i := range dashboard.Parameters {
		parameter := &dashboard.Parameters[i]
		if parameter.ID == input || parameter.Slug == input || strings.EqualFold(parameter.Name, input) {
			return parameter, parameter.ID
		}
	}

	return nil, input
}

func buildDashboardParamMappingRows(dashboard *client.Dashboard, parameter *client.DashParameter) []dashboardParamMappingRow {
	if dashboard == nil || parameter == nil {
		return nil
	}

	rows := make([]dashboardParamMappingRow, 0)
	for _, dashCard := range dashboard.DashCards {
		for _, mapping := range dashCard.ParameterMappings {
			if mapping.ParameterID != parameter.ID {
				continue
			}

			row := dashboardParamMappingRow{
				DashcardID: dashCard.ID,
				Target:     stringifyDashboardTarget(mapping.Target),
			}
			if dashCard.CardID != nil {
				row.CardID = strconv.Itoa(*dashCard.CardID)
			}
			if dashCard.Card != nil {
				row.CardName = dashCard.Card.Name
			}
			rows = append(rows, row)
		}
	}

	return rows
}

func stringifyDashboardTarget(target []any) string {
	if len(target) == 0 {
		return ""
	}

	data, err := json.Marshal(target)
	if err != nil {
		return fmt.Sprintf("%v", target)
	}

	return string(data)
}
