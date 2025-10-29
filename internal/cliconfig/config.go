package cliconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	//nolint:gochecknoglobals // use go build -ldflags "-X github.com/xinnjie/onekeymap-cli/internal/cliconfig.telemetryEndpoint=${TELEMETRY_ENDPOINT}"
	telemetryEndpoint = ""
	//nolint:gochecknoglobals // use go build -ldflags "-X github.com/xinnjie/onekeymap-cli/internal/cliconfig.telemetryHeaders=${TELEMETRY_HEADERS}"
	telemetryHeaders = ""
)

// EditorConfig holds per-editor configuration settings.
type EditorConfig struct {
	// KeymapPath is the path to the editor's keymap file.
	KeymapPath string `mapstructure:"keymap_path"`
	// TODO(xinnjie): Sync command is not implement yet
	// SyncEnabled specifies whether keymap syncing is enabled for this editor.
	SyncEnabled bool `mapstructure:"sync_enabled"`
}

// TelemetryConfig holds OpenTelemetry configuration.
type TelemetryConfig struct {
	// Enabled controls whether telemetry is enabled (default: false).
	// When this field is explicitly set in config (true or false), no prompt will be shown.
	// When not set, a prompt will be displayed to ask user's preference.
	Enabled bool `mapstructure:"enabled"`
	// Endpoint is the OTLP exporter endpoint (host:port, without http:// or https://).
	// Example: "otlp-gateway-prod-us-central-0.grafana.net:443"
	Endpoint string `mapstructure:"endpoint"`
	// Headers are custom headers for OTLP exporter (e.g., for authentication).
	// Format: "key1=value1,key2=value2"
	Headers string `mapstructure:"headers"`
}

type Config struct {
	// OneKeyMap is the path to the main onekeymap configuration file.
	OneKeyMap string `mapstructure:"onekeymap"`
	// Telemetry holds OpenTelemetry configuration.
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
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
// - ONEKEYMAP_TELEMETRY_ENABLED -> telemetry.enabled (bool)
// - ONEKEYMAP_TELEMETRY_ENDPOINT -> telemetry.endpoint (string)
// - ONEKEYMAP_TELEMETRY_HEADERS -> telemetry.headers (string, "key1=value1,key2=value2")
// - ONEKEYMAP_TELEMETRY_INSECURE -> telemetry.insecure (bool)
// - ONEKEYMAP_SERVER_LISTEN -> server.listen (string, e.g. "tcp://127.0.0.1:50051" or "unix:///tmp/onekeymap.sock")

// NewConfig initializes and returns a new Config object.
// It sets defaults, binds environment variables, reads config files, and unmarshals the result.
func NewConfig(sandbox bool) (*Config, error) {
	if !sandbox {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get home directory: %w", err)
		}
		viper.SetDefault("onekeymap", filepath.Join(homeDir, ".config", "onekeymap", "onekeymap.json"))
	}
	// Note: We don't set default for telemetry.enabled to detect if it's explicitly configured
	viper.SetDefault("telemetry.endpoint", telemetryEndpoint)
	viper.SetDefault("telemetry.headers", telemetryHeaders)
	viper.SetDefault("server.listen", "")

	// Set environment variable handling
	viper.SetEnvPrefix("ONEKEYMAP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if !sandbox {
		// Set configuration file name and paths
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("${HOME}/.config/onekeymap")
		viper.AddConfigPath("/etc/onekeymap")

		// Read configuration file if it exists
		if err := viper.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			// Config file not found is ignored
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
	return nil
}

// IsTelemetryExplicitlySet returns true if telemetry.enabled is explicitly set in config or environment.
func IsTelemetryExplicitlySet() bool {
	return viper.IsSet("telemetry.enabled")
}
