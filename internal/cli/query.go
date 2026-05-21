package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/andreagrandi/mb-cli/internal/mberr"
	"github.com/andreagrandi/mb-cli/internal/validation"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query commands",
}

var querySQLCmd = &cobra.Command{
	Use:   "sql",
	Short: "Run a native SQL query",
	Args:  cobra.NoArgs,
	RunE:  runQuerySQL,
}

var queryFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Run a structured query with field filters",
	Args:  cobra.NoArgs,
	RunE:  runQueryFilter,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.AddCommand(querySQLCmd)
	queryCmd.AddCommand(queryFilterCmd)

	querySQLCmd.Flags().String("db", "", "Database ID or name substring (required)")
	querySQLCmd.Flags().String("sql", "", "SQL query to execute (required)")
	querySQLCmd.Flags().String("export", "", "Export format: csv, json, xlsx")
	querySQLCmd.Flags().Int("limit", 0, "Append LIMIT to SQL query")
	querySQLCmd.Flags().String("fields", "", "Comma-separated list of columns to include in output")
	querySQLCmd.MarkFlagRequired("db")
	querySQLCmd.MarkFlagRequired("sql")

	queryFilterCmd.Flags().String("db", "", "Database ID or name substring (required)")
	queryFilterCmd.Flags().String("table", "", "Table ID or name substring (required)")
	queryFilterCmd.Flags().StringSlice("where", nil, "Filter in field=value format (repeatable)")
	queryFilterCmd.Flags().Int("limit", 0, "Maximum number of rows to return")
	queryFilterCmd.Flags().String("export", "", "Export format: csv, json, xlsx")
	queryFilterCmd.Flags().String("fields", "", "Comma-separated list of columns to include in output")
	queryFilterCmd.MarkFlagRequired("db")
	queryFilterCmd.MarkFlagRequired("table")
	queryFilterCmd.MarkFlagRequired("where")
}

func runQuerySQL(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	dbFlag, _ := cmd.Flags().GetString("db")
	sql, _ := cmd.Flags().GetString("sql")
	export, _ := cmd.Flags().GetString("export")
	limit, _ := cmd.Flags().GetInt("limit")

	if err := validation.ValidateSQL(sql); err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	dbID, err := resolveDatabaseID(ctx, c, dbFlag)
	if err != nil {
		return err
	}

	if limit > 0 {
		sql = fmt.Sprintf("%s LIMIT %d", sql, limit)
	}

	if export != "" {
		data, err := c.ExportNativeQuery(ctx, dbID, sql, export)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	}

	result, err := c.RunNativeQuery(ctx, dbID, sql)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	fields, _ := cmd.Flags().GetString("fields")

	columns := make([]string, len(result.Data.Columns))
	for i, col := range result.Data.Columns {
		columns[i] = col.Name
	}

	columns, rows := formatter.FilterColumns(columns, result.Data.Rows, fields)
	return formatter.FormatQueryResults(format, columns, rows, os.Stdout)
}

func resolveDatabaseID(ctx context.Context, c *client.Client, dbFlag string) (int, error) {
	if id, err := strconv.Atoi(dbFlag); err == nil {
		return id, nil
	}

	databases, err := c.ListDatabases(ctx, false)
	if err != nil {
		return 0, fmt.Errorf("failed to list databases for name resolution: %w", err)
	}

	return matchDatabaseByName(databases, dbFlag)
}

// MatchDatabaseByName finds a database by case-insensitive substring match.
// Exported for testing.
func MatchDatabaseByName(databases []client.Database, name string) (int, error) {
	return matchDatabaseByName(databases, name)
}

