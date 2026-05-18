package tests

import (
	"os"
	"testing"

	"github.com/andreagrandi/mb-cli/internal/config"
)

func TestLoadConfig_Success(t *testing.T) {
	os.Setenv("MB_HOST", "https://metabase.example.com")
	os.Setenv("MB_API_KEY", "test-api-key")
	os.Unsetenv("MB_SESSION_TOKEN")
	defer os.Unsetenv("MB_HOST")
	defer os.Unsetenv("MB_API_KEY")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Host != "https://metabase.example.com" {
		t.Errorf("expected host 'https://metabase.example.com', got '%s'", cfg.Host)
	}

	if cfg.APIKey != "test-api-key" {
		t.Errorf("expected api key 'test-api-key', got '%s'", cfg.APIKey)
	}

	if cfg.SessionToken != "" {
		t.Errorf("expected empty session token, got '%s'", cfg.SessionToken)
	}
}

func TestLoadConfig_SessionTokenOnly(t *testing.T) {
	os.Setenv("MB_HOST", "https://metabase.example.com")
	os.Unsetenv("MB_API_KEY")
	os.Setenv("MB_SESSION_TOKEN", "test-session-token")
	defer os.Unsetenv("MB_HOST")
	defer os.Unsetenv("MB_SESSION_TOKEN")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.SessionToken != "test-session-token" {
		t.Errorf("expected session token 'test-session-token', got '%s'", cfg.SessionToken)
	}

	if cfg.APIKey != "" {
		t.Errorf("expected empty api key, got '%s'", cfg.APIKey)
	}
}

func TestLoadConfig_BothAuthMethodsRejected(t *testing.T) {
	os.Setenv("MB_HOST", "https://metabase.example.com")
	os.Setenv("MB_API_KEY", "test-api-key")
	os.Setenv("MB_SESSION_TOKEN", "test-session-token")
	defer os.Unsetenv("MB_HOST")
	defer os.Unsetenv("MB_API_KEY")
	defer os.Unsetenv("MB_SESSION_TOKEN")

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error when both MB_API_KEY and MB_SESSION_TOKEN are set")
	}

	expected := "MB_API_KEY and MB_SESSION_TOKEN are mutually exclusive, set only one"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestLoadConfig_MissingHost(t *testing.T) {
	os.Unsetenv("MB_HOST")
	os.Setenv("MB_API_KEY", "test-api-key")
	os.Unsetenv("MB_SESSION_TOKEN")
	defer os.Unsetenv("MB_API_KEY")

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error when MB_HOST is missing")
	}

	expected := "MB_HOST environment variable is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestLoadConfig_NoAuthMethod(t *testing.T) {
	os.Setenv("MB_HOST", "https://metabase.example.com")
	os.Unsetenv("MB_API_KEY")
	os.Unsetenv("MB_SESSION_TOKEN")
	defer os.Unsetenv("MB_HOST")

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error when neither MB_API_KEY nor MB_SESSION_TOKEN is set")
	}

	expected := "either MB_API_KEY or MB_SESSION_TOKEN environment variable is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestLoadConfig_BothMissing(t *testing.T) {
	os.Unsetenv("MB_HOST")
	os.Unsetenv("MB_API_KEY")
	os.Unsetenv("MB_SESSION_TOKEN")

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("expected error when all env vars are missing")
	}
}
