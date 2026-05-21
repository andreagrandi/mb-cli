package config

import (
	"os"

	"github.com/andreagrandi/mb-cli/internal/mberr"
)

type Config struct {
	Host         string
	APIKey       string
	SessionToken string
}

func LoadConfig() (*Config, error) {
	host := os.Getenv("MB_HOST")
	if host == "" {
		return nil, &mberr.ConfigError{Message: "MB_HOST environment variable is required"}
	}

	apiKey := os.Getenv("MB_API_KEY")
	sessionToken := os.Getenv("MB_SESSION_TOKEN")

	if apiKey == "" && sessionToken == "" {
		return nil, &mberr.ConfigError{Message: "either MB_API_KEY or MB_SESSION_TOKEN environment variable is required"}
	}

	if apiKey != "" && sessionToken != "" {
		return nil, &mberr.ConfigError{Message: "MB_API_KEY and MB_SESSION_TOKEN are mutually exclusive, set only one"}
	}

	return &Config{
		Host:         host,
		APIKey:       apiKey,
		SessionToken: sessionToken,
	}, nil
}
