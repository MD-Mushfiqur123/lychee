// Package config provides configuration loading for the Lychee server using viper.
// It reads ~/.lychee/config.yaml and falls back to sensible defaults.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the Lychee server.
type Config struct {
	Host     string `mapstructure:"host"`
	ModelDir string `mapstructure:"model_dir"`
	LogLevel string `mapstructure:"log_level"`
	HuggingFace struct {
		CacheDir string `mapstructure:"cache_dir"`
		Token    string `mapstructure:"token"`
	} `mapstructure:"huggingface"`
}

// DefaultConfig returns sensible default configuration values.
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost:11434",
		ModelDir: "",
		LogLevel: "info",
	}
}

// Load reads ~/.lychee/config.yaml and unmarshals it into a Config.
// If the config file does not exist, defaults are returned without error.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("config: could not determine home directory, using defaults", "error", err)
		return cfg, nil
	}

	configDir := filepath.Join(home, ".lychee")
	configFile := filepath.Join(configDir, "config.yaml")

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// Set defaults from DefaultConfig
	v.SetDefault("host", cfg.Host)
	v.SetDefault("model_dir", cfg.ModelDir)
	v.SetDefault("log_level", cfg.LogLevel)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Debug("config: no config file found, using defaults", "path", configFile)
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("config: unable to decode config: %w", err)
	}

	slog.Info("config: loaded", "path", configFile)
	return cfg, nil
}
