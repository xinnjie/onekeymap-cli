package xcode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

// Xcode uses specific symbols for modifier keys:
// @ = Cmd (Command)
// ^ = Ctrl (Control)
// ~ = Alt/Option
// $ = Shift
// No separator between modifier keys and main key

// ParseKeybinding converts Xcode format (@k, ^g, ~@l) directly to keybinding.Keybinding
func ParseKeybinding(keybind string) (keybinding.Keybinding, error) {
	if keybind == "" {
		return keybinding.Keybinding{}, errors.New("empty keybind")
	}

	runes := []rune(keybind)
	var modifiers []keycode.KeyModifier

	// Find the main key (last character that's not a modifier symbol)
	var mainKeyRune rune
	i := len(runes) - 1

	if i < 0 {
		return keybinding.Keybinding{}, errors.New("empty keybind")
	}

	mainKeyRune = runes[i]
	i--

	// Parse modifiers from left to right
	for j := 0; j <= i; j++ {
		switch runes[j] {
		case '@':
			modifiers = append(modifiers, keycode.KeyModifierMeta)
		case '^':
			modifiers = append(modifiers, keycode.KeyModifierCtrl)
		case '~':
			modifiers = append(modifiers, keycode.KeyModifierAlt)
		case '$':
			modifiers = append(modifiers, keycode.KeyModifierShift)
		default:
			return keybinding.Keybinding{}, fmt.Errorf("unknown modifier symbol: %c", runes[j])
		}
	}

	// Determine the key code from the main key rune
	var kc keycode.KeyCode

	// First try to look up special keys
	if keyCode, ok := getKeyCodeFromRune(mainKeyRune); ok {
		kc = keyCode
	} else {
		switch {
		case mainKeyRune >= 'a' && mainKeyRune <= 'z':
			// Lowercase letter
			kc = keycode.KeyCode(string(mainKeyRune))
		case mainKeyRune >= 'A' && mainKeyRune <= 'Z':
			// Uppercase letter (treat as lowercase)
			kc = keycode.KeyCode(strings.ToLower(string(mainKeyRune)))
		case mainKeyRune >= '0' && mainKeyRune <= '9':
			// Digit
			kc = keycode.KeyCode(string(mainKeyRune))
		default:
			return keybinding.Keybinding{}, fmt.Errorf("unsupported key rune: %q", mainKeyRune)
		}
	}

	// Convert to string format for NewKeybinding
	var parts []string
	for _, mod := range modifiers {
		switch mod {
		case keycode.KeyModifierMeta:
			parts = append(parts, "cmd")
		case keycode.KeyModifierCtrl:
			parts = append(parts, "ctrl")
		case keycode.KeyModifierShift:
			parts = append(parts, "shift")
		case keycode.KeyModifierAlt:
			parts = append(parts, "alt")
		}
	}
	parts = append(parts, string(kc))
	keybindStr := strings.Join(parts, "+")

	return keybinding.NewKeybinding(keybindStr, keybinding.ParseOption{
		Platform:  platform.PlatformMacOS,
		Separator: "+",
	})
}

// FormatKeybinding converts internal format back to Xcode format
func FormatKeybinding(keybind keybinding.Keybinding) (string, error) {
	if len(keybind.KeyChords) == 0 {
		return "", errors.New("invalid key binding: empty key chords")
	}

	if len(keybind.KeyChords) != 1 {
		return "", fmt.Errorf(
			"xcode doesn't support multi-key-chords keybinding: %v",
			keybind.String(keybinding.FormatOption{Separator: "+"}),
		)
	}

	chord := keybind.KeyChords[0]

	// Handle modifier-only chord (exactly one modifier, no keycode)
	if chord.KeyCode == "" {
		if len(chord.Modifiers) != 1 {
			return "", errors.New("invalid key chord: empty key code")
		}
		switch chord.Modifiers[0] {
		case keycode.KeyModifierMeta:
			return "cmd", nil
		case keycode.KeyModifierCtrl:
			return "ctrl", nil
		case keycode.KeyModifierShift:
			return "shift", nil
		case keycode.KeyModifierAlt:
			return "alt", nil
		default:
			return "", errors.New("invalid key chord: unknown modifier")
		}
	}

	var b strings.Builder

	has := func(mod keycode.KeyModifier) bool {
		for _, m := range chord.Modifiers {
			if m == mod {
				return true
			}
		}
		return false
	}

	if has(keycode.KeyModifierMeta) {
		b.WriteRune('@')
	}
	if has(keycode.KeyModifierCtrl) {
		b.WriteRune('^')
	}
	if has(keycode.KeyModifierShift) {
		b.WriteRune('$')
	}
	if has(keycode.KeyModifierAlt) {
		b.WriteRune('~')
	}

	// Convert KeyCode directly to Xcode rune
	r, ok := getRuneFromKeyCode(chord.KeyCode)
	if !ok {
		return "", fmt.Errorf("unsupported key code for Xcode: %v", chord.KeyCode)
	}
	b.WriteRune(r)

	return b.String(), nil
}
