package keychord

import (
	"fmt"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// vscode key doc: https://github.com/microsoft/vscode/blob/main/src/vs/base/common/keyCodes.ts

var modifierMap = map[string]keymapv1.KeyModifier{
	"shift": keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
	"ctrl":  keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
	"alt":   keymapv1.KeyModifier_KEY_MODIFIER_ALT,
	// meta is Command(⌘) key on macOS, Windows(⊞) key on Windows, super key on Linux
	"meta": keymapv1.KeyModifier_KEY_MODIFIER_META,
	"cmd":  keymapv1.KeyModifier_KEY_MODIFIER_META,
	"win":  keymapv1.KeyModifier_KEY_MODIFIER_META,
}

// keys except a-z, A-Z, 0-9
// see https://github.com/microsoft/vscode/blob/7ca850c73f7fa37e1fc80090f14268b0b6b504bb/src/vs/base/common/keyCodes.ts#L493
var keys = map[string]struct{}{
	"enter":           {},
	"tab":             {},
	"space":           {},
	"backspace":       {},
	"delete":          {},
	"insert":          {},
	"home":            {},
	"end":             {},
	"pageup":          {},
	"pagedown":        {},
	"escape":          {},
	"f1":              {},
	"f2":              {},
	"f3":              {},
	"f4":              {},
	"f5":              {},
	"f6":              {},
	"f7":              {},
	"f8":              {},
	"f9":              {},
	"f10":             {},
	"f11":             {},
	"f12":             {},
	"up":              {},
	"down":            {},
	"left":            {},
	"right":           {},
	"numpad0":         {},
	"numpad1":         {},
	"numpad2":         {},
	"numpad3":         {},
	"numpad4":         {},
	"numpad5":         {},
	"numpad6":         {},
	"numpad7":         {},
	"numpad8":         {},
	"numpad9":         {},
	"numpad_add":      {},
	"numpad_subtract": {},
	"numpad_multiply": {},
	"numpad_divide":   {},
	"numpad_enter":    {},
	"numpad_decimal":  {},
}

type KeyChord struct {
	*keymapv1.KeyChord
}

func NewKeyChord(protoKeyChord *keymapv1.KeyChord) *KeyChord {
	return &KeyChord{KeyChord: protoKeyChord}
}

// Parse takes a vscode-like keybind string like "ctrl+shift+f" or "ctrl-shift-f" and converts it
// into a structured KeyChord proto message.
func Parse(keybind string, modifierSeparator string) (*KeyChord, error) {
	if keybind == "" {
		return nil, fmt.Errorf("cannot parse empty string")
	}

	lowerKeybind := strings.ToLower(keybind)
	parts := strings.Split(lowerKeybind, modifierSeparator)

	chord := &keymapv1.KeyChord{}

	// The last part is potentially the key code.
	lastKey := parts[len(parts)-1]
	potentialModifiers := parts[:len(parts)-1]

	// Handle cases like "ctrl+alt++" where the key is the separator.
	// In this case, Split results in an empty string at the end.
	if lastKey == "" && strings.HasSuffix(lowerKeybind, modifierSeparator) {
		chord.KeyCode = modifierSeparator
	} else {
		// Check if the last part is a modifier.
		if modifier, ok := modifierMap[lastKey]; ok {
			// It's a modifier. The keycode will be empty.
			chord.Modifiers = append(chord.Modifiers, modifier)
		} else {
			// It's a key code.
			chord.KeyCode = lastKey
		}
	}

	for _, part := range potentialModifiers {
		if part == "" {
			// This can happen with "++" -> ["", ""]. The first part is empty.
			// We've already handled the key, so we just need to parse modifiers.
			continue
		}
		if modifier, ok := modifierMap[part]; ok {
			chord.Modifiers = append(chord.Modifiers, modifier)
		} else {
			// If it's not a modifier, it must be a key code.
			// But we already found a key code (or decided the last part was a modifier).
			// So this is an invalid sequence.
			return nil, fmt.Errorf("invalid key chord string: multiple key codes found ('%s' and '%s')", part, chord.KeyCode)
		}
	}

	// Final validation
	if chord.KeyCode != "" {
		// Validate the key code.
		if len(chord.KeyCode) > 1 {
			if _, ok := keys[chord.KeyCode]; !ok {
				return nil, fmt.Errorf("invalid key code: '%s'", chord.KeyCode)
			}
		}
	} else { // KeyCode is empty
		// Allow exactly one modifier without a key code, e.g. "shift" or "ctrl".
		// Disallow zero or multiple modifiers without a key code.
		if len(chord.Modifiers) != 1 {
			return nil, fmt.Errorf("invalid key chord string: no key code found in '%s'", keybind)
		}
	}

	return &KeyChord{KeyChord: chord}, nil
}

// Format takes a structured KeyChord proto message and returns its canonical
// string representation, e.g., "ctrl+shift+f".
func (kc *KeyChord) Format(p platform.Platform) ([]string, error) {
	if kc == nil || kc.KeyChord == nil {
		return nil, fmt.Errorf("invalid key chord: nil")
	}

	var parts []string
	// Add modifiers in a fixed, canonical order: meta, ctrl, shift, alt.
	if containsModifier(kc.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_META) {
		switch p {
		case platform.PlatformMacOS:
			parts = append(parts, "cmd")
		case platform.PlatformWindows:
			parts = append(parts, "win")
		default:
			parts = append(parts, "meta")
		}
	}
	if containsModifier(kc.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_CTRL) {
		parts = append(parts, "ctrl")
	}
	if containsModifier(kc.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_SHIFT) {
		parts = append(parts, "shift")
	}
	if containsModifier(kc.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_ALT) {
		parts = append(parts, "alt")
	}

	if kc.KeyCode != "" {
		parts = append(parts, kc.KeyCode)
		return parts, nil
	}

	// Allow exactly one modifier without a key code (e.g., ["shift"]).
	if len(kc.Modifiers) == 1 {
		return parts, nil
	}

	return nil, fmt.Errorf("invalid key chord: empty key code")
}

func containsModifier(modifiers []keymapv1.KeyModifier, target keymapv1.KeyModifier) bool {
	for _, m := range modifiers {
		if m == target {
			return true
		}
	}
	return false
}
