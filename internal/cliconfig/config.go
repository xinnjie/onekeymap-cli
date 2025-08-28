package cliconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration settings
type Config struct {
	Verbose                  bool   `mapstructure:"verbose"`
	Quiet                    bool   `mapstructure:"quiet"`
	OneKeyMap                string `mapstructure:"onekeymap"`
	OtelExporterOtlpEndpoint string `mapstructure:"otel.exporter.otlp.endpoint"`
}

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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			fmt.Printf("Warning: Error reading config file: %v\n", err)
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Verbose && c.Quiet {
		return fmt.Errorf("verbose and quiet modes cannot be enabled simultaneously")
	}
	return nil
}
