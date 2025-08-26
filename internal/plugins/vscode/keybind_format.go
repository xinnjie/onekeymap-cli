package vscode

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
)

const vscodeKeyChordSeparator = "+"

func parseKeybinding(keybind string) (*keymap.KeyBinding, error) {
	return keymap.ParseKeyBinding(keybind, vscodeKeyChordSeparator)
}

func formatKeybinding(keybind *keymap.KeyBinding) (string, error) {
	return keybind.Format(platform.PlatformMacOS, vscodeKeyChordSeparator)
}
