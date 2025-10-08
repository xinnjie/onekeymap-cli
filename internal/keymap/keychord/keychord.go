package keychord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// vscode key doc: https://github.com/microsoft/vscode/blob/main/src/vs/base/common/keyCodes.ts

//nolint:gochecknoglobals // modifier lookup table is initialized once and used read-only at runtime
var modifierMap = map[string]keymapv1.KeyModifier{
	"shift": keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
	"ctrl":  keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
	"alt":   keymapv1.KeyModifier_KEY_MODIFIER_ALT,
	// meta is Command(⌘) key on macOS, Windows(⊞) key on Windows, super key on Linux
	"meta": keymapv1.KeyModifier_KEY_MODIFIER_META,
	"cmd":  keymapv1.KeyModifier_KEY_MODIFIER_META,
	"win":  keymapv1.KeyModifier_KEY_MODIFIER_META,
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
		return nil, errors.New("cannot parse empty string")
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
		if kc, ok := keycode.FromString(modifierSeparator); ok {
			chord.KeyCode = kc
		} else {
			return nil, fmt.Errorf("invalid key code: '%s'", modifierSeparator)
		}
	} else {
		// Check if the last part is a modifier.
		if modifier, ok := modifierMap[lastKey]; ok {
			// It's a modifier. The keycode will be empty.
			chord.Modifiers = append(chord.Modifiers, modifier)
		} else {
			// It's a key code.
			if kc, ok := keycode.FromString(lastKey); ok {
				chord.KeyCode = kc
			} else {
				return nil, fmt.Errorf("invalid key code: '%s'", lastKey)
			}
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
			keyStr, _ := keycode.ToString(chord.GetKeyCode())
			return nil, fmt.Errorf("invalid key chord string: multiple key codes found ('%s' and '%s')", part, keyStr)
		}
	}

	// Final validation
	if chord.GetKeyCode() == keymapv1.KeyCode_KEY_CODE_UNSPECIFIED {
		// Allow exactly one modifier without a key code, e.g. "shift" or "ctrl".
		// Disallow zero or multiple modifiers without a key code.
		if len(chord.GetModifiers()) != 1 {
			return nil, fmt.Errorf("invalid key chord string: no key code found in '%s'", keybind)
		}
	}

	return &KeyChord{KeyChord: chord}, nil
}

// Format takes a structured KeyChord proto message and returns its canonical
// string representation, e.g., "ctrl+shift+f".
func (kc *KeyChord) Format(p platform.Platform) ([]string, error) {
	if kc == nil || kc.KeyChord == nil {
		return nil, errors.New("invalid key chord: nil")
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

	if kc.KeyCode != keymapv1.KeyCode_KEY_CODE_UNSPECIFIED {
		if keyStr, ok := keycode.ToString(kc.KeyCode); ok {
			parts = append(parts, keyStr)
			return parts, nil
		}
		return nil, fmt.Errorf("invalid key code: %v", kc.KeyCode)
	}

	// Allow exactly one modifier without a key code (e.g., ["shift"]).
	if len(kc.Modifiers) == 1 {
		return parts, nil
	}

	return nil, errors.New("invalid key chord: empty key code")
}

func containsModifier(modifiers []keymapv1.KeyModifier, target keymapv1.KeyModifier) bool {
	for _, m := range modifiers {
		if m == target {
			return true
		}
	}
	return false
}
