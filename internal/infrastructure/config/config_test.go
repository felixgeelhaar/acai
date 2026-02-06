package config_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/infrastructure/config"
)

func TestDefault_HasSaneDefaults(t *testing.T) {
	cfg := config.Default()

	if cfg.Granola.APIURL != "https://api.granola.ai" {
		t.Errorf("got api url %q", cfg.Granola.APIURL)
	}
	if cfg.MCP.Transport != "stdio" {
		t.Errorf("got transport %q", cfg.MCP.Transport)
	}
	if cfg.MCP.ServerName != "granola-mcp" {
		t.Errorf("got server name %q", cfg.MCP.ServerName)
	}
	if cfg.Cache.TTL != 15*time.Minute {
		t.Errorf("got cache ttl %v", cfg.Cache.TTL)
	}
	if cfg.Resilience.Retry.MaxAttempts != 3 {
		t.Errorf("got max attempts %d", cfg.Resilience.Retry.MaxAttempts)
	}
	if cfg.Resilience.Timeout != 30*time.Second {
		t.Errorf("got timeout %v", cfg.Resilience.Timeout)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("got log level %q", cfg.Logging.Level)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("GRANOLA_MCP_GRANOLA_API_URL", "https://custom.api.com")
	t.Setenv("GRANOLA_MCP_GRANOLA_API_TOKEN", "test-token")
	t.Setenv("GRANOLA_MCP_LOGGING_LEVEL", "debug")
	t.Setenv("GRANOLA_MCP_MCP_HTTP_PORT", "9090")

	cfg := config.Load()

	if cfg.Granola.APIURL != "https://custom.api.com" {
		t.Errorf("got api url %q", cfg.Granola.APIURL)
	}
	if cfg.Granola.APIToken != "test-token" {
		t.Errorf("got api token %q", cfg.Granola.APIToken)
	}
	if cfg.Granola.AuthMethod != "api_token" {
		t.Errorf("got auth method %q, want api_token when token is set", cfg.Granola.AuthMethod)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("got log level %q", cfg.Logging.Level)
	}
	if cfg.MCP.HTTPPort != 9090 {
		t.Errorf("got http port %d", cfg.MCP.HTTPPort)
	}
}

func TestLoad_PolicyFileEnv(t *testing.T) {
	t.Setenv("GRANOLA_MCP_POLICY_FILE", "/etc/granola/policy.yaml")

	cfg := config.Load()

	if !cfg.Policy.Enabled {
		t.Error("expected policy enabled when file path is set")
	}
	if cfg.Policy.FilePath != "/etc/granola/policy.yaml" {
		t.Errorf("got policy file %q", cfg.Policy.FilePath)
	}
}

func TestDefault_PolicyDisabled(t *testing.T) {
	cfg := config.Default()

	if cfg.Policy.Enabled {
		t.Error("expected policy disabled by default")
	}
	if cfg.Policy.FilePath != "" {
		t.Errorf("got policy file %q, want empty", cfg.Policy.FilePath)
	}
}
