package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/andreagrandi/mb-cli/internal/client"
	"github.com/andreagrandi/mb-cli/internal/formatter"
	"github.com/spf13/cobra"
)

var (
	assumptionQueryPattern = regexp.MustCompile(`(?i)\b(sample|placeholder|example|demo|assum(?:e|ed|ption)|estimate|estimated)\b`)
	hardcodedQueryPattern  = regexp.MustCompile(`(?i)\b(where|and|or)\b[^;\n]*(=|>=|<=|>|<)\s*('[^']+'|\d{2,})`)
)

var dashboardAnalyzeCmd = &cobra.Command{
	Use:   "analyze <id>",
	Short: "Summarize dashboard structure and dependencies",
	Args:  cobra.ExactArgs(1),
	RunE:  runDashboardAnalyze,
}

type dashboardAnalysis struct {
	DashboardID     int                          `json:"dashboard_id"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description,omitempty"`
	Tabs            []dashboardAnalysisTab       `json:"tabs"`
	Dashcards       []dashboardAnalysisDashcard  `json:"dashcards"`
	Parameters      []dashboardAnalysisParameter `json:"parameters"`
	BaseCards       []dashboardAnalysisBaseCard  `json:"base_cards"`
	FlaggedCards    []dashboardAnalysisFlagged   `json:"flagged_cards,omitempty"`
	TotalDashcards  int                          `json:"total_dashcards"`
	TotalParameters int                          `json:"total_parameters"`
}

type dashboardAnalysisTab struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	DashcardIDs []int  `json:"dashcard_ids,omitempty"`
	CardIDs     []int  `json:"card_ids,omitempty"`
}

type dashboardAnalysisDashcard struct {
	DashcardID      int      `json:"dashcard_id"`
	Tab             string   `json:"tab"`
	CardID          *int     `json:"card_id,omitempty"`
	Name            string   `json:"name,omitempty"`
	QueryType       string   `json:"query_type,omitempty"`
	Display         string   `json:"display,omitempty"`
	ParameterIDs    []string `json:"parameter_ids,omitempty"`
	SourceCardID    *int     `json:"source_card_id,omitempty"`
	SourceCardChain []int    `json:"source_card_chain,omitempty"`
	BaseCardID      *int     `json:"base_card_id,omitempty"`
	Flags           []string `json:"flags,omitempty"`
}

type dashboardAnalysisParameter struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Slug        string                     `json:"slug"`
	Type        string                     `json:"type"`
	MappedCards []dashboardParamMappingRow `json:"mapped_cards,omitempty"`
}

type dashboardAnalysisBaseCard struct {
	CardID                int      `json:"card_id"`
	Name                  string   `json:"name,omitempty"`
	QueryType             string   `json:"query_type,omitempty"`
	ReferencedByDashcards []int    `json:"referenced_by_dashcards,omitempty"`
	Flags                 []string `json:"flags,omitempty"`
}

type dashboardAnalysisFlagged struct {
	CardID int      `json:"card_id"`
	Name   string   `json:"name,omitempty"`
	Flags  []string `json:"flags"`
}

func init() {
	dashboardCmd.AddCommand(dashboardAnalyzeCmd)
}

func runDashboardAnalyze(cmd *cobra.Command, args []string) error {
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

	analysis, err := analyzeDashboard(ctx, c, id)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return formatter.Output(cmd, analysis)
	}

	return formatDashboardAnalysisTable(os.Stdout, analysis)
}

