package xcode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// Xcode uses specific symbols for modifier keys:
// @ = Cmd (Command)
// ^ = Ctrl (Control)
// ~ = Alt/Option
// $ = Shift
// No separator between modifier keys and main key

// ParseKeybinding converts Xcode format (@k, ^g, ~@l) directly to keymap.KeyBinding
func ParseKeybinding(keybind string) (*keymap.KeyBinding, error) {
	if keybind == "" {
		return nil, errors.New("empty keybind")
	}

	runes := []rune(keybind)
	var modifiers []keymapv1.KeyModifier

	// Find the main key (last character that's not a modifier symbol)
	var mainKeyRune rune
	i := len(runes) - 1

	if i < 0 {
		return nil, errors.New("empty keybind")
	}

	mainKeyRune = runes[i]
	i--

	// Parse modifiers from left to right
	for j := 0; j <= i; j++ {
		switch runes[j] {
		case '@':
			modifiers = append(modifiers, keymapv1.KeyModifier_KEY_MODIFIER_META)
		case '^':
			modifiers = append(modifiers, keymapv1.KeyModifier_KEY_MODIFIER_CTRL)
		case '~':
			modifiers = append(modifiers, keymapv1.KeyModifier_KEY_MODIFIER_ALT)
		case '$':
			modifiers = append(modifiers, keymapv1.KeyModifier_KEY_MODIFIER_SHIFT)
		default:
			return nil, fmt.Errorf("unknown modifier symbol: %c", runes[j])
		}
	}

	// Determine the key code from the main key rune
	var keyCode keymapv1.KeyCode

	// First try to look up special keys
	if kc, ok := getKeyCodeFromRune(mainKeyRune); ok {
		keyCode = kc
	} else {
		switch {
		case mainKeyRune >= 'a' && mainKeyRune <= 'z':
			// Lowercase letter
			keyCode = keymapv1.KeyCode_A + keymapv1.KeyCode(mainKeyRune-'a')
		case mainKeyRune >= 'A' && mainKeyRune <= 'Z':
			// Uppercase letter (treat as lowercase)
			keyCode = keymapv1.KeyCode_A + keymapv1.KeyCode(mainKeyRune-'A')
		case mainKeyRune >= '0' && mainKeyRune <= '9':
			// Digit
			keyCode = keymapv1.KeyCode_DIGIT_0 + keymapv1.KeyCode(mainKeyRune-'0')
		default:
			return nil, fmt.Errorf("unsupported key rune: %q", mainKeyRune)
		}
	}

	chord := &keymapv1.KeyChord{
		KeyCode:   keyCode,
		Modifiers: modifiers,
	}

	return keymap.NewKeyBinding(&keymapv1.KeybindingReadable{
		KeyChords: &keymapv1.Keybinding{
			Chords: []*keymapv1.KeyChord{chord},
		},
	}), nil
}

// FormatKeybinding converts internal format back to Xcode format
func FormatKeybinding(keybind *keymap.KeyBinding) (string, error) {
	if keybind == nil || keybind.GetKeyChords() == nil || len(keybind.GetKeyChords().GetChords()) == 0 {
		return "", errors.New("invalid key binding: empty key chords")
	}

	chords := keybind.GetKeyChords().GetChords()
	if len(chords) != 1 {
		return "", errors.New("xcode doesn't support key chords")
	}

	chord := chords[0]

	// Handle modifier-only chord (exactly one modifier, no keycode)
	if chord.GetKeyCode() == keymapv1.KeyCode_KEY_CODE_UNSPECIFIED {
		if len(chord.GetModifiers()) != 1 {
			return "", errors.New("invalid key chord: empty key code")
		}
		switch chord.GetModifiers()[0] {
		case keymapv1.KeyModifier_KEY_MODIFIER_META:
			return "cmd", nil
		case keymapv1.KeyModifier_KEY_MODIFIER_CTRL:
			return "ctrl", nil
		case keymapv1.KeyModifier_KEY_MODIFIER_SHIFT:
			return "shift", nil
		case keymapv1.KeyModifier_KEY_MODIFIER_ALT:
			return "alt", nil
		default:
			return "", errors.New("invalid key chord: unknown modifier")
		}
	}

	var b strings.Builder

	has := func(mod keymapv1.KeyModifier) bool {
		for _, m := range chord.GetModifiers() {
			if m == mod {
				return true
			}
		}
		return false
	}

	if has(keymapv1.KeyModifier_KEY_MODIFIER_META) {
		b.WriteRune('@')
	}
	if has(keymapv1.KeyModifier_KEY_MODIFIER_CTRL) {
		b.WriteRune('^')
	}
	if has(keymapv1.KeyModifier_KEY_MODIFIER_SHIFT) {
		b.WriteRune('$')
	}
	if has(keymapv1.KeyModifier_KEY_MODIFIER_ALT) {
		b.WriteRune('~')
	}

	// Convert KeyCode directly to Xcode rune
	r, ok := getRuneFromKeyCode(chord.GetKeyCode())
	if !ok {
		return "", fmt.Errorf("unsupported key code for Xcode: %v", chord.GetKeyCode())
	}
	b.WriteRune(r)

	return b.String(), nil
}
