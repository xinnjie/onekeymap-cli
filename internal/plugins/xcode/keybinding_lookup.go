package xcode

import (
	"encoding/json"
	"io"

	"github.com/xinnjie/onekeymap-cli/internal/keybindinglookup"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
)

// xcodeKeybindingLookup implements KeybindingLookup interface for Xcode
type xcodeKeybindingLookup struct{}

// NewXcodeKeybindingLookup creates a new KeybindingLookup instance
func NewXcodeKeybindingLookup() keybindinglookup.KeybindingLookup {
	return &xcodeKeybindingLookup{}
}

// Compile-time interface check
var _ keybindinglookup.KeybindingLookup = (*xcodeKeybindingLookup)(nil)

// Lookup implements KeybindingLookup interface
// editorSpecificConfig contains Xcode plist XML configuration
// returns keybindingConfig as JSON strings of matching xcodeKeybinding entries
func (h *xcodeKeybindingLookup) Lookup(
	editorSpecificConfig io.Reader,
	keybinding *keymap.KeyBinding,
) (keybindingConfig []string, err error) {
	// Read and parse the editor-specific config
	configData, err := io.ReadAll(editorSpecificConfig)
	if err != nil {
		return nil, err
	}

	// Parse plist XML using existing parseXcodeConfig function
	plistData, err := parseXcodeConfig(configData)
	if err != nil {
		return nil, err
	}

	xcodeConfig := plistData.MenuKeyBindings.KeyBindings

	// Format the target keybinding to Xcode format for comparison
	// Use existing formatKeybinding function which handles Xcode's format
	targetKey, err := formatKeybinding(keybinding)
	if err != nil {
		return nil, err
	}

	// Find matching keybindings
	var matchingKeybindings []string
	for _, binding := range xcodeConfig {
		if binding.KeyboardShortcut == targetKey {
			// Marshal the matching keybinding to JSON string
			bindingJSON, err := json.Marshal(binding)
			if err != nil {
				return nil, err
			}
			matchingKeybindings = append(matchingKeybindings, string(bindingJSON))
		}
	}

	return matchingKeybindings, nil
}
