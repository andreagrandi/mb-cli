package cli

import (
	"strconv"

	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var fieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Field inspection commands",
}

var fieldGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get field details",
	Args:  cobra.ExactArgs(1),
	RunE:  runFieldGet,
}

var fieldSummaryCmd = &cobra.Command{
	Use:   "summary <id>",
	Short: "Summary statistics for a field",
	Args:  cobra.ExactArgs(1),
	RunE:  runFieldSummary,
}

var fieldValuesCmd = &cobra.Command{
	Use:   "values <id>",
	Short: "Distinct values for a field",
	Args:  cobra.ExactArgs(1),
	RunE:  runFieldValues,
}

func init() {
	rootCmd.AddCommand(fieldCmd)

	fieldCmd.AddCommand(fieldGetCmd)
	fieldCmd.AddCommand(fieldSummaryCmd)
	fieldCmd.AddCommand(fieldValuesCmd)
}

func runFieldGet(cmd *cobra.Command, args []string) error {
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

	field, err := c.GetField(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, field)
}

func runFieldSummary(cmd *cobra.Command, args []string) error {
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

	summary, err := c.GetFieldSummary(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, summary)
}

func runFieldValues(cmd *cobra.Command, args []string) error {
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

	values, err := c.GetFieldValues(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, values)
}
