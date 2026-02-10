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
}

type GranolaConfig struct {
	APIURL         string
	AuthMethod     string
	APIToken       string
	DataSource     string // "auto" (default), "api", "local_cache"
	LocalCachePath string // override path to cache-v3.json
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

	// Layer 2: config file (if exists)
	if path, err := DefaultConfigPath(); err == nil {
		applyFileConfig(cfg, path)
	}

	// Layer 3: env vars (highest priority)
	applyEnvOverrides(cfg)

	return cfg
}

// DefaultConfigPath returns the path to the default config file (~/.acai/config.yaml).
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".acai", "config.yaml"), nil
}

// applyFileConfig reads the YAML config file and overlays non-empty fields onto cfg.
func applyFileConfig(cfg *Config, path string) {
	fileCfg, err := ReadConfigFile(path)
	if err != nil {
		return
	}

	if fileCfg.DataSource != "" {
		cfg.Granola.DataSource = fileCfg.DataSource
	}
	if fileCfg.Granola.APIURL != "" {
		cfg.Granola.APIURL = fileCfg.Granola.APIURL
	}
	if fileCfg.Granola.CachePath != "" {
		cfg.Granola.LocalCachePath = fileCfg.Granola.CachePath
	}
}

// applyEnvOverrides applies environment variable overrides to cfg.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ACAI_GRANOLA_API_URL"); v != "" {
		cfg.Granola.APIURL = v
	}
	if v := os.Getenv("ACAI_GRANOLA_API_TOKEN"); v != "" {
		cfg.Granola.APIToken = v
	}
	if v := os.Getenv("ACAI_DATA_SOURCE"); v != "" {
		cfg.Granola.DataSource = v
	}
	if v := os.Getenv("ACAI_GRANOLA_CACHE_PATH"); v != "" {
		cfg.Granola.LocalCachePath = v
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
	if v := os.Getenv("ACAI_POLICY_FILE"); v != "" {
		cfg.Policy.FilePath = v
		cfg.Policy.Enabled = true
	}
}

func Default() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	return &Config{
		Granola: GranolaConfig{
			APIURL:     "https://public-api.granola.ai",
			AuthMethod: "api_token",
			DataSource: "auto",
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
