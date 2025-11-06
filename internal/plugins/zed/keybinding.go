package zed

import (
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
)

const zedKeyChordSeparator = "-"

func ParseZedKeybind(keybind string) (*keymap.KeyBinding, error) {
	return keymap.ParseKeyBinding(keybind, zedKeyChordSeparator)
}

func FormatZedKeybind(keybind *keymap.KeyBinding) (string, error) {
	return keybind.Format(platform.PlatformMacOS, zedKeyChordSeparator)
}
