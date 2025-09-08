package helix

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/bimap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap/keycode"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var (
	helixModifierMapping = bimap.NewBiMapFromMap(map[string]keymapv1.KeyModifier{
		"C": keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
		"A": keymapv1.KeyModifier_KEY_MODIFIER_ALT,
		"S": keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
		"M": keymapv1.KeyModifier_KEY_MODIFIER_META,
	})

	// see https://github.com/helix-editor/helix/blob/22a3b10dd8ab907367ae1fe57d9703e22b30d391/book/src/remapping.md?plain=1#L92
	helixKeyMapping = bimap.NewBiMapFromMap(map[string]keymapv1.KeyCode{
		// Special keys from Helix docs
		"backspace": keymapv1.KeyCode_DELETE, // Note: Helix 'backspace' is our 'delete' (backwards delete)
		"space":     keymapv1.KeyCode_SPACE,
		"ret":       keymapv1.KeyCode_RETURN,
		"esc":       keymapv1.KeyCode_ESCAPE,
		"del":       keymapv1.KeyCode_FORWARD_DELETE, // Note: Helix 'del' is our 'forward_delete'
		"ins":       keymapv1.KeyCode_INSERT,
		"left":      keymapv1.KeyCode_LEFT_ARROW,
		"right":     keymapv1.KeyCode_RIGHT_ARROW,
		"up":        keymapv1.KeyCode_UP_ARROW,
		"down":      keymapv1.KeyCode_DOWN_ARROW,
		"home":      keymapv1.KeyCode_HOME,
		"end":       keymapv1.KeyCode_END,
		"pageup":    keymapv1.KeyCode_PAGE_UP,
		"pagedown":  keymapv1.KeyCode_PAGE_DOWN,
		"tab":       keymapv1.KeyCode_TAB,
		// TODO(xinnjie) shift+, -> <
		"lt":    keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
		"gt":    keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
		"minus": keymapv1.KeyCode_MINUS,

		// Function keys
		"F1":  keymapv1.KeyCode_F1,
		"F2":  keymapv1.KeyCode_F2,
		"F3":  keymapv1.KeyCode_F3,
		"F4":  keymapv1.KeyCode_F4,
		"F5":  keymapv1.KeyCode_F5,
		"F6":  keymapv1.KeyCode_F6,
		"F7":  keymapv1.KeyCode_F7,
		"F8":  keymapv1.KeyCode_F8,
		"F9":  keymapv1.KeyCode_F9,
		"F10": keymapv1.KeyCode_F10,
		"F11": keymapv1.KeyCode_F11,
		"F12": keymapv1.KeyCode_F12,
	})
)

func formatKeybinding(kb *keymap.KeyBinding) (string, error) {
	var outChords []string

	for _, chord := range kb.GetKeyChords().GetChords() {
		if keycode.IsNumpad(chord.KeyCode) {
			return "", ErrNotSupportKeyChords
		}

		// Special case for C--
		if chord.KeyCode == keymapv1.KeyCode_MINUS &&
			len(chord.Modifiers) == 1 &&
			chord.Modifiers[0] == keymapv1.KeyModifier_KEY_MODIFIER_CTRL {
			outChords = append(outChords, "C--")
			continue
		}

		var parts []string
		// Modifiers first, in M-C-S-A order for consistency
		if slices.Contains(chord.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_META) {
			if mod, ok := helixModifierMapping.GetInverse(keymapv1.KeyModifier_KEY_MODIFIER_META); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_CTRL) {
			if mod, ok := helixModifierMapping.GetInverse(keymapv1.KeyModifier_KEY_MODIFIER_CTRL); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_SHIFT) {
			if mod, ok := helixModifierMapping.GetInverse(keymapv1.KeyModifier_KEY_MODIFIER_SHIFT); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keymapv1.KeyModifier_KEY_MODIFIER_ALT) {
			if mod, ok := helixModifierMapping.GetInverse(keymapv1.KeyModifier_KEY_MODIFIER_ALT); ok {
				parts = append(parts, mod)
			}
		}

		// Then the key
		keyStr, err := toHelixKey(chord.KeyCode)
		if err != nil {
			return "", err
		}
		parts = append(parts, keyStr)

		outChords = append(outChords, strings.Join(parts, "-"))
	}

	return strings.Join(outChords, " "), nil
}

func toHelixKey(kc keymapv1.KeyCode) (string, error) {
	// Check named mappings first
	if helixKey, ok := helixKeyMapping.GetInverse(kc); ok {
		return helixKey, nil
	}

	// Fallback to single characters
	keyStr, ok := keycode.ToString(kc)
	if !ok {
		return "", fmt.Errorf("unsupported keycode: %v", kc)
	}

	if len(keyStr) == 1 {
		return keyStr, nil
	}

	return "", fmt.Errorf("cannot format key for helix: %s", keyStr)
}

func parseKeybinding(s string) (*keymap.KeyBinding, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty keybinding")
	}

	var chords []*keymapv1.KeyChord
	for _, chordStr := range strings.Fields(s) {
		parts := strings.Split(chordStr, "-")
		var modifiers []keymapv1.KeyModifier
		var keyStr string

		// The last part is the key, unless it's an empty string from a trailing '-'
		if len(parts) > 0 {
			keyStr = parts[len(parts)-1]
			// Special case: C-- means Ctrl+Minus
			if keyStr == "" && len(parts) > 1 && parts[len(parts)-2] != "" {
				keyStr = "minus"
			} else if keyStr == "" {
				return nil, fmt.Errorf("invalid key chord format: %s", chordStr)
			}
		}

		// All other parts are modifiers
		modParts := parts[:len(parts)-1]
		for _, modStr := range modParts {
			if mod, ok := helixModifierMapping.Get(modStr); ok {
				modifiers = append(modifiers, mod)
			} else if modStr != "" {
				// Also handle long-form modifiers from older configs
				switch modStr {
				case "cmd", "win", "meta":
					modifiers = append(modifiers, keymapv1.KeyModifier_KEY_MODIFIER_META)
				default:
					return nil, fmt.Errorf("unknown modifier: %s", modStr)
				}
			}
		}

		kc, err := fromHelixKey(keyStr)
		if err != nil {
			return nil, err
		}

		chords = append(chords, &keymapv1.KeyChord{
			Modifiers: modifiers,
			KeyCode:   kc,
		})
	}

	return keymap.NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{Chords: chords}}), nil
}

func fromHelixKey(s string) (keymapv1.KeyCode, error) {
	// Check named mappings first
	if kc, ok := helixKeyMapping.Get(s); ok {
		return kc, nil
	}

	// Fallback to single characters
	if len(s) == 1 {
		if kc, ok := keycode.FromString(s); ok {
			return kc, nil
		}
	}

	return keymapv1.KeyCode_KEY_CODE_UNSPECIFIED, fmt.Errorf("unknown helix key: %s", s)
}
