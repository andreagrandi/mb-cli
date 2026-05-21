package cli

import (
	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/andreagrandi/mb-cli/internal/validation"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across Metabase items",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().String("models", "", "Filter by type (comma-separated: table,card,database,dashboard,collection,metric)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	if err := validation.ValidateSearchQuery(args[0]); err != nil {
		return err
	}

	modelsFlag, _ := cmd.Flags().GetString("models")
	models := client.ParseModels(modelsFlag)

	ctx, cancel := requestContext(cmd)
	defer cancel()

	results, err := c.Search(ctx, args[0], models)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, results)
}
