package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Feed represents a single RSS/Atom feed configuration
type Feed struct {
	Name         string `json:"name" mapstructure:"name"`
	URL          string `json:"url" mapstructure:"url"`
	TitleKey     string `json:"title-key,omitempty" mapstructure:"title-key"`
	BodyKey      string `json:"body-key,omitempty" mapstructure:"body-key"`
	TimestampKey string `json:"timestamp-key,omitempty" mapstructure:"timestamp-key"`
}

// Config represents the application configuration
type Config struct {
	Feeds []Feed `json:"feeds" mapstructure:"feeds"`
}

// SetDefaults sets default configuration values
func SetDefaults() {
	viper.SetDefault("feeds", []map[string]interface{}{
		{
			"name":          "Arch Linux News",
			"url":           "https://archlinux.org/feeds/news/",
			"title-key":     "title",
			"body-key":      "summary",
			"timestamp-key": "published",
		},
	})
}

// Load loads the configuration from viper
func Load() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults for feed keys if not specified
	for i := range cfg.Feeds {
		if cfg.Feeds[i].TitleKey == "" {
			cfg.Feeds[i].TitleKey = "title"
		}
		if cfg.Feeds[i].BodyKey == "" {
			cfg.Feeds[i].BodyKey = "summary"
		}
		if cfg.Feeds[i].TimestampKey == "" {
			cfg.Feeds[i].TimestampKey = "published"
		}
	}

	// Validate configuration
	for _, feed := range cfg.Feeds {
		if feed.URL == "" {
			return nil, fmt.Errorf("feed URL cannot be empty")
		}
	}

	return &cfg, nil
}

// GetConfigPath returns the path where the read status file should be stored
func GetConfigPath() (string, error) {
	// Try to use the same directory as the config file
	if viper.ConfigFileUsed() != "" {
		return filepath.Dir(viper.ConfigFileUsed()), nil
	}

	// Fallback to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Try XDG_CONFIG_HOME first
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return xdgConfig, nil
	}

	// Default to ~/.config or home directory
	configDir := filepath.Join(home, ".config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return home, nil
	}

	return configDir, nil
}
