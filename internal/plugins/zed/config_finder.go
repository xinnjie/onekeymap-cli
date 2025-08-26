package zed

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

var ErrConfigNotFound = errors.New("configuration file not found")

// DefaultConfigPath returns the default path for Zed's keymap.json file.
// On macOS, this is typically ~/.config/zed/keymap.json.
func (p *zedPlugin) DefaultConfigPath() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin": // macOS
		configPath = filepath.Join(home, ".config", "zed", "keymap.json")
	default:
		// For now, we only support macOS as requested.
		return nil, errors.New("automatic path discovery is only supported on macOS")
	}

	return []string{configPath}, nil
}
