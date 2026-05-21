package client

import "context"

// RedactedValue is the replacement string for redacted PII values.
const RedactedValue = "[REDACTED]"

var piiSemanticTypes = map[string]bool{
	"type/Email":     true,
	"type/Name":      true,
	"type/Phone":     true,
	"type/Address":   true,
	"type/City":      true,
	"type/State":     true,
	"type/ZipCode":   true,
	"type/Country":   true,
	"type/Latitude":  true,
	"type/Longitude": true,
	"type/Birthdate": true,
	"type/AvatarURL": true,
	"type/URL":       true,
	"type/ImageURL":  true,
	"type/Company":   true,
}

// IsPIISemanticType returns true if the given Metabase semantic type is considered PII.
func IsPIISemanticType(semanticType string) bool {
	return piiSemanticTypes[semanticType]
}

// RedactQueryResult replaces values in PII columns with RedactedValue.
func RedactQueryResult(result *QueryResult) {
	redactIndices := make(map[int]bool)
	for i, col := range result.Data.Columns {
		if IsPIISemanticType(col.SemanticType) {
			redactIndices[i] = true
		}
	}

	if len(redactIndices) == 0 {
		return
	}

	for r := range result.Data.Rows {
		for i := range result.Data.Rows[r] {
			if redactIndices[i] {
				result.Data.Rows[r][i] = RedactedValue
			}
		}
	}
}

// EnrichSemanticTypes fills in missing semantic types on result columns by looking
// up field metadata from the database. This is needed for native SQL queries where
// Metabase does not return semantic types in the result columns.
func (c *Client) EnrichSemanticTypes(ctx context.Context, result *QueryResult, databaseID int) {
	needsEnrichment := false
	for _, col := range result.Data.Columns {
		if col.SemanticType == "" {
			needsEnrichment = true
			break
		}
	}

	if !needsEnrichment {
		return
	}

	fields, err := c.GetDatabaseFields(ctx, databaseID)
	if err != nil {
		return
	}

	// Build a map of field name → semantic type. If multiple fields share a name
	// and any has a PII semantic type, prefer the PII type (err on side of caution).
	fieldTypes := make(map[string]string)
	for _, f := range fields {
		name := f.Name
		existing, ok := fieldTypes[name]
		if !ok || (f.SemanticType != "" && (!ok || !IsPIISemanticType(existing))) {
			fieldTypes[name] = f.SemanticType
		}
	}

	for i := range result.Data.Columns {
		if result.Data.Columns[i].SemanticType == "" {
			if st, ok := fieldTypes[result.Data.Columns[i].Name]; ok {
				result.Data.Columns[i].SemanticType = st
			}
		}
	}
}

// RedactFieldValues replaces all values in a FieldValues struct with RedactedValue.
func RedactFieldValues(values *FieldValues) {
	for i := range values.Values {
		for j := range values.Values[i] {
			values.Values[i][j] = RedactedValue
		}
	}
}
