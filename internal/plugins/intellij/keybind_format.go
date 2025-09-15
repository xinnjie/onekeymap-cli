package intellij

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap/keychord"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func parseKeyBinding(ks KeyboardShortcutXML) (*keymap.KeyBinding, error) {
	parts1, err := parseKeyStroke(ks.First)
	if err != nil {
		return nil, err
	}
	first := strings.Join(parts1, "+")
	kb, err := keymap.ParseKeyBinding(first, "+")
	if err != nil {
		return nil, err
	}
	if ks.Second == "" {
		return kb, nil
	}
	parts2, err := parseKeyStroke(ks.Second)
	if err != nil {
		return nil, err
	}
	second := strings.Join(parts2, "+")
	return keymap.ParseKeyBinding(first+" "+second, "+")
}

func formatKeybinding(keybind *keymap.KeyBinding) (*KeyboardShortcutXML, error) {
	// Only support up to two chords for IntelliJ (first and optional second keystroke).
	chords := keybind.GetKeyChords().GetChords()
	if len(chords) > 2 {
		return nil, errors.New("too many chords for intellij, only first two are supported")
	}

	first, err := keyChordToIJKeyStroke(chords[0])
	if err != nil {
		return nil, fmt.Errorf("failed to format first keystroke: %w", err)
	}
	var second string
	if len(chords) == 2 {
		s, err := keyChordToIJKeyStroke(chords[1])
		if err != nil {
			return nil, fmt.Errorf("failed to format second keystroke: %w", err)
		}
		second = s
	}

	return &KeyboardShortcutXML{First: first, Second: second}, nil
}

func keyChordToIJKeyStroke(kc *keymapv1.KeyChord) (string, error) {
	if kc == nil {
		return "", errors.New("invalid key chord: nil")
	}
	// Use platform.PlatformLinus beacuse intellij convert `meta` key to `cmd` on macos internally
	parts, err := keychord.NewKeyChord(kc).Format(platform.PlatformLinux)
	if err != nil {
		return "", err
	}
	s := formatKeyChord(parts)
	if s == "" {
		return "", errors.New("invalid key chord: empty key code")
	}
	return s, nil
}
