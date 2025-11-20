package zed

import (
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

const zedKeyChordSeparator = "-"

func ParseZedKeybind(keybind string) (keybinding.Keybinding, error) {
	return keybinding.NewKeybinding(keybind, keybinding.ParseOption{
		Platform:  platform.PlatformMacOS,
		Separator: zedKeyChordSeparator,
	})
}

// FIXME(xinnjie): Format need platform param
func FormatZedKeybind(kb keybinding.Keybinding) (string, error) {
	return kb.String(keybinding.FormatOption{
		Platform:  platform.PlatformMacOS,
		Separator: zedKeyChordSeparator,
	}), nil
}
