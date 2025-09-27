package helix

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns the default path for Helix's config.toml file.
// On macOS, this is typically ~/.config/helix/config.toml.
func (p *helixPlugin) ConfigDetect(opts ...pluginapi.ConfigDetectOption) ([]string, error) {
	options := &pluginapi.ConfigDetectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin": // macOS
		configPath = filepath.Join(home, ".config", "helix", "config.toml")
	default:
		// For now, we only support macOS as requested.
		return nil, errors.New("automatic path discovery is only supported on macOS")
	}

	return []string{configPath}, nil
}
