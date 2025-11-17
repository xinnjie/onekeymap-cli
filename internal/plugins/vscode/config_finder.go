package vscode

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

var ErrConfigNotFound = errors.New("configuration file not found")

// ConfigDetect returns the default path for VSCode's keybindings.json file.
func (p *vsCodePlugin) ConfigDetect(opt pluginapi2.ConfigDetectOptions) (paths []string, installed bool, err error) {
	return detectConfigForVSCodeVariant("Code", "code", opt)
}

// detectConfigForVSCodeVariant is a helper function to detect config paths for VSCode-based editors.
func detectConfigForVSCodeVariant(
	appDir, commandName string,
	opt pluginapi2.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin": // macOS
		configPath = filepath.Join(home, "Library", "Application Support", appDir, "User", "keybindings.json")
	case "linux":
		configPath = filepath.Join(home, ".config", appDir, "User", "keybindings.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), appDir, "User", "keybindings.json")
	default:
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS, Linux, and Windows, %w",
			pluginapi2.ErrNotSupported,
		)
	}

	if opt.Sandbox {
		installed = false
	} else {
		// Outside of sandbox, `exec.LookPath` is the most reliable way to see if the command is in the user's PATH.
		_, err := exec.LookPath(commandName)
		installed = err == nil
	}

	return []string{configPath}, installed, nil
}
