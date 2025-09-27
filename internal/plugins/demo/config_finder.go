package demo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

// DefaultConfigPath returns a reasonable default config path for the demo plugin.
// It does not correspond to a real editor, but provides a stable location for testing.
func (p *demoPlugin) DefaultConfigPath(opts ...pluginapi.DefaultConfigPathOption) ([]string, error) {
	options := &pluginapi.DefaultConfigPathOptions{}
	for _, opt := range opts {
		opt(options)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = filepath.Join(home, ".config", "onekeymap", "demo.keybindings.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "onekeymap", "demo.keybindings.json")
	default:
		return nil, errors.New("unsupported operating system")
	}

	return []string{configPath}, nil
}
