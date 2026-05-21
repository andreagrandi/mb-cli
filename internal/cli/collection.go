package cli

import (
	"strings"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var collectionCmd = &cobra.Command{
	Use:   "collection",
	Short: "Collection commands",
}

var collectionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections",
	Args:  cobra.NoArgs,
	RunE:  runCollectionList,
}

var collectionGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get collection details (use 'root' for the root collection)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCollectionGet,
}

var collectionItemsCmd = &cobra.Command{
	Use:   "items <id>",
	Short: "List cards, dashboards, and nested collections in a collection",
	Args:  cobra.ExactArgs(1),
	RunE:  runCollectionItems,
}

func init() {
	rootCmd.AddCommand(collectionCmd)

	collectionCmd.AddCommand(collectionListCmd)
	collectionCmd.AddCommand(collectionGetCmd)
	collectionCmd.AddCommand(collectionItemsCmd)

	collectionItemsCmd.Flags().String("models", "", "Filter by item type (comma-separated: card,dashboard,collection,dataset,snippet)")
}

func runCollectionList(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	collections, err := c.ListCollections(ctx)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, collections)
}

func runCollectionGet(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := requestContext(cmd)
	defer cancel()

	collection, err := c.GetCollection(ctx, normalizeCollectionID(args[0]))
	if err != nil {
		return err
	}

	return formatter.Output(cmd, collection)
}

func runCollectionItems(cmd *cobra.Command, args []string) error {
	c, err := newClient(cmd)
	if err != nil {
		return err
	}

	modelsFlag, _ := cmd.Flags().GetString("models")
	models := client.ParseModels(modelsFlag)

	ctx, cancel := requestContext(cmd)
	defer cancel()

	items, err := c.GetCollectionItems(ctx, normalizeCollectionID(args[0]), models)
	if err != nil {
		return err
	}

	return formatter.Output(cmd, items.Data)
}

// normalizeCollectionID lower-cases the "root" alias so the root collection can
// be requested case-insensitively; numeric IDs are passed through unchanged.
func normalizeCollectionID(id string) string {
	if strings.EqualFold(id, "root") {
		return "root"
	}
	return id
}
