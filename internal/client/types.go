package client

import (
	"encoding/json"
	"fmt"
)

// Database represents a Metabase database.
type Database struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Engine  string  `json:"engine"`
	Details any     `json:"details,omitempty"`
	Tables  []Table `json:"tables,omitempty"`
}

// DatabaseMetadata represents full database metadata including tables and fields.
type DatabaseMetadata struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	Engine string          `json:"engine"`
	Tables []TableMetadata `json:"tables,omitempty"`
}

// Table represents a Metabase table.
type Table struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Schema      string `json:"schema"`
	DBId        int    `json:"db_id"`
	EntityType  string `json:"entity_type,omitempty"`
}

// TableMetadata represents table metadata including field details.
type TableMetadata struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Schema      string  `json:"schema"`
	DBId        int     `json:"db_id"`
	Fields      []Field `json:"fields,omitempty"`
}

// Field represents a Metabase field (column).
type Field struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	BaseType     string `json:"base_type"`
	DatabaseType string `json:"database_type"`
	SemanticType string `json:"semantic_type,omitempty"`
	TableID      int    `json:"table_id"`
	TableName    string `json:"table_name,omitempty"`
}

// ForeignKey represents a foreign key relationship.
type ForeignKey struct {
	Relationship string     `json:"relationship"`
	Origin       FKFieldRef `json:"origin"`
	Destination  FKFieldRef `json:"destination"`
}

// FKFieldRef represents a field reference in a foreign key.
type FKFieldRef struct {
	ID    int        `json:"id"`
	Name  string     `json:"name"`
	Table FKTableRef `json:"table"`
}

// FKTableRef represents a table reference in a foreign key.
type FKTableRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// QueryResult represents the result of a dataset query.
type QueryResult struct {
	Data QueryResultData `json:"data"`
}

// QueryResultData holds the columns and rows of a query result.
type QueryResultData struct {
	Columns []ResultColumn `json:"cols"`
	Rows    [][]any        `json:"rows"`
}

// ResultColumn describes a column in a query result.
type ResultColumn struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	BaseType     string `json:"base_type"`
	SemanticType string `json:"semantic_type,omitempty"`
}

// DatasetQuery represents a query request to the Metabase dataset API.
type DatasetQuery struct {
	Database int              `json:"database"`
	Type     string           `json:"type"`
	Native   *NativeQuery     `json:"native,omitempty"`
	Query    *StructuredQuery `json:"query,omitempty"`
}

// NativeQuery represents the native SQL query part of a dataset query.
type NativeQuery struct {
	Query        string                 `json:"query"`
	TemplateTags map[string]TemplateTag `json:"template-tags,omitempty"`
}

// TemplateTag represents a native query template tag.
type TemplateTag struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display-name,omitempty"`
	Type        string `json:"type,omitempty"`
	WidgetType  string `json:"widget-type,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// StructuredQuery represents an MBQL structured query.
type StructuredQuery struct {
	SourceTable  any   `json:"source-table"`
	SourceCardID *int  `json:"source-card,omitempty"`
	Filter       []any `json:"filter,omitempty"`
	Limit        int   `json:"limit,omitempty"`
}

// Card represents a Metabase saved question (card).
type Card struct {
	ID                    int            `json:"id"`
	Name                  string         `json:"name"`
	Description           string         `json:"description,omitempty"`
	DatabaseID            int            `json:"database_id"`
	Display               string         `json:"display"`
	QueryType             string         `json:"query_type,omitempty"`
	CollectionID          *int           `json:"collection_id,omitempty"`
	Archived              bool           `json:"archived"`
	DatasetQuery          *DatasetQuery  `json:"dataset_query,omitempty"`
	ResultMetadata        []Field        `json:"result_metadata,omitempty"`
	VisualizationSettings map[string]any `json:"visualization_settings,omitempty"`
}

// Dashboard represents a Metabase dashboard.
type Dashboard struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	DashCards   []DashCard      `json:"dashcards,omitempty"`
	Parameters  []DashParameter `json:"parameters,omitempty"`
	Tabs        []DashTab       `json:"tabs,omitempty"`
	Archived    bool            `json:"archived"`
}

// DashCard represents a card placed on a dashboard.
type DashCard struct {
	ID                int                    `json:"id"`
	CardID            *int                   `json:"card_id,omitempty"`
	Card              *Card                  `json:"card,omitempty"`
	TabID             *int                   `json:"dashboard_tab_id,omitempty"`
	ParameterMappings []DashParameterMapping `json:"parameter_mappings,omitempty"`
}

// DashParameter represents a dashboard filter parameter.
type DashParameter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

// DashTab represents a dashboard tab.
type DashTab struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DashParameterMapping describes how a dashboard parameter maps to a card target.
type DashParameterMapping struct {
	CardID      int    `json:"card_id,omitempty"`
	ParameterID string `json:"parameter_id"`
	Target      []any  `json:"target,omitempty"`
}

// ParameterValues represents valid values for a dashboard parameter.
type ParameterValues struct {
	Values        []ParameterValue `json:"values"`
	HasMoreValues bool             `json:"has_more_values"`
}

// ParameterValue represents a value and optional display label for a parameter.
type ParameterValue struct {
	Value any    `json:"value"`
	Label string `json:"label,omitempty"`
}

// QueryParameter represents a parameter passed to a card or dashboard query.
type QueryParameter struct {
	ID     string `json:"id"`
	Type   string `json:"type,omitempty"`
	Target []any  `json:"target,omitempty"`
	Value  any    `json:"value"`
}

// UnmarshalJSON supports Metabase parameter value tuples: [value] or [value, label].
func (p *ParameterValue) UnmarshalJSON(data []byte) error {
	var tuple []any
	if err := json.Unmarshal(data, &tuple); err == nil {
		switch len(tuple) {
		case 0:
			p.Value = nil
			p.Label = ""
			return nil
		case 1:
			p.Value = tuple[0]
			p.Label = ""
			return nil
		default:
			p.Value = tuple[0]
			p.Label = fmt.Sprintf("%v", tuple[1])
			return nil
		}
	}

	var scalar any
	if err := json.Unmarshal(data, &scalar); err != nil {
		return err
	}

	p.Value = scalar
	p.Label = ""
	return nil
}

// SearchResult represents an item returned by the Metabase search API.
type SearchResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Model        string `json:"model"`
	DatabaseID   int    `json:"database_id,omitempty"`
	TableID      int    `json:"table_id,omitempty"`
	CollectionID *int   `json:"collection_id,omitempty"`
	Archived     bool   `json:"archived"`
}

// FieldSummary represents summary statistics for a field.
type FieldSummary struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// FieldValues represents distinct values for a field.
type FieldValues struct {
	FieldID int     `json:"field_id"`
	Values  [][]any `json:"values"`
}

// Collection represents a Metabase collection. The ID is typically a number,
// but the root collection is identified by the string "root".
type Collection struct {
	ID          any    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Location    string `json:"location,omitempty"`
	Archived    bool   `json:"archived"`
}

// CollectionItem represents an item contained in a collection, such as a card,
// dashboard, or nested collection.
type CollectionItem struct {
	ID          int    `json:"id"`
	Model       string `json:"model"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CollectionItems represents the items returned for a collection.
type CollectionItems struct {
	Total int              `json:"total"`
	Data  []CollectionItem `json:"data"`
}
