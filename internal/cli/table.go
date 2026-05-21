package cli

import (
	"os"
	"strconv"

	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Table exploration commands",
}

var tableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tables",
	Args:  cobra.NoArgs,
	RunE:  runTableList,
}

var tableGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get table details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTableGet,
}

var tableMetadataCmd = &cobra.Command{
	Use:   "metadata <id>",
	Short: "Table metadata with fields",
	Args:  cobra.ExactArgs(1),
	RunE:  runTableMetadata,
}

var tableFKsCmd = &cobra.Command{
	Use:   "fks <id>",
	Short: "Foreign key relationships",
	Args:  cobra.ExactArgs(1),
	RunE:  runTableFKs,
}

var tableDataCmd = &cobra.Command{
	Use:   "data <id>",
	Short: "Get raw table data",
	Args:  cobra.ExactArgs(1),
	RunE:  runTableData,
}

func init() {
	rootCmd.AddCommand(tableCmd)

	tableCmd.AddCommand(tableListCmd)
	tableCmd.AddCommand(tableGetCmd)
	tableCmd.AddCommand(tableMetadataCmd)
	tableCmd.AddCommand(tableFKsCmd)
	tableCmd.AddCommand(tableDataCmd)
}

func runTableList(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	tables, err := c.ListTables(ctx)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, tables)
}

func runTableGet(cmd *cobra.Command, args []string) error {
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

	table, err := c.GetTable(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, table)
}

func runTableMetadata(cmd *cobra.Command, args []string) error {
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

	meta, err := c.GetTableMetadata(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, meta)
}

func runTableFKs(cmd *cobra.Command, args []string) error {
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

	fks, err := c.GetTableFKs(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, fks)
}

func runTableData(cmd *cobra.Command, args []string) error {
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

	result, err := c.GetTableData(ctx, id)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")

	columns := make([]string, len(result.Data.Columns))
	for i, col := range result.Data.Columns {
		columns[i] = col.Name
	}

	return formatter.FormatQueryResults(format, columns, result.Data.Rows, os.Stdout)
}
