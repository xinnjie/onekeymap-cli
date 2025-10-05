package vscode

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

// ConfigDetect returns the default path for VSCode's keybindings.json file.
func (p *vsCodePlugin) ConfigDetect(opt pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin": // macOS
		configPath = filepath.Join(home, "Library", "Application Support", "Code", "User", "keybindings.json")
	case "linux":
		configPath = filepath.Join(home, ".config", "Code", "User", "keybindings.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Code", "User", "keybindings.json")
	default:
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, %w",
			pluginapi.ErrNotSupported,
		)
	}

	if opt.Sandbox {
		installed = false
	} else {
		// Outside of sandbox, `exec.LookPath` is the most reliable way to see if `code` is in the user's PATH.
		_, err := exec.LookPath("code")
		installed = err == nil
	}

	return []string{configPath}, installed, nil
}
