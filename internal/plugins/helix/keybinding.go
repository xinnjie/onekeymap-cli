package helix

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/bimap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
)

//nolint:gochecknoglobals // static lookup tables shared across parse/format functions; initialized once and read-only
var (
	helixModifierMapping = bimap.NewBiMapFromMap(map[string]keycode.KeyModifier{
		"C": keycode.KeyModifierCtrl,
		"A": keycode.KeyModifierAlt,
		"S": keycode.KeyModifierShift,
		"M": keycode.KeyModifierMeta,
	})

	// see https://github.com/helix-editor/helix/blob/22a3b10dd8ab907367ae1fe57d9703e22b30d391/book/src/remapping.md?plain=1#L92
	helixKeyMapping = bimap.NewBiMapFromMap(map[string]keycode.KeyCode{
		// Special keys from Helix docs
		"backspace": keycode.KeyCodeBackspace, // Note: Helix 'backspace' is our 'delete' (backwards delete)
		"space":     keycode.KeyCodeSpace,
		"ret":       keycode.KeyCodeEnter,
		"esc":       keycode.KeyCodeEscape,
		"del":       keycode.KeyCodeDelete, // Note: Helix 'del' is our 'forward_delete'
		"ins":       keycode.KeyCodeInsert,
		"left":      keycode.KeyCodeLeft,
		"right":     keycode.KeyCodeRight,
		"up":        keycode.KeyCodeUp,
		"down":      keycode.KeyCodeDown,
		"home":      keycode.KeyCodeHome,
		"end":       keycode.KeyCodeEnd,
		"pageup":    keycode.KeyCodePageUp,
		"pagedown":  keycode.KeyCodePageDown,
		"tab":       keycode.KeyCodeTab,
		// TODO(xinnjie) shift+, -> < and shift+. -> > are not supported yet
		// "lt":    keycode.???,
		// "gt":    keycode.???,
		"minus": keycode.KeyCodeMinus,

		// Function keys
		"F1":  keycode.KeyCodeF1,
		"F2":  keycode.KeyCodeF2,
		"F3":  keycode.KeyCodeF3,
		"F4":  keycode.KeyCodeF4,
		"F5":  keycode.KeyCodeF5,
		"F6":  keycode.KeyCodeF6,
		"F7":  keycode.KeyCodeF7,
		"F8":  keycode.KeyCodeF8,
		"F9":  keycode.KeyCodeF9,
		"F10": keycode.KeyCodeF10,
		"F11": keycode.KeyCodeF11,
		"F12": keycode.KeyCodeF12,
	})
)

func formatKeybinding(kb keybinding.Keybinding) (string, error) {
	var outChords []string

	for _, chord := range kb.KeyChords {
		if chord.KeyCode.IsNumpad() {
			return "", ErrNotSupportKeyChords
		}

		// Special case for C--
		if chord.KeyCode == keycode.KeyCodeMinus &&
			len(chord.Modifiers) == 1 &&
			chord.Modifiers[0] == keycode.KeyModifierCtrl {
			outChords = append(outChords, "C--")
			continue
		}

		var parts []string
		// Modifiers first, in M-C-S-A order for consistency
		if slices.Contains(chord.Modifiers, keycode.KeyModifierMeta) {
			if mod, ok := helixModifierMapping.GetInverse(keycode.KeyModifierMeta); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keycode.KeyModifierCtrl) {
			if mod, ok := helixModifierMapping.GetInverse(keycode.KeyModifierCtrl); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keycode.KeyModifierShift) {
			if mod, ok := helixModifierMapping.GetInverse(keycode.KeyModifierShift); ok {
				parts = append(parts, mod)
			}
		}
		if slices.Contains(chord.Modifiers, keycode.KeyModifierAlt) {
			if mod, ok := helixModifierMapping.GetInverse(keycode.KeyModifierAlt); ok {
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

func toHelixKey(kc keycode.KeyCode) (string, error) {
	// Check named mappings first
	if helixKey, ok := helixKeyMapping.GetInverse(kc); ok {
		return helixKey, nil
	}

	// Fallback to single characters
	keyStr := string(kc)
	ok := keyStr != ""
	if !ok {
		return "", fmt.Errorf("unsupported keycode: %v", kc)
	}

	if len(keyStr) == 1 {
		return keyStr, nil
	}

	return "", fmt.Errorf("cannot format key for helix: %s", keyStr)
}