func analyzeDashboard(ctx context.Context, c *client.Client, id int) (*dashboardAnalysis, error) {
	dashboard, err := c.GetDashboard(ctx, id)
	if err != nil {
		return nil, err
	}

	analysis := &dashboardAnalysis{
		DashboardID:     dashboard.ID,
		Name:            dashboard.Name,
		Description:     dashboard.Description,
		Parameters:      make([]dashboardAnalysisParameter, 0, len(dashboard.Parameters)),
		Dashcards:       make([]dashboardAnalysisDashcard, 0, len(dashboard.DashCards)),
		Tabs:            buildDashboardAnalysisTabs(dashboard),
		TotalDashcards:  len(dashboard.DashCards),
		TotalParameters: len(dashboard.Parameters),
	}

	for _, parameter := range dashboard.Parameters {
		analysis.Parameters = append(analysis.Parameters, dashboardAnalysisParameter{
			ID:          parameter.ID,
			Name:        parameter.Name,
			Slug:        parameter.Slug,
			Type:        parameter.Type,
			MappedCards: buildDashboardParamMappingRows(dashboard, &parameter),
		})
	}

	tabNames := make(map[int]string, len(dashboard.Tabs))
	for _, tab := range dashboard.Tabs {
		tabNames[tab.ID] = tab.Name
	}

	cardCache := make(map[int]*client.Card)
	baseCards := make(map[int]*dashboardAnalysisBaseCard)
	flaggedCards := make(map[int]*dashboardAnalysisFlagged)

	for _, dashCard := range dashboard.DashCards {
		entry := dashboardAnalysisDashcard{
			DashcardID: dashCard.ID,
			Tab:        "Ungrouped",
			CardID:     dashCard.CardID,
		}
		if dashCard.TabID != nil {
			if tabName, ok := tabNames[*dashCard.TabID]; ok {
				entry.Tab = tabName
			}
		}
		if dashCard.Card != nil {
			entry.Name = dashCard.Card.Name
			entry.QueryType = dashCard.Card.QueryType
			entry.Display = dashCard.Card.Display
		}
		entry.ParameterIDs = collectDashcardParameterIDs(dashCard)

		if dashCard.CardID != nil {
			fullCard, chain, baseCard, err := traceCardLineage(ctx, c, cardCache, *dashCard.CardID)
			if err != nil {
				return nil, err
			}
			if fullCard != nil {
				if entry.Name == "" {
					entry.Name = fullCard.Name
				}
				if entry.QueryType == "" {
					entry.QueryType = fullCard.QueryType
				}
				if entry.Display == "" {
					entry.Display = fullCard.Display
				}
				entry.Flags = analyzeCardFlags(fullCard)
				if len(entry.Flags) > 0 {
					flaggedCards[fullCard.ID] = &dashboardAnalysisFlagged{CardID: fullCard.ID, Name: fullCard.Name, Flags: entry.Flags}
				}
				if sourceCardID := extractSourceCardID(fullCard); sourceCardID != nil {
					entry.SourceCardID = sourceCardID
				}
			}
			if len(chain) > 0 {
				entry.SourceCardChain = chain
			}
			if baseCard != nil {
				baseCardID := baseCard.ID
				entry.BaseCardID = &baseCardID
				baseEntry, ok := baseCards[baseCard.ID]
				if !ok {
					baseEntry = &dashboardAnalysisBaseCard{
						CardID:    baseCard.ID,
						Name:      baseCard.Name,
						QueryType: baseCard.QueryType,
						Flags:     analyzeCardFlags(baseCard),
					}
					baseCards[baseCard.ID] = baseEntry
				}
				baseEntry.ReferencedByDashcards = append(baseEntry.ReferencedByDashcards, dashCard.ID)
				if len(baseEntry.Flags) > 0 {
					flaggedCards[baseCard.ID] = &dashboardAnalysisFlagged{CardID: baseCard.ID, Name: baseCard.Name, Flags: baseEntry.Flags}
				}
			}
		}

		analysis.Dashcards = append(analysis.Dashcards, entry)
	}

	analysis.BaseCards = flattenBaseCards(baseCards)
	analysis.FlaggedCards = flattenFlaggedCards(flaggedCards)

	return analysis, nil
}

