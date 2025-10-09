package zed

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

var ErrConfigNotFound = errors.New("configuration file not found")

// ConfigDetect returns the default path for Zed's keymap.json file.
// On macOS, this is typically ~/.config/zed/keymap.json.
func (p *zedPlugin) ConfigDetect(_ pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = filepath.Join(home, ".config", "zed", "keymap.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Zed", "keymap.json")
	default:
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, Linux, and Windows, %w",
			pluginapi.ErrNotSupported,
		)
	}

	_, err = exec.LookPath("zed")
	installed = err == nil

	return []string{configPath}, installed, nil
}
