package vscode

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/keybindinglookup"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
)

// vscodeKeybindingLookup implements KeybindingLookup interface for VSCode
type vscodeKeybindingLookup struct{}

// NewVSCodeKeybindingLookup creates a new KeybindingLookup instance
func NewVSCodeKeybindingLookup() keybindinglookup.KeybindingLookup {
	return &vscodeKeybindingLookup{}
}

// Compile-time interface check
var _ keybindinglookup.KeybindingLookup = (*vscodeKeybindingLookup)(nil)

// Lookup implements KeybindingLookup interface
// editorSpecificConfig contains VSCodeKeybindingConfig JSON
// returns keybindingConfig as JSON strings of matching vscodeKeybinding entries
func (h *vscodeKeybindingLookup) Lookup(
	editorSpecificConfig io.Reader,
	keybinding *keymap.KeyBinding,
) (keybindingConfig []string, err error) {
	// Read and parse the editor-specific config
	configData, err := io.ReadAll(editorSpecificConfig)
	if err != nil {
		return nil, err
	}

	// Remove comments (lines starting with //)
	lines := strings.Split(string(configData), "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanJSON := strings.Join(cleanLines, "\n")

	var vscodeConfig vscodeKeybindingConfig
	if err := json.Unmarshal([]byte(cleanJSON), &vscodeConfig); err != nil {
		return nil, err
	}

	// Format the target keybinding to VSCode format for comparison
	targetKey, err := keybinding.Format(platform.PlatformMacOS, "+")
	if err != nil {
		return nil, err
	}

	// Find matching keybindings
	var matchingKeybindings []string
	for _, binding := range vscodeConfig {
		if binding.Key == targetKey {
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
