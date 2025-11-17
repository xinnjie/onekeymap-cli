package xcode

import (
	"os"
	"path/filepath"

	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// detectXcodeConfig returns the default path to Xcode's keybinding configuration files.
func detectXcodeConfig(opts pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	// Xcode keybinding files are stored in:
	// ~/Library/Developer/Xcode/UserData/KeyBindings/
	keybindingsDir := filepath.Join(homeDir, "Library", "Developer", "Xcode", "UserData", "KeyBindings")

	var configPaths []string

	// Check if the KeyBindings directory exists
	if info, err := os.Stat(keybindingsDir); err == nil && info.IsDir() {
		// Look for .idekeybindings files
		entries, err := os.ReadDir(keybindingsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && filepath.Ext(entry.Name()) == ".idekeybindings" {
					configPaths = append(configPaths, filepath.Join(keybindingsDir, entry.Name()))
				}
			}
		}
	}

	// If no custom keybindings found, suggest the default location for a new file
	if len(configPaths) == 0 {
		// Ensure the directory exists
		if err := os.MkdirAll(keybindingsDir, 0750); err == nil {
			defaultPath := filepath.Join(keybindingsDir, "Default.idekeybindings")
			configPaths = append(configPaths, defaultPath)
		}
	}

	// Determine if Xcode is installed
	if opts.Sandbox {
		// In sandbox mode, we cannot reliably detect installation status
		installed = false
	} else {
		// Check if Xcode is installed by looking for the application bundle
		xcodeApp := "/Applications/Xcode.app"
		if _, err := os.Stat(xcodeApp); os.IsNotExist(err) {
			return nil, false, nil
		}
		installed = true
	}

	return configPaths, installed, nil
}
