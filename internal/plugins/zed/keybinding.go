package zed

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

const zedKeyChordSeparator = "-"

// ParseZedKeybind parses a Zed keybinding string into a Keybinding struct.
// The plat parameter specifies the platform context for parsing modifier keys.
// If plat is empty, it defaults to the current runtime platform.
func ParseZedKeybind(keybind string, plat platform.Platform) (keybinding.Keybinding, error) {
	if plat == "" {
		plat = platform.Current()
	}
	return keybinding.NewKeybinding(keybind, keybinding.ParseOption{
		Platform:  plat,
		Separator: zedKeyChordSeparator,
	})
}

// FormatZedKeybind formats a keybinding for Zed editor.
// The plat parameter specifies the target platform for formatting modifier keys.
// If plat is empty, it defaults to the current runtime platform.
func FormatZedKeybind(kb keybinding.Keybinding, plat platform.Platform) (string, error) {
	if plat == "" {
		plat = platform.Current()
	}
	return kb.String(keybinding.FormatOption{
		Platform:  plat,
		Separator: zedKeyChordSeparator,
	}), nil
}
