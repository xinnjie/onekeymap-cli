package keychord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

//nolint:gochecknoglobals // modifier lookup table is initialized once and used read-only at runtime
var modifierMap = map[string]keycode.KeyModifier{
	"shift": keycode.KeyModifierShift,
	"ctrl":  keycode.KeyModifierCtrl,
	"alt":   keycode.KeyModifierAlt,
	// meta is Command(⌘) key on macOS, Windows(⊞) key on Windows, super key on Linux
	"meta": keycode.KeyModifierMeta,
	"cmd":  keycode.KeyModifierMeta,
	"win":  keycode.KeyModifierMeta,
}

//nolint:gochecknoglobals // validKeyCodes set is initialized once and used read-only at runtime
var validKeyCodes = map[string]struct{}{
	"a": {}, "b": {}, "c": {}, "d": {}, "e": {}, "f": {}, "g": {}, "h": {}, "i": {}, "j": {}, "k": {}, "l": {}, "m": {}, "n": {}, "o": {}, "p": {}, "q": {}, "r": {}, "s": {}, "t": {}, "u": {}, "v": {}, "w": {}, "x": {}, "y": {}, "z": {},
	"0": {}, "1": {}, "2": {}, "3": {}, "4": {}, "5": {}, "6": {}, "7": {}, "8": {}, "9": {},
	"capslock": {}, "shift": {}, "fn": {}, "ctrl": {}, "alt": {}, "cmd": {}, "rightcmd": {}, "rightalt": {}, "rightctrl": {}, "rightshift": {}, "enter": {}, "\\": {}, "`": {}, ",": {}, "=": {}, "-": {}, "+": {}, ".": {}, "'": {}, ";": {}, "/": {}, "space": {}, "tab": {}, "[": {}, "]": {},
	"pageup": {}, "pagedown": {}, "home": {}, "end": {}, "up": {}, "right": {}, "down": {}, "left": {}, "escape": {}, "backspace": {}, "delete": {}, "insert": {}, "mute": {}, "volumeup": {}, "volumedown": {},
	"f1": {}, "f2": {}, "f3": {}, "f4": {}, "f5": {}, "f6": {}, "f7": {}, "f8": {}, "f9": {}, "f10": {}, "f11": {}, "f12": {}, "f13": {}, "f14": {}, "f15": {}, "f16": {}, "f17": {}, "f18": {}, "f19": {}, "f20": {},
	"numpad0": {}, "numpad1": {}, "numpad2": {}, "numpad3": {}, "numpad4": {}, "numpad5": {}, "numpad6": {}, "numpad7": {}, "numpad8": {}, "numpad9": {},
	"numpad_clear": {}, "numpad_decimal": {}, "numpad_divide": {}, "numpad_enter": {}, "numpad_equals": {}, "numpad_subtract": {}, "numpad_multiply": {}, "numpad_add": {},
}

type KeyChord struct {
	Modifiers []keycode.KeyModifier
	KeyCode   keycode.KeyCode
}

// NewKeyChord parse takes a vscode-like keybind string like "ctrl+shift+f" or "ctrl-shift-f" and converts it
// into a structured KeyChord.
func NewKeyChord(keychord string, opt ParseOption) (KeyChord, error) {
	if keychord == "" {
		return KeyChord{}, errors.New("cannot parse empty string")
	}

	lowerKeybind := strings.ToLower(keychord)
	separator := opt.Separator
	if separator == "" {
		separator = "+"
	}

	parts := strings.Split(lowerKeybind, separator)

	chord := KeyChord{}

	// The last part is potentially the key code.
	lastKey := parts[len(parts)-1]
	potentialModifiers := parts[:len(parts)-1]

	// Handle cases like "ctrl+alt++" where the key is the separator.
	// In this case, Split results in an empty string at the end.
	if lastKey == "" && strings.HasSuffix(lowerKeybind, separator) {
		if err := handleSeparatorAsKey(&chord, separator); err != nil {
			return KeyChord{}, err
		}
	} else {
		if err := handleLastPart(&chord, lastKey); err != nil {
			return KeyChord{}, err
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
			return KeyChord{}, fmt.Errorf("invalid key chord string: multiple key codes found ('%s' and '%s')", part, chord.KeyCode)
		}
	}

	// Final validation
	if chord.KeyCode == "" {
		// Allow exactly one modifier without a key code, e.g. "shift" or "ctrl".
		// Disallow zero or multiple modifiers without a key code.
		if len(chord.Modifiers) != 1 {
			return KeyChord{}, fmt.Errorf("invalid key chord string: no key code found in '%s'", keychord)
		}
	}

	return chord, nil
}

type ParseOption struct {
	Platform  platform.Platform
	Separator string
}

// String returns the string representation of the key chord. like "ctrl+a"
func (kc KeyChord) String(opt FormatOption) string {
	var parts []string
	// Add modifiers in a fixed, canonical order: meta, ctrl, shift, alt.
	if containsModifier(kc.Modifiers, keycode.KeyModifierMeta) {
		switch opt.Platform {
		case platform.PlatformMacOS:
			parts = append(parts, "cmd")
		case platform.PlatformWindows:
			parts = append(parts, "win")
		default:
			parts = append(parts, "meta")
		}
	}
	if containsModifier(kc.Modifiers, keycode.KeyModifierCtrl) {
		parts = append(parts, "ctrl")
	}
	if containsModifier(kc.Modifiers, keycode.KeyModifierShift) {
		parts = append(parts, "shift")
	}
	if containsModifier(kc.Modifiers, keycode.KeyModifierAlt) {
		parts = append(parts, "alt")
	}

	if kc.KeyCode != "" {
		// Since KeyCode is string, we just append it.
		parts = append(parts, string(kc.KeyCode))
		return strings.Join(parts, opt.Separator)
	}

	// Allow exactly one modifier without a key code (e.g., ["shift"]).
	if len(kc.Modifiers) == 1 {
		return strings.Join(parts, opt.Separator)
	}

	return ""
}

type FormatOption struct {
	Platform  platform.Platform
	Separator string
}

func handleSeparatorAsKey(chord *KeyChord, modifierSeparator string) error {
	if _, ok := validKeyCodes[modifierSeparator]; ok {
		chord.KeyCode = keycode.KeyCode(modifierSeparator)
		return nil
	}
	return fmt.Errorf("invalid key code: '%s'", modifierSeparator)
}

func handleLastPart(chord *KeyChord, lastKey string) error {
	// Check if the last part is a modifier.
	if modifier, ok := modifierMap[lastKey]; ok {
		// It's a modifier. The keycode will be empty.
		chord.Modifiers = append(chord.Modifiers, modifier)
		return nil
	}
	// It's a key code.
	if _, ok := validKeyCodes[lastKey]; ok {
		chord.KeyCode = keycode.KeyCode(lastKey)
		return nil
	}
	return fmt.Errorf("invalid key code: '%s'", lastKey)
}

func containsModifier(modifiers []keycode.KeyModifier, target keycode.KeyModifier) bool {
	for _, m := range modifiers {
		if m == target {
			return true
		}
	}
	return false
}
