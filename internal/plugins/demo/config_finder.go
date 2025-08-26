package demo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultConfigPath returns a reasonable default config path for the demo plugin.
// It does not correspond to a real editor, but provides a stable location for testing.
func (p *demoPlugin) DefaultConfigPath() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		return []string{filepath.Join(home, ".config", "onekeymap", "demo.keybindings.json")}, nil
	case "windows":
		return []string{filepath.Join(os.Getenv("APPDATA"), "onekeymap", "demo.keybindings.json")}, nil
	default:
		return nil, errors.New("unsupported operating system")
	}
}
