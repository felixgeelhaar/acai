package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the persistent YAML configuration file.
// Only settings are stored here — tokens live in credentials.json.
type FileConfig struct {
	DataSource string            `yaml:"data_source,omitempty"`
	Granola    GranolaFileConfig `yaml:"granola,omitempty"`
}

// GranolaFileConfig holds Granola-specific file configuration.
type GranolaFileConfig struct {
	APIURL    string `yaml:"api_url,omitempty"`
	CachePath string `yaml:"cache_path,omitempty"`
}

// ReadConfigFile reads a YAML config file from path.
// Returns an empty FileConfig (no error) if the file does not exist.
func ReadConfigFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &FileConfig{}, nil
		}
		return nil, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteConfigFile writes a FileConfig to path as YAML.
// Creates parent directories and sets file permissions to 0600.
func WriteConfigFile(path string, cfg FileConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
