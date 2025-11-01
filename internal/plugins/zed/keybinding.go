package zed

import (
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
)

const zedKeyChordSeparator = "-"

func parseZedKeybind(keybind string) (*keymap.KeyBinding, error) {
	return keymap.ParseKeyBinding(keybind, zedKeyChordSeparator)
}

func formatZedKeybind(keybind *keymap.KeyBinding) (string, error) {
	return keybind.Format(platform.PlatformMacOS, zedKeyChordSeparator)
}
