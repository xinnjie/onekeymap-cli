package demo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns a reasonable default config path for the demo plugin.
// It does not correspond to a real editor, but provides a stable location for testing.
func (p *demoPlugin) ConfigDetect(opt pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	var configPath string
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = filepath.Join(home, ".config", "onekeymap", "demo.keybindings.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "onekeymap", "demo.keybindings.json")
	default:
		return nil, false, errors.New("unsupported operating system")
	}

	return []string{configPath}, true, nil
}
