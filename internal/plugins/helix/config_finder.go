package helix

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns the default path for Helix's config.toml file.
// On macOS, this is typically ~/.config/helix/config.toml.
func (p *helixPlugin) ConfigDetect(opt pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin": // macOS
		configPath = filepath.Join(home, ".config", "helix", "config.toml")
	default:
		// For now, we only support macOS as requested.
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, %w",
			pluginapi.ErrNotSupported,
		)
	}

	_, err = exec.LookPath("hx")
	installed = err == nil

	return []string{configPath}, installed, nil
}
