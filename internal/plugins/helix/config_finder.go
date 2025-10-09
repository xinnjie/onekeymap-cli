package helix

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns the default path for Helix's config.toml file.
// On macOS, this is typically ~/.config/helix/config.toml.
func (p *helixPlugin) ConfigDetect(_ pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = filepath.Join(home, ".config", "helix", "config.toml")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return nil, false, fmt.Errorf("APPDATA environment variable not set, %w", pluginapi.ErrNotSupported)
		}
		configPath = filepath.Join(appData, "helix", "config.toml")
	default:
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, Linux, and Windows, %w",
			pluginapi.ErrNotSupported,
		)
	}

	_, err = exec.LookPath("hx")
	installed = err == nil

	return []string{configPath}, installed, nil
}
