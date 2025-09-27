package vscode

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var ErrConfigNotFound = errors.New("configuration file not found")

// ConfigDetect returns the default path for VSCode's keybindings.json file.
func (p *vsCodePlugin) ConfigDetect(opts ...pluginapi.ConfigDetectOption) ([]string, error) {
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
		configPath = filepath.Join(home, "Library", "Application Support", "Code", "User", "keybindings.json")
	case "linux":
		configPath = filepath.Join(home, ".config", "Code", "User", "keybindings.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Code", "User", "keybindings.json")
	default:
		return nil, errors.New("unsupported operating system")
	}

	return []string{configPath}, nil
}
