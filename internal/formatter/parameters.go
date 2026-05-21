package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/andreagrandi/mb-cli/internal/client"
)

// FormatCardParametersTable renders the parameters a saved question accepts,
// followed by a ready-to-edit `card run` invocation.
func FormatCardParametersTable(card *client.Card, writer io.Writer) error {
	if card == nil {
		_, err := fmt.Fprintln(writer, "No data")
		return err
	}

	tw := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "card_id\t%d\n", card.ID)
	fmt.Fprintf(tw, "card_name\t%s\n", card.Name)
	fmt.Fprintf(tw, "query_type\t%s\n", card.QueryType)
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Parameters"); err != nil {
		return err
	}

	params := card.CardParameters()
	if len(params) == 0 {
		_, err := fmt.Fprintf(writer, "This card has no parameters; run it with 'mb-cli card run %d'.\n", card.ID)
		return err
	}

	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "name\tdisplay_name\ttype\twidget_type\trequired\tdefault"); err != nil {
		return err
	}
	for _, param := range params {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%t\t%s\n",
			param.Name, param.DisplayName, param.Type, param.WidgetType, param.Required, stringify(param.Default)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	_, err := fmt.Fprintf(writer, "Run with:\n  mb-cli card run %d %s\n", card.ID, cardRunParamHint(params))
	return err
}

func cardRunParamHint(params []client.CardParameter) string {
	hints := make([]string, 0, len(params))
	for _, param := range params {
		hints = append(hints, fmt.Sprintf("--param %s=<value>", param.Name))
	}
	return strings.Join(hints, " ")
}

// FormatDashboardParametersListTable renders a dashboard's parameters, how many
// cards each one filters, and the commands to inspect and run them.
func FormatDashboardParametersListTable(dashboard *client.Dashboard, writer io.Writer) error {
	if dashboard == nil {
		_, err := fmt.Fprintln(writer, "No data")
		return err
	}

	tw := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "dashboard_id\t%d\n", dashboard.ID)
	fmt.Fprintf(tw, "dashboard_name\t%s\n", dashboard.Name)
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Parameters"); err != nil {
		return err
	}

	if len(dashboard.Parameters) == 0 {
		_, err := fmt.Fprintf(writer, "This dashboard has no parameters; run a card with 'mb-cli dashboard run-card %d <dashcard-id> <card-id>'.\n", dashboard.ID)
		return err
	}

	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "id\tname\tslug\ttype\tmapped_cards"); err != nil {
		return err
	}
	for i := range dashboard.Parameters {
		parameter := &dashboard.Parameters[i]
		count := countParameterMappings(dashboard, parameter.ID)
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
			parameter.ID, parameter.Name, parameter.Slug, parameter.Type, count); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Inspect a parameter's valid values with:\n  mb-cli dashboard params values %d <param>\n", dashboard.ID); err != nil {
		return err
	}
	_, err := fmt.Fprintf(writer, "Run a card with:\n  mb-cli dashboard run-card %d <dashcard-id> <card-id> --param <param>=<value>\n", dashboard.ID)
	return err
}

func countParameterMappings(dashboard *client.Dashboard, parameterID string) int {
	count := 0
	for _, dashCard := range dashboard.DashCards {
		for _, mapping := range dashCard.ParameterMappings {
			if mapping.ParameterID == parameterID {
				count++
			}
		}
	}
	return count
}