func buildDashboardAnalysisTabs(dashboard *client.Dashboard) []dashboardAnalysisTab {
	tabIndexes := make(map[int]int, len(dashboard.Tabs))
	tabs := make([]dashboardAnalysisTab, 0, len(dashboard.Tabs)+1)
	for _, tab := range dashboard.Tabs {
		tabs = append(tabs, dashboardAnalysisTab{ID: tab.ID, Name: tab.Name})
		tabIndexes[tab.ID] = len(tabs) - 1
	}

	ungroupedIndex := -1
	for _, dashCard := range dashboard.DashCards {
		cardID := 0
		if dashCard.CardID != nil {
			cardID = *dashCard.CardID
		}
		if dashCard.TabID == nil {
			if ungroupedIndex == -1 {
				tabs = append(tabs, dashboardAnalysisTab{Name: "Ungrouped"})
				ungroupedIndex = len(tabs) - 1
			}
			tabs[ungroupedIndex].DashcardIDs = append(tabs[ungroupedIndex].DashcardIDs, dashCard.ID)
			if cardID != 0 {
				tabs[ungroupedIndex].CardIDs = append(tabs[ungroupedIndex].CardIDs, cardID)
			}
			continue
		}

		if tabIndex, ok := tabIndexes[*dashCard.TabID]; ok {
			tabs[tabIndex].DashcardIDs = append(tabs[tabIndex].DashcardIDs, dashCard.ID)
			if cardID != 0 {
				tabs[tabIndex].CardIDs = append(tabs[tabIndex].CardIDs, cardID)
			}
		}
	}

	return tabs
}

func collectDashcardParameterIDs(dashCard client.DashCard) []string {
	if len(dashCard.ParameterMappings) == 0 {
		return nil
	}

	ids := make([]string, 0, len(dashCard.ParameterMappings))
	seen := make(map[string]bool, len(dashCard.ParameterMappings))
	for _, mapping := range dashCard.ParameterMappings {
		if seen[mapping.ParameterID] {
			continue
		}
		seen[mapping.ParameterID] = true
		ids = append(ids, mapping.ParameterID)
	}
	sort.Strings(ids)
	return ids
}

