package vscode

import (
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

const vscodeKeyChordSeparator = "+"

func ParseKeybinding(keybind string) (*keybinding.Keybinding, error) {
	kb, err := keybinding.NewKeybinding(keybind, keybinding.ParseOption{
		Platform:  platform.PlatformMacOS,
		Separator: vscodeKeyChordSeparator,
	})
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

func FormatKeybinding(keybind *keybinding.Keybinding) (string, error) {
	return keybind.String(keybinding.FormatOption{
		Platform:  platform.PlatformMacOS,
		Separator: vscodeKeyChordSeparator,
	}), nil
}
