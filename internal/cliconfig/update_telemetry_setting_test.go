package cliconfig_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/cliconfig"
)

// nolint:gochecknoglobals // TestConfig for testing
var testExampleConfig = `# Test Configuration
onekeymap: ~/.config/onekeymap/onekeymap.json

# Telemetry configuration
telemetry:
  # Can also be enabled with --telemetry flag
  # enabled: false
  endpoint: "test.example.com"

# Other settings
server:
  listen: ":8080"
`

func setupTestDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// TestUpdateTelemetrySettingsWithOptions tests the main function with dependency injection
func TestUpdateTelemetrySettingsWithOptions(t *testing.T) {
	tests := []struct {
		name             string
		enabled          bool
		existingConfig   string
		getConfigFile    func() string
		getHomeDir       func() (string, error)
		mkdirAll         func(path string, perm os.FileMode) error
		getTemplate      func() ([]byte, error)
		wantError        bool
		wantConfigExists bool
		wantConfig       string
	}{
		{
			name:    "create new config with telemetry enabled",
			enabled: true,
			getConfigFile: func() string {
				return "" // no existing config
			},
			getHomeDir: func() (string, error) {
				return "/tmp/test", nil
			},
			mkdirAll: os.MkdirAll,
			getTemplate: func() ([]byte, error) {
				return []byte(testExampleConfig), nil
			},
			wantConfigExists: true,
			wantConfig: `# Test Configuration
onekeymap: ~/.config/onekeymap/onekeymap.json
# Telemetry configuration
telemetry:
  enabled: true
  # Can also be enabled with --telemetry flag
  # enabled: false
  endpoint: "test.example.com"
# Other settings
server:
  listen: ":8080"
`,
		},
		{
			name:    "create new config with telemetry disabled",
			enabled: false,
			getConfigFile: func() string {
				return "" // no existing config
			},
			getHomeDir: func() (string, error) {
				return "/tmp/test", nil
			},
			mkdirAll: os.MkdirAll,
			getTemplate: func() ([]byte, error) {
				return []byte(testExampleConfig), nil
			},
			wantConfigExists: true,
			wantConfig: `# Test Configuration
onekeymap: ~/.config/onekeymap/onekeymap.json
# Telemetry configuration
telemetry:
  enabled: false
  # Can also be enabled with --telemetry flag
  # enabled: false
  endpoint: "test.example.com"
# Other settings
server:
  listen: ":8080"
`,
		},
		{
			name:    "update existing config",
			enabled: true,
			existingConfig: `onekeymap: test.json
telemetry:
  enabled: false
  endpoint: example.com
`,
			getConfigFile: func() string {
				return "/tmp/existing_config.yaml"
			},
			wantConfigExists: true,
			wantConfig: `onekeymap: test.json
telemetry:
  enabled: true
  endpoint: example.com
`,
		},
		{
			name:    "error getting home directory",
			enabled: true,
			getConfigFile: func() string {
				return ""
			},
			getHomeDir: func() (string, error) {
				return "", errors.New("home directory error")
			},
			wantError: true,
		},
		{
			name:    "error creating directory",
			enabled: true,
			getConfigFile: func() string {
				return ""
			},
			getHomeDir: func() (string, error) {
				return "/tmp/test", nil
			},
			mkdirAll: func(_ string, _ os.FileMode) error {
				return errors.New("mkdir error")
			},
			wantError: true,
		},
		{
			name:    "error getting template",
			enabled: true,
			getConfigFile: func() string {
				return ""
			},
			getHomeDir: func() (string, error) {
				return "/tmp/test", nil
			},
			mkdirAll: func(_ string, _ os.FileMode) error {
				return nil
			},
			getTemplate: func() ([]byte, error) {
				return nil, errors.New("template error")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := setupTestDir(t)

			// Override functions for tests that create new configs
			if tt.name == "create new config with telemetry enabled" ||
				tt.name == "create new config with telemetry disabled" {
				tt.getHomeDir = func() (string, error) {
					return tempDir, nil
				}
			}

			// Setup existing config if needed
			var existingConfigPath string
			if tt.existingConfig != "" {
				existingConfigPath = filepath.Join(tempDir, "existing_config.yaml")
				err := os.WriteFile(existingConfigPath, []byte(tt.existingConfig), 0600)
				require.NoError(t, err)
			}

			// Override getConfigFile if we have existing config
			if tt.existingConfig != "" && tt.getConfigFile != nil {
				originalGetConfigFile := tt.getConfigFile
				tt.getConfigFile = func() string {
					path := originalGetConfigFile()
					if path != "" {
						return existingConfigPath
					}
					return path
				}
			}

			opts := cliconfig.UpdateOptions{
				GetConfigFile: tt.getConfigFile,
				GetHomeDir:    tt.getHomeDir,
				MkdirAll:      tt.mkdirAll,
				GetTemplate:   tt.getTemplate,
			}

			err := cliconfig.UpdateTelemetrySettingsWithOptions(tt.enabled, opts)

			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.wantConfigExists {
				var configPath string
				if tt.existingConfig != "" {
					configPath = existingConfigPath
				} else {
					configPath = filepath.Join(tempDir, ".config", "onekeymap", "config.yaml")
				}

				assert.FileExists(t, configPath)

				content, err := os.ReadFile(configPath)
				require.NoError(t, err)

				assert.Equal(t, tt.wantConfig, string(content))
			}
		})
	}
}
