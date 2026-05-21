package tests

import (
	"strings"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/cli"
)

func TestContextContent(t *testing.T) {
	content := cli.ContextContent()

	if content == "" {
		t.Fatal("context content is empty")
	}

	requiredSections := []string{
		"# mb-cli - Agent Context",
		"## Authentication",
		"## Commands",
		"## Global Flags",
		"## Flags That Do NOT Exist",
		"## Database Name Resolution",
		"## Output Formats",
		"## Agent Workflows",
		"## Safe Querying",
		"## PII Redaction",
		"## Examples",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("context content missing section: %s", section)
		}
	}
}

func TestContextContentContainsAgentWorkflows(t *testing.T) {
	content := cli.ContextContent()

	workflows := []string{
		"### Explore a database schema",
		"### Inspect a dashboard",
		"### Explore a saved question (card)",
		"### Answer an ad-hoc data question",
	}

	for _, workflow := range workflows {
		if !strings.Contains(content, workflow) {
			t.Errorf("context content missing workflow: %s", workflow)
		}
	}
}

func TestContextContentContainsSafeQueryingGuidance(t *testing.T) {
	content := cli.ContextContent()

	markers := []string{
		"SELECT only",
		"Always bound result size",
		"Prefer `query filter` over hand-written SQL",
	}

	for _, marker := range markers {
		if !strings.Contains(content, marker) {
			t.Errorf("context content missing safe querying marker: %s", marker)
		}
	}
}

func TestContextContentContainsPIIRedactionGuidance(t *testing.T) {
	content := cli.ContextContent()

	markers := []string{
		"PII redaction is **enabled by default**",
		"Do NOT disable PII redaction",
		"--redact-pii=false",
		"[REDACTED]",
	}

	for _, marker := range markers {
		if !strings.Contains(content, marker) {
			t.Errorf("context content missing PII redaction marker: %s", marker)
		}
	}
}

func TestContextContentContainsKeyCommands(t *testing.T) {
	content := cli.ContextContent()

	commands := []string{
		"database list",
		"table list",
		"field get",
		"query sql",
		"card list",
		"dashboard list",
		"dashboard analyze",
		"search",
		"context",
		"version",
	}

	for _, cmd := range commands {
		if !strings.Contains(content, cmd) {
			t.Errorf("context content missing command: %s", cmd)
		}
	}
}

func TestContextContentContainsEnvVars(t *testing.T) {
	content := cli.ContextContent()

	envVars := []string{"MB_HOST", "MB_API_KEY"}
	for _, v := range envVars {
		if !strings.Contains(content, v) {
			t.Errorf("context content missing env var: %s", v)
		}
	}
}
