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
	*keymapv1.Binding
}

func NewKeyBinding(binding *keymapv1.Binding) *KeyBinding {
	return &KeyBinding{Binding: binding}
}

// newBindingProto creates a Binding proto from a vscode-like key sequence string.
func newBindingProto(keyChords ...string) []*keymapv1.Binding {
	var bindings []*keymapv1.Binding
	for _, keyChord := range keyChords {
		bindings = append(bindings, &keymapv1.Binding{
			KeyChords: MustParseKeyBinding(keyChord).KeyChords,
		})
	}
	return bindings
}

// NewActioinBinding creates an ActionBinding with a single Binding.
func NewActioinBinding(action string, keyChords ...string) *keymapv1.ActionBinding {
	return &keymapv1.ActionBinding{
		Id:       action,
		Bindings: newBindingProto(keyChords...),
	}
}

// NewActionBindingWithComment creates an ActionBinding with one Binding and a comment.
func NewActionBindingWithComment(action, keyChords, comment string) *keymapv1.ActionBinding {
	ab := NewActioinBinding(action, keyChords)
	ab.Comment = comment
	return ab
}

// NewActionBindingWithDescription creates an ActionBinding with one Binding and a description.
func NewActionBindingWithDescription(action, keyChords, description string) *keymapv1.ActionBinding {
	ab := NewActioinBinding(action, keyChords)
	ab.Description = description
	return ab
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
	return NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{Chords: chords}}), nil
}

// Parse a vscode-like keybind string, e.g. ctrl+c to KeyBinding
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

func MustFormatKeyBinding(kb *KeyBinding, p platform.Platform) string {
	f, err := kb.Format(p, oneKeymapDefaultKeyChordSeparator)
	if err != nil {
		panic(err)
	}
	return f
}
