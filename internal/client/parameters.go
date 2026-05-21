package client

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// CardParameters returns the parameters a saved question accepts, derived from
// its native query template tags and sorted by name for stable output. Card
// and snippet template tags are excluded because they reference other queries
// rather than accepting a user-supplied value.
func (c *Card) CardParameters() []CardParameter {
	if c == nil || c.DatasetQuery == nil || c.DatasetQuery.Native == nil {
		return []CardParameter{}
	}

	tags := c.DatasetQuery.Native.TemplateTags
	params := make([]CardParameter, 0, len(tags))
	for key, tag := range tags {
		if tag.Type == "card" || tag.Type == "snippet" {
			continue
		}
		params = append(params, CardParameter{
			Name:        key,
			DisplayName: tag.DisplayName,
			Type:        tag.Type,
			WidgetType:  tag.WidgetType,
			Required:    tag.Required,
			Default:     tag.Default,
		})
	}

	sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })
	return params
}

func buildCardQueryParameters(card *Card, params map[string]string) []QueryParameter {
	if len(params) == 0 {
		return nil
	}

	resolved := make([]QueryParameter, 0, len(params))
	tags := map[string]TemplateTag(nil)
	if card != nil && card.DatasetQuery != nil && card.DatasetQuery.Native != nil {
		tags = card.DatasetQuery.Native.TemplateTags
	}

	for key, rawValue := range params {
		param := QueryParameter{
			ID:    key,
			Value: coerceQueryParameterValue(rawValue),
		}
		if _, tag, ok := resolveTemplateTag(tags, key); ok && tag.ID != "" {
			param.ID = tag.ID
		}
		resolved = append(resolved, param)
	}

	return resolved
}

func buildDashboardQueryParameters(dashboard *Dashboard, params map[string]string) []QueryParameter {
	if len(params) == 0 {
		return nil
	}

	resolved := make([]QueryParameter, 0, len(params))
	for key, rawValue := range params {
		param := QueryParameter{
			ID:    key,
			Value: coerceQueryParameterValue(rawValue),
		}
		if dashboardParam, ok := resolveDashboardParameterValue(dashboard, key); ok {
			param.ID = dashboardParam.ID
			param.Type = dashboardParam.Type
		}
		resolved = append(resolved, param)
	}

	return resolved
}

func resolveTemplateTag(tags map[string]TemplateTag, input string) (string, TemplateTag, bool) {
	if len(tags) == 0 {
		return "", TemplateTag{}, false
	}
	if tag, ok := tags[input]; ok {
		return input, tag, true
	}

	for key, tag := range tags {
		if tag.ID == input || strings.EqualFold(tag.Name, input) || strings.EqualFold(tag.DisplayName, input) {
			return key, tag, true
		}
	}

	return "", TemplateTag{}, false
}

func resolveDashboardParameterValue(dashboard *Dashboard, input string) (*DashParameter, bool) {
	if dashboard == nil {
		return nil, false
	}
	for i := range dashboard.Parameters {
		parameter := &dashboard.Parameters[i]
		if parameter.ID == input || parameter.Slug == input || strings.EqualFold(parameter.Name, input) {
			return parameter, true
		}
	}

	return nil, false
}

func coerceQueryParameterValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, `"`) {
		var decoded any
		if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
			return decoded
		}
	}

	switch strings.ToLower(trimmed) {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}

	if i, err := strconv.Atoi(trimmed); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return f
	}

	return raw
}
