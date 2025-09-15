package cliconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// EditorConfig holds per-editor configuration settings.
type EditorConfig struct {
	// KeymapPath is the path to the editor's keymap file.
	KeymapPath string `mapstructure:"keymap_path"`
	// TODO(xinnjie): Sync command is not implement yet
	// SyncEnabled specifies whether keymap syncing is enabled for this editor.
	SyncEnabled bool `mapstructure:"sync_enabled"`
}

type Config struct {
	// Verbose enables verbose output.
	Verbose bool `mapstructure:"verbose"`
	// Quiet suppresses all output except for errors.
	Quiet bool `mapstructure:"quiet"`
	// OneKeyMap is the path to the main onekeymap configuration file.
	OneKeyMap string `mapstructure:"onekeymap"`
	// OtelExporterOtlpEndpoint is the OpenTelemetry OTLP exporter endpoint.
	OtelExporterOtlpEndpoint string `mapstructure:"otel.exporter.otlp.endpoint"`
	// ServerListen is the server listen address (e.g., "tcp://127.0.0.1:50051" or "unix:///tmp/onekeymap.sock").
	ServerListen string `mapstructure:"server.listen"`
	// Editors holds configuration for different editors.
	Editors map[string]EditorConfig `mapstructure:"editors"`
}

// Environment variables mapping
//
// Viper is configured with:
// - Prefix: ONEKEYMAP
// - Key replacer: '.' -> '_'
//
// Therefore, the following environment variables map to config keys:
// - ONEKEYMAP_VERBOSE -> verbose (bool)
// - ONEKEYMAP_QUIET -> quiet (bool)
// - ONEKEYMAP_ONEKEYMAP -> onekeymap (string, file path)
// - ONEKEYMAP_OTEL_EXPORTER_OTLP_ENDPOINT -> otel.exporter.otlp.endpoint (string)
// - ONEKEYMAP_SERVER_LISTEN -> server.listen (string, e.g. "tcp://127.0.0.1:50051" or "unix:///tmp/onekeymap.sock")

// NewConfig initializes and returns a new Config object.
// It sets defaults, binds environment variables, reads config files, and unmarshals the result.
func NewConfig() (*Config, error) {
	// Set default values
	viper.SetDefault("verbose", false)
	viper.SetDefault("quiet", false)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get home directory: %w", err)
	}
	viper.SetDefault("onekeymap", filepath.Join(homeDir, ".config", "onekeymap", "onekeymap.json"))
	viper.SetDefault("otel.exporter.otlp.endpoint", "")
	viper.SetDefault("server.listen", "")

	// Set configuration file name and paths
	viper.SetConfigName("onekeymap")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("${HOME}/.config/onekeymap")
	viper.AddConfigPath("/etc/onekeymap")

	// Set environment variable handling
	viper.SetEnvPrefix("ONEKEYMAP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read configuration file if it exists
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// TODO(xinnjie): Low priority, may be use cobra.command to log
			//nolint:forbidigo
			fmt.Printf("Config file not found: %v\n", err)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal the configuration into a new Config instance
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Verbose && c.Quiet {
		return errors.New("verbose and quiet modes cannot be enabled simultaneously")
	}
	return nil
}
