package keymap

import (
	"fmt"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap/keychord"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

const oneKeymapDefaultKeyChordSeparator = "+"

// KeyBinding represents a sequence of one or more key chords that must be
// pressed in order to trigger an action. This allows for multi-key sequences
// like "shift shift" or "ctrl+k ctrl+s".
type KeyBinding struct {
	*keymapv1.KeyBinding
}

func NewKeyBinding(protoKeyBinding *keymapv1.KeyBinding) *KeyBinding {
	return &KeyBinding{KeyBinding: protoKeyBinding}
}

func NewBinding(action, keyChords string) *keymapv1.KeyBinding {
	return &keymapv1.KeyBinding{
		Id:        action,
		KeyChords: MustParseKeyBinding(keyChords).KeyChords,
	}
}

func NewBindingWithComment(action, keyChords, comment string) *keymapv1.KeyBinding {
	return &keymapv1.KeyBinding{
		Id:        action,
		KeyChords: MustParseKeyBinding(keyChords).KeyChords,
		Comment:   comment,
	}
}

func NewBindingWithDescription(action, keyChords, description string) *keymapv1.KeyBinding {
	return &keymapv1.KeyBinding{
		Id:          action,
		KeyChords:   MustParseKeyBinding(keyChords).KeyChords,
		Description: description,
	}
}

// ParseKeyBinding parse from vscode-like keybind into `KeyBinding` struct e.g. ctrl+c
// separator between key chords can be customized.
func ParseKeyBinding(keybind string, modifierSeparator string) (*KeyBinding, error) {
	parts := strings.Split(keybind, " ")
	chords := make([]*keymapv1.KeyChord, 0, len(parts))
	for _, part := range parts {
		kc, err := keychord.Parse(part, modifierSeparator)
		if err != nil {
			return nil, err
		}
		chords = append(chords, kc.KeyChord)
	}
	return NewKeyBinding(&keymapv1.KeyBinding{Id: "", KeyChords: &keymapv1.KeyChordSequence{Chords: chords}}), nil
}

func MustParseKeyBinding(keybind string) *KeyBinding {
	kb, err := ParseKeyBinding(keybind, oneKeymapDefaultKeyChordSeparator)
	if err != nil {
		panic(err)
	}
	return kb
}

// Format formats the key binding into a vscode-like keybind string, e.g. ctrl+c
// separator between key chords can be customized.
func (kb *KeyBinding) Format(p platform.Platform, keyChordSeparator string) (string, error) {
	if kb == nil || len(kb.GetKeyChords().GetChords()) == 0 {
		return "", fmt.Errorf("invalid key binding: empty key chords")
	}
	var parts []string
	for _, protoChord := range kb.GetKeyChords().GetChords() {
		kc := keychord.NewKeyChord(protoChord)
		s, err := kc.Format(p)
		if err != nil {
			return "", err
		}
		parts = append(parts, strings.Join(s, keyChordSeparator))
	}
	return strings.Join(parts, " "), nil
}

func (kb *KeyBinding) String() string {
	// Format the key binding for macOS by default (most common case)
	formattedKeys, err := kb.Format(platform.PlatformMacOS, "+")
	if err != nil {
		// Fallback to empty string if formatting fails
		formattedKeys = ""
	}
	return kb.Id + "|" + formattedKeys
}

func MustFormatKeyBinding(kb *KeyBinding, p platform.Platform) string {
	f, err := kb.Format(p, oneKeymapDefaultKeyChordSeparator)
	if err != nil {
		panic(err)
	}
	return f
}
