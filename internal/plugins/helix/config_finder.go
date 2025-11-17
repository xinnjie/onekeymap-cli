package helix

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// ConfigDetect returns the default path for Helix's config.toml file.
// On macOS, this is typically ~/.config/helix/config.toml.
func (p *helixPlugin) ConfigDetect(_ pluginapi2.ConfigDetectOptions) (paths []string, installed bool, err error) {
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
			return nil, false, fmt.Errorf("APPDATA environment variable not set, %w", pluginapi2.ErrNotSupported)
		}
		configPath = filepath.Join(appData, "helix", "config.toml")
	default:
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, Linux, and Windows, %w",
			pluginapi2.ErrNotSupported,
		)
	}

	_, err = exec.LookPath("hx")
	installed = err == nil

	return []string{configPath}, installed, nil
}
