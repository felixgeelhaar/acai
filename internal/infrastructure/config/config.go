// Package config handles application configuration loading.
// Configuration follows 12-factor: file defaults, overridable by environment variables.
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Granola    GranolaConfig
	MCP        MCPConfig
	Cache      CacheConfig
	Resilience ResilienceConfig
	Privacy    PrivacyConfig
	Policy     PolicyConfig
	Sync       SyncConfig
	Logging    LoggingConfig
	Webhook    WebhookConfig
}

type WebhookConfig struct {
	Secret string
}

type GranolaConfig struct {
	APIURL     string
	AuthMethod string
	APIToken   string
}

type MCPConfig struct {
	ServerName       string
	Transport        string
	HTTPPort         int
	EnabledResources []string
}

type CacheConfig struct {
	Enabled bool
	Dir     string
	TTL     time.Duration
}

type ResilienceConfig struct {
	CircuitBreaker CircuitBreakerConfig
	RateLimit      RateLimitConfig
	Retry          RetryConfig
	Timeout        time.Duration
}

type CircuitBreakerConfig struct {
	FailureThreshold uint32
	SuccessThreshold uint32
	HalfOpenTimeout  time.Duration
}

type RateLimitConfig struct {
	Rate     int
	Interval time.Duration
}

type RetryConfig struct {
	MaxAttempts  int
	Backoff      string
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

type PrivacyConfig struct {
	RedactSpeakers bool
	RedactKeywords []string
	LocalOnly      bool
}

type PolicyConfig struct {
	Enabled  bool
	FilePath string
}

type SyncConfig struct {
	PollingInterval time.Duration
	AutoSync        bool
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() *Config {
	cfg := Default()

	if v := os.Getenv("ACAI_GRANOLA_API_URL"); v != "" {
		cfg.Granola.APIURL = v
	}
	if v := os.Getenv("ACAI_GRANOLA_API_TOKEN"); v != "" {
		cfg.Granola.APIToken = v
		cfg.Granola.AuthMethod = "api_token"
	}
	if v := os.Getenv("ACAI_MCP_TRANSPORT"); v != "" {
		cfg.MCP.Transport = v
	}
	if v := os.Getenv("ACAI_MCP_HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.MCP.HTTPPort = port
		}
	}
	if v := os.Getenv("ACAI_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Cache.TTL = d
		}
	}
	if v := os.Getenv("ACAI_LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("ACAI_LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("ACAI_WEBHOOK_SECRET"); v != "" {
		cfg.Webhook.Secret = v
	}
	if v := os.Getenv("ACAI_POLICY_FILE"); v != "" {
		cfg.Policy.FilePath = v
		cfg.Policy.Enabled = true
	}

	return cfg
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Granola: GranolaConfig{
			APIURL:     "https://api.granola.ai",
			AuthMethod: "oauth",
		},
		MCP: MCPConfig{
			ServerName: "acai",
			Transport:  "stdio",
			HTTPPort:   8080,
			EnabledResources: []string{
				"meeting", "transcript", "summary", "action_item", "metadata",
			},
		},
		Cache: CacheConfig{
			Enabled: true,
			Dir:     filepath.Join(homeDir, ".acai", "cache"),
			TTL:     15 * time.Minute,
		},
		Resilience: ResilienceConfig{
			CircuitBreaker: CircuitBreakerConfig{
				FailureThreshold: 5,
				SuccessThreshold: 2,
				HalfOpenTimeout:  30 * time.Second,
			},
			RateLimit: RateLimitConfig{
				Rate:     100,
				Interval: time.Minute,
			},
			Retry: RetryConfig{
				MaxAttempts:  3,
				Backoff:      "exponential",
				InitialDelay: 500 * time.Millisecond,
				MaxDelay:     10 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		Privacy: PrivacyConfig{
			RedactSpeakers: false,
			RedactKeywords: []string{},
			LocalOnly:      false,
		},
		Sync: SyncConfig{
			PollingInterval: 5 * time.Minute,
			AutoSync:        false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
	}
}
