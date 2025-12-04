package vscode

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

const vscodeKeyChordSeparator = "+"

// ParseKeybinding parses a VSCode keybinding string into a Keybinding struct.
// The plat parameter specifies the platform context for parsing modifier keys.
// If plat is empty, it defaults to the current runtime platform.
func ParseKeybinding(keybind string, plat platform.Platform) (*keybinding.Keybinding, error) {
	if plat == "" {
		plat = platform.Current()
	}
	kb, err := keybinding.NewKeybinding(keybind, keybinding.ParseOption{
		Platform:  plat,
		Separator: vscodeKeyChordSeparator,
	})
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// FormatKeybinding formats a Keybinding struct into a VSCode keybinding string.
// The plat parameter specifies the target platform for formatting modifier keys.
// If plat is empty, it defaults to the current runtime platform.
func FormatKeybinding(keybind *keybinding.Keybinding, plat platform.Platform) (string, error) {
	if plat == "" {
		plat = platform.Current()
	}
	return keybind.String(keybinding.FormatOption{
		Platform:  plat,
		Separator: vscodeKeyChordSeparator,
	}), nil
}
