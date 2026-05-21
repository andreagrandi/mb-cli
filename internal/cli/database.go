package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/config"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var databaseCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Database exploration commands",
}

var databaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all databases",
	Args:  cobra.NoArgs,
	RunE:  runDatabaseList,
}

var databaseGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get database details",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabaseGet,
}

var databaseMetadataCmd = &cobra.Command{
	Use:   "metadata <id>",
	Short: "Full metadata (tables + fields)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabaseMetadata,
}

var databaseFieldsCmd = &cobra.Command{
	Use:   "fields <id>",
	Short: "List all fields in database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabaseFields,
}

var databaseSchemasCmd = &cobra.Command{
	Use:   "schemas <id>",
	Short: "List schema names",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabaseSchemas,
}

var databaseSchemaCmd = &cobra.Command{
	Use:   "schema <id> <schema>",
	Short: "Tables in a specific schema",
	Args:  cobra.ExactArgs(2),
	RunE:  runDatabaseSchema,
}

func init() {
	rootCmd.AddCommand(databaseCmd)

	databaseCmd.AddCommand(databaseListCmd)
	databaseCmd.AddCommand(databaseGetCmd)
	databaseCmd.AddCommand(databaseMetadataCmd)
	databaseCmd.AddCommand(databaseFieldsCmd)
	databaseCmd.AddCommand(databaseSchemasCmd)
	databaseCmd.AddCommand(databaseSchemaCmd)
}

func newClient(cmd *cobra.Command) (*client.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	c := client.NewClient(cfg)
	verbose, _ := cmd.Flags().GetBool("verbose")
	c.Verbose = verbose

	redactPII := true
	if cmd.Flags().Changed("redact-pii") {
		redactPII, _ = cmd.Flags().GetBool("redact-pii")
	} else if v, ok := os.LookupEnv("MB_REDACT_PII"); ok {
		redactPII = v != "false"
	}
	c.RedactPII = redactPII

	if !redactPII {
		fmt.Fprintln(os.Stderr, "Warning: PII redaction is disabled")
	}

	return c, nil
}

func runDatabaseList(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	databases, err := c.ListDatabases(ctx, false)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, databases)
}

func runDatabaseGet(cmd *cobra.Command, args []string) error {
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

	db, err := c.GetDatabase(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, db)
}

func runDatabaseMetadata(cmd *cobra.Command, args []string) error {
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

	meta, err := c.GetDatabaseMetadata(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, meta)
}

func runDatabaseFields(cmd *cobra.Command, args []string) error {
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

	fields, err := c.GetDatabaseFields(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, fields)
}

func runDatabaseSchemas(cmd *cobra.Command, args []string) error {
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

	schemas, err := c.ListDatabaseSchemas(ctx, id)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, schemas)
}

func runDatabaseSchema(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	schema := args[1]

	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	tables, err := c.GetDatabaseSchema(ctx, id, schema)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, tables)
}
