package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the TOML configuration structure
type Config struct {
	Source              string   `toml:"source"`
	Dest                string   `toml:"dest"`
	Mode                string   `toml:"mode"`
	DryRun              bool     `toml:"dry_run"`
	UseSourceGitignore  bool     `toml:"use_source_gitignore"`
	ExcludeHiddenDirs   bool     `toml:"exclude_hidden_dirs"`
	OnlySyncignore      bool     `toml:"only_syncignore"`
	IgnoreSrc           []string `toml:"ignore_src"`
	IgnoreDest          []string `toml:"ignore_dest"`
	Only                []string `toml:"only"`
	LogLevel            string   `toml:"log_level"`
	LogFile             string   `toml:"log_file"`
	LogFormat           string   `toml:"log_format"`
	Report              string   `toml:"report"`
}

// LoadConfig loads configuration from a TOML file
// If configPath is empty, it will look for common config files
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// If no config path specified, look for common config files
	if configPath == "" {
		// Look in current directory first
		candidates := []string{
			"sync.toml",
			".sync.toml",
		}

		// Look in source directory if it exists in environment
		if srcDir := os.Getenv("SYNC_TOOLS_SOURCE"); srcDir != "" {
			candidates = append(candidates,
				filepath.Join(srcDir, "sync.toml"),
				filepath.Join(srcDir, ".sync.toml"),
			)
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}

		// If still no config file found, return empty config
		if configPath == "" {
			return &config, nil
		}
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	// Decode the TOML file
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate mode if specified
	if config.Mode != "" && config.Mode != "one-way" && config.Mode != "two-way" {
		return fmt.Errorf("invalid mode: %s (must be 'one-way' or 'two-way')", config.Mode)
	}

	// Validate log level if specified
	validLogLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	if config.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if config.LogLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level: %s (must be one of: %v)", config.LogLevel, validLogLevels)
		}
	}

	// Validate log format if specified
	if config.LogFormat != "" && config.LogFormat != "text" && config.LogFormat != "json" {
		return fmt.Errorf("invalid log format: %s (must be 'text' or 'json')", config.LogFormat)
	}

	return nil
}