func traceCardLineage(ctx context.Context, c *client.Client, cache map[int]*client.Card, cardID int) (*client.Card, []int, *client.Card, error) {
	current, err := getAnalyzedCard(ctx, c, cache, cardID)
	if err != nil {
		return nil, nil, nil, err
	}

	chain := make([]int, 0)
	seen := map[int]bool{cardID: true}
	for current != nil {
		sourceCardID := extractSourceCardID(current)
		if sourceCardID == nil || seen[*sourceCardID] {
			return cache[cardID], chain, current, nil
		}
		chain = append(chain, *sourceCardID)
		seen[*sourceCardID] = true

		current, err = getAnalyzedCard(ctx, c, cache, *sourceCardID)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return cache[cardID], chain, nil, nil
}

func getAnalyzedCard(ctx context.Context, c *client.Client, cache map[int]*client.Card, cardID int) (*client.Card, error) {
	if card, ok := cache[cardID]; ok {
		return card, nil
	}

	card, err := c.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	cache[cardID] = card
	return card, nil
}

func extractSourceCardID(card *client.Card) *int {
	if card == nil || card.DatasetQuery == nil || card.DatasetQuery.Query == nil {
		return nil
	}
	return card.DatasetQuery.Query.SourceCardID
}

func analyzeCardFlags(card *client.Card) []string {
	if card == nil || card.DatasetQuery == nil {
		return nil
	}

	flags := make([]string, 0)
	if card.DatasetQuery.Native != nil {
		query := card.DatasetQuery.Native.Query
		lowerQuery := strings.ToLower(query)
		if assumptionQueryPattern.MatchString(lowerQuery) {
			flags = append(flags, "mentions sample or assumption data")
		}
		if hardcodedQueryPattern.MatchString(query) {
			flags = append(flags, "contains hardcoded filter literals")
		}
	}
	if card.DatasetQuery.Query != nil && len(card.DatasetQuery.Query.Filter) > 0 {
		flags = append(flags, "contains fixed MBQL filters")
	}

	return uniqueStrings(flags)
}

func flattenBaseCards(baseCards map[int]*dashboardAnalysisBaseCard) []dashboardAnalysisBaseCard {
	ids := make([]int, 0, len(baseCards))
	for id := range baseCards {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	results := make([]dashboardAnalysisBaseCard, 0, len(ids))
	for _, id := range ids {
		entry := baseCards[id]
		sort.Ints(entry.ReferencedByDashcards)
		entry.ReferencedByDashcards = uniqueInts(entry.ReferencedByDashcards)
		results = append(results, *entry)
	}
	return results
}

func flattenFlaggedCards(flaggedCards map[int]*dashboardAnalysisFlagged) []dashboardAnalysisFlagged {
	ids := make([]int, 0, len(flaggedCards))
	for id := range flaggedCards {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	results := make([]dashboardAnalysisFlagged, 0, len(ids))
	for _, id := range ids {
		entry := flaggedCards[id]
		entry.Flags = uniqueStrings(entry.Flags)
		results = append(results, *entry)
	}
	return results
}

func formatDashboardAnalysisTable(writer io.Writer, analysis *dashboardAnalysis) error {
	tw := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "dashboard_id\t%d\n", analysis.DashboardID)
	fmt.Fprintf(tw, "name\t%s\n", analysis.Name)
	fmt.Fprintf(tw, "description\t%s\n", analysis.Description)
	fmt.Fprintf(tw, "total_dashcards\t%d\n", analysis.TotalDashcards)
	fmt.Fprintf(tw, "total_parameters\t%d\n", analysis.TotalParameters)
	fmt.Fprintf(tw, "unique_base_cards\t%d\n", len(analysis.BaseCards))
	fmt.Fprintf(tw, "flagged_cards\t%d\n", len(analysis.FlaggedCards))
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Tabs"); err != nil {
		return err
	}
	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "id\tname\tdashcards\tcards"); err != nil {
		return err
	}
	for _, tab := range analysis.Tabs {
		id := ""
		if tab.ID != 0 {
			id = strconv.Itoa(tab.ID)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", id, tab.Name, joinInts(tab.DashcardIDs), joinInts(tab.CardIDs)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Dashcards"); err != nil {
		return err
	}
	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "dashcard_id\ttab\tcard_id\tname\tquery_type\tsource_card_id\tbase_card_id\tparams\tflags"); err != nil {
		return err
	}
	for _, dashCard := range analysis.Dashcards {
		cardID := ""
		if dashCard.CardID != nil {
			cardID = strconv.Itoa(*dashCard.CardID)
		}
		sourceCardID := ""
		if dashCard.SourceCardID != nil {
			sourceCardID = strconv.Itoa(*dashCard.SourceCardID)
		}
		baseCardID := ""
		if dashCard.BaseCardID != nil {
			baseCardID = strconv.Itoa(*dashCard.BaseCardID)
		}
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", dashCard.DashcardID, dashCard.Tab, cardID, dashCard.Name, dashCard.QueryType, sourceCardID, baseCardID, strings.Join(dashCard.ParameterIDs, ","), strings.Join(dashCard.Flags, "; ")); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Parameters"); err != nil {
		return err
	}
	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "id\tslug\ttype\tmapped_cards"); err != nil {
		return err
	}
	for _, parameter := range analysis.Parameters {
		mapped := make([]string, 0, len(parameter.MappedCards))
		for _, row := range parameter.MappedCards {
			label := row.CardID
			if row.CardName != "" {
				label = fmt.Sprintf("%s:%s", row.CardID, row.CardName)
			}
			mapped = append(mapped, label)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", parameter.ID, parameter.Slug, parameter.Type, strings.Join(mapped, ", ")); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Base Cards"); err != nil {
		return err
	}
	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "card_id\tname\tquery_type\treferenced_by_dashcards\tflags"); err != nil {
		return err
	}
	for _, baseCard := range analysis.BaseCards {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", baseCard.CardID, baseCard.Name, baseCard.QueryType, joinInts(baseCard.ReferencedByDashcards), strings.Join(baseCard.Flags, "; ")); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if len(analysis.FlaggedCards) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Flagged Cards"); err != nil {
		return err
	}
	tw = tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "card_id\tname\tflags"); err != nil {
		return err
	}
	for _, flagged := range analysis.FlaggedCards {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\n", flagged.CardID, flagged.Name, strings.Join(flagged.Flags, "; ")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range uniqueInts(values) {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ",")
}

func uniqueInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, 0, len(values))
	seen := make(map[int]bool, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
