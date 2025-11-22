package intellij

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

const maxIntellijChords = 2

func ParseKeyBinding(ks KeyboardShortcutXML) (keybinding.Keybinding, error) {
	parts1, err := parseKeyStroke(ks.First)
	if err != nil {
		return keybinding.Keybinding{}, err
	}
	first := strings.Join(parts1, "+")
	var keybindStr string
	if ks.Second == "" {
		keybindStr = first
	} else {
		parts2, err := parseKeyStroke(ks.Second)
		if err != nil {
			return keybinding.Keybinding{}, err
		}
		second := strings.Join(parts2, "+")
		keybindStr = first + " " + second
	}
	return keybinding.NewKeybinding(keybindStr, keybinding.ParseOption{
		Platform:  platform.PlatformLinux,
		Separator: "+",
	})
}

func FormatKeybinding(keybind keybinding.Keybinding) (*KeyboardShortcutXML, error) {
	// Only support up to two chords for IntelliJ (first and optional second keystroke).
	if len(keybind.KeyChords) > maxIntellijChords {
		return nil, errors.New("too many chords for intellij, only first two are supported")
	}

	first, err := keyChordToIJKeyStroke(keybind.KeyChords[0])
	if err != nil {
		return nil, fmt.Errorf("failed to format first keystroke: %w", err)
	}
	var second string
	if len(keybind.KeyChords) == maxIntellijChords {
		s, err := keyChordToIJKeyStroke(keybind.KeyChords[1])
		if err != nil {
			return nil, fmt.Errorf("failed to format second keystroke: %w", err)
		}
		second = s
	}

	return &KeyboardShortcutXML{First: first, Second: second}, nil
}

func keyChordToIJKeyStroke(kc keychord.KeyChord) (string, error) {
	if len(kc.Modifiers) == 0 && kc.KeyCode == "" {
		return "", errors.New("invalid key chord: empty")
	}
	// Use platform.PlatformLinux because intellij convert `meta` key to `cmd` on macos internally
	parts := kc.String(keychord.FormatOption{
		Platform:  platform.PlatformLinux,
		Separator: "+",
	})
	s := formatKeyChord(strings.Split(parts, "+"))
	if s == "" {
		return "", errors.New("invalid key chord: empty key code")
	}
	return s, nil
}
