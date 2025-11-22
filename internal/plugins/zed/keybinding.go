package zed

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

const zedKeyChordSeparator = "-"

func ParseZedKeybind(keybind string) (keybinding.Keybinding, error) {
	return keybinding.NewKeybinding(keybind, keybinding.ParseOption{
		Platform:  platform.PlatformMacOS,
		Separator: zedKeyChordSeparator,
	})
}

// FormatZedKeybind formats a keybinding for Zed editor. FIXME(xinnjie): Format need platform param
func FormatZedKeybind(kb keybinding.Keybinding) (string, error) {
	return kb.String(keybinding.FormatOption{
		Platform:  platform.PlatformMacOS,
		Separator: zedKeyChordSeparator,
	}), nil
}