func runQueryFilter(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	dbFlag, _ := cmd.Flags().GetString("db")
	tableFlag, _ := cmd.Flags().GetString("table")
	whereClauses, _ := cmd.Flags().GetStringSlice("where")
	limit, _ := cmd.Flags().GetInt("limit")
	export, _ := cmd.Flags().GetString("export")

	for _, w := range whereClauses {
		if err := validation.ValidateNoControlChars(w, "where clause"); err != nil {
			return err
		}
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	dbID, err := resolveDatabaseID(ctx, c, dbFlag)
	if err != nil {
		return err
	}

	tableID, err := resolveTableID(ctx, c, dbID, tableFlag)
	if err != nil {
		return err
	}

	tableMeta, err := c.GetTableMetadata(ctx, tableID)
	if err != nil {
		return fmt.Errorf("failed to get table metadata: %w", err)
	}

	var filters [][]any
	for _, w := range whereClauses {
		fieldName, value, err := ParseWhereClause(w)
		if err != nil {
			return err
		}
		fieldID, err := resolveFieldID(tableMeta.Fields, fieldName)
		if err != nil {
			return err
		}
		filters = append(filters, []any{"=", []any{"field", fieldID, nil}, value})
	}

	if export != "" {
		data, err := c.ExportStructuredQuery(ctx, dbID, tableID, filters, limit, export)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	}

	result, err := c.RunStructuredQuery(ctx, dbID, tableID, filters, limit)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	fields, _ := cmd.Flags().GetString("fields")

	columns := make([]string, len(result.Data.Columns))
	for i, col := range result.Data.Columns {
		columns[i] = col.Name
	}

	columns, rows := formatter.FilterColumns(columns, result.Data.Rows, fields)
	return formatter.FormatQueryResults(format, columns, rows, os.Stdout)
}

func resolveTableID(ctx context.Context, c *client.Client, dbID int, tableFlag string) (int, error) {
	if id, err := strconv.Atoi(tableFlag); err == nil {
		return id, nil
	}

	meta, err := c.GetDatabaseMetadata(ctx, dbID)
	if err != nil {
		return 0, fmt.Errorf("failed to get database metadata for table resolution: %w", err)
	}

	return matchTableByName(meta.Tables, tableFlag)
}

// ResolveTableID is exported for testing.
func ResolveTableID(ctx context.Context, c *client.Client, dbID int, tableFlag string) (int, error) {
	return resolveTableID(ctx, c, dbID, tableFlag)
}

// MatchTableByName finds a table by case-insensitive substring match. Exported for testing.
func MatchTableByName(tables []client.TableMetadata, name string) (int, error) {
	return matchTableByName(tables, name)
}

func matchTableByName(tables []client.TableMetadata, name string) (int, error) {
	var matches []client.TableMetadata
	search := strings.ToLower(name)
	for _, t := range tables {
		if strings.Contains(strings.ToLower(t.Name), search) {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		return 0, &mberr.ResolutionError{
			Kind:    mberr.ResourceTable,
			Message: fmt.Sprintf("no table matching '%s' found", name),
		}
	case 1:
		return matches[0].ID, nil
	default:
		names := make([]string, len(matches))
		for i, t := range matches {
			names[i] = fmt.Sprintf("%s (id=%d)", t.Name, t.ID)
		}
		return 0, &mberr.ResolutionError{
			Kind:    mberr.ResourceTable,
			Message: fmt.Sprintf("ambiguous table name '%s', matches: %s. Use table ID instead", name, strings.Join(names, ", ")),
		}
	}
}

// ParseWhereClause parses a "field=value" string into field name and value. Exported for testing.
func ParseWhereClause(clause string) (string, string, error) {
	idx := strings.Index(clause, "=")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid where clause '%s': expected format field=value", clause)
	}
	field := strings.TrimSpace(clause[:idx])
	value := strings.TrimSpace(clause[idx+1:])
	if field == "" {
		return "", "", fmt.Errorf("invalid where clause '%s': field name is empty", clause)
	}
	return field, value, nil
}

func resolveFieldID(fields []client.Field, name string) (int, error) {
	search := strings.ToLower(name)
	for _, f := range fields {
		if strings.ToLower(f.Name) == search || strings.ToLower(f.DisplayName) == search {
			return f.ID, nil
		}
	}
	return 0, &mberr.ResolutionError{
		Kind:    mberr.ResourceField,
		Message: fmt.Sprintf("no field matching '%s' found in table", name),
	}
}

func matchDatabaseByName(databases []client.Database, name string) (int, error) {
	var matches []client.Database
	search := strings.ToLower(name)
	for _, db := range databases {
		if strings.Contains(strings.ToLower(db.Name), search) {
			matches = append(matches, db)
		}
	}

	switch len(matches) {
	case 0:
		return 0, &mberr.ResolutionError{
			Kind:    mberr.ResourceDatabase,
			Message: fmt.Sprintf("no database matching '%s' found", name),
		}
	case 1:
		return matches[0].ID, nil
	default:
		names := make([]string, len(matches))
		for i, db := range matches {
			names[i] = fmt.Sprintf("%s (id=%d)", db.Name, db.ID)
		}
		return 0, &mberr.ResolutionError{
			Kind:    mberr.ResourceDatabase,
			Message: fmt.Sprintf("ambiguous database name '%s', matches: %s. Use database ID instead", name, strings.Join(names, ", ")),
		}
	}
}
