package intellij

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/xinnjie/onekeymap-cli/internal/bimap"
	"github.com/xinnjie/onekeymap-cli/internal/keymap/keycode"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// Shared well-known key name mappings between IntelliJ naming and our normalized names.
// IntelliJ side uses upper-case tokens like OPEN_BRACKET, BACK_QUOTE, etc.
// Our normalized side uses lower-case names or single-character symbols.
// Source: https://github.com/kasecato/vscode-intellij-idea-keybindings/blob/master/resource/KeystrokeKeyMapping.json
var (
	//nolint:gochecknoglobals // shared lookup table; initialized once and read-only at runtime; mirrors external IntelliJ naming tokens
	modifierMapping = bimap.NewBiMapFromMap(map[string]string{
		"control": "ctrl",
		"shift":   "shift",
		"alt":     "alt",
		"meta":    "meta",
	})

	//nolint:gochecknoglobals // shared key mapping table; initialized once and read-only; values reflect upstream key names
	strokeMapping = bimap.NewBiMapFromMap(map[string]keymapv1.KeyCode{
		// Alphanumeric and punctuation
		"BACK_QUOTE":    keymapv1.KeyCode_BACKTICK,
		"MINUS":         keymapv1.KeyCode_MINUS,
		"EQUALS":        keymapv1.KeyCode_EQUAL,
		"OPEN_BRACKET":  keymapv1.KeyCode_LEFT_BRACKET,
		"CLOSE_BRACKET": keymapv1.KeyCode_RIGHT_BRACKET,
		"BACK_SLASH":    keymapv1.KeyCode_BACKSLASH,
		"SEMICOLON":     keymapv1.KeyCode_SEMICOLON,
		"QUOTE":         keymapv1.KeyCode_QUOTE,
		"COMMA":         keymapv1.KeyCode_COMMA,
		"PERIOD":        keymapv1.KeyCode_PERIOD,
		"SLASH":         keymapv1.KeyCode_SLASH,
		"PLUS":          keymapv1.KeyCode_PLUS,

		// Navigation and control keys
		"LEFT":       keymapv1.KeyCode_LEFT_ARROW,
		"UP":         keymapv1.KeyCode_UP_ARROW,
		"RIGHT":      keymapv1.KeyCode_RIGHT_ARROW,
		"DOWN":       keymapv1.KeyCode_DOWN_ARROW,
		"END":        keymapv1.KeyCode_END,
		"HOME":       keymapv1.KeyCode_HOME,
		"PAGE_UP":    keymapv1.KeyCode_PAGE_UP,
		"PAGE_DOWN":  keymapv1.KeyCode_PAGE_DOWN,
		"TAB":        keymapv1.KeyCode_TAB,
		"ENTER":      keymapv1.KeyCode_RETURN,
		"ESCAPE":     keymapv1.KeyCode_ESCAPE,
		"BACK_SPACE": keymapv1.KeyCode_DELETE, // This is 'backspace' on most keyboards
		"SPACE":      keymapv1.KeyCode_SPACE,
		"DELETE":     keymapv1.KeyCode_FORWARD_DELETE,
		"CAPS_LOCK":  keymapv1.KeyCode_CAPS_LOCK,
		"INSERT":     keymapv1.KeyCode_INSERT_OR_HELP,

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
		"F13": keymapv1.KeyCode_F13,
		"F14": keymapv1.KeyCode_F14,
		"F15": keymapv1.KeyCode_F15,
		"F16": keymapv1.KeyCode_F16,
		"F17": keymapv1.KeyCode_F17,
		"F18": keymapv1.KeyCode_F18,
		"F19": keymapv1.KeyCode_F19,
		"F20": keymapv1.KeyCode_F20,

		// Numpad digits and ops
		"NUMPAD0":      keymapv1.KeyCode_NUMPAD_0,
		"NUMPAD1":      keymapv1.KeyCode_NUMPAD_1,
		"NUMPAD2":      keymapv1.KeyCode_NUMPAD_2,
		"NUMPAD3":      keymapv1.KeyCode_NUMPAD_3,
		"NUMPAD4":      keymapv1.KeyCode_NUMPAD_4,
		"NUMPAD5":      keymapv1.KeyCode_NUMPAD_5,
		"NUMPAD6":      keymapv1.KeyCode_NUMPAD_6,
		"NUMPAD7":      keymapv1.KeyCode_NUMPAD_7,
		"NUMPAD8":      keymapv1.KeyCode_NUMPAD_8,
		"NUMPAD9":      keymapv1.KeyCode_NUMPAD_9,
		"DIVIDE":       keymapv1.KeyCode_NUMPAD_DIVIDE,
		"MULTIPLY":     keymapv1.KeyCode_NUMPAD_MULTIPLY,
		"SUBTRACT":     keymapv1.KeyCode_NUMPAD_MINUS,
		"ADD":          keymapv1.KeyCode_NUMPAD_PLUS,
		"DECIMAL":      keymapv1.KeyCode_NUMPAD_DECIMAL,
		"NUMPAD_ENTER": keymapv1.KeyCode_NUMPAD_ENTER,
	})
)

// parseKeyStroke parses an IntelliJ keystroke string (space separated, e.g., "control alt S")
// into a normalized internal representation (e.g., ["ctrl","alt","s"]).
func parseKeyStroke(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("empty keystroke")
	}

	tokens := strings.Fields(raw)
	var result []string
	var keySeen bool

	for _, tok := range tokens {
		lt := strings.ToLower(tok)

		// Handle modifiers
		if normMod, ok := modifierMapping.Get(lt); ok {
			result = append(result, normMod)
			continue
		} else if _, ok := modifierMapping.GetInverse(lt); ok {
			result = append(result, lt)
			continue
		}

		// Handle key (non-modifier)
		if keySeen {
			return nil, fmt.Errorf("invalid keystroke: multiple keys in '%s'", raw)
		}
		keySeen = true

		norm, ok := normalizeIJKey(tok)
		if !ok {
			return nil, fmt.Errorf("invalid key code: '%s'", tok)
		}
		keyStr, ok := keycode.ToString(norm)
		if !ok {
			return nil, fmt.Errorf("cannot convert keycode to string: %v", norm)
		}
		result = append(result, keyStr)
	}

	return result, nil
}

// normalizeIJKey converts an IntelliJ key token (possibly uppercase word) into
// our normalized token (lowercase, real character where applicable).
func normalizeIJKey(tok string) (keymapv1.KeyCode, bool) {
	t := strings.TrimSpace(tok)
	if t == "" {
		return keymapv1.KeyCode_KEY_CODE_UNSPECIFIED, false
	}

	// Accept single ASCII character keys as-is (lowercased)
	if len(t) == 1 {
		r := rune(t[0])
		// letter or digit or punctuation we explicitly handle below
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("`-=[].'", r) {
			return keycode.FromString(strings.ToLower(t))
		}
	}

	u := strings.ToUpper(t)

	// Well-known named keys via shared mapping
	if v, ok := strokeMapping.Get(u); ok {
		return v, true
	}

	// Accept already-normalized names present in our key set
	return keycode.FromString(strings.ToLower(t))
}

// formatKeyChord formats normalized key parts (e.g., ["ctrl","alt","s"]) into IntelliJ form
// (e.g., "control alt S").
func formatKeyChord(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var modifiers []string
	var key keymapv1.KeyCode

	for _, p := range parts {
		lp := strings.ToLower(p)
		if _, ok := modifierMapping.GetInverse(lp); ok {
			modifiers = append(modifiers, lp)
		} else {
			kc, ok := keycode.FromString(lp)
			if ok {
				key = kc
			}
		}
	}

	var out []string
	for _, mod := range modifiers {
		if ijMod, ok := modifierMapping.GetInverse(mod); ok {
			out = append(out, ijMod)
		}
	}

	if key != keymapv1.KeyCode_KEY_CODE_UNSPECIFIED {
		out = append(out, toIJKey(key))
	}
	return strings.Join(out, " ")
}

// toIJKey converts a normalized key token (lowercase, real character where applicable)
// into the corresponding IntelliJ key token (uppercase word).
//
// It handles the following cases:
// - single character letters/digits
// - function keys (e.g., "f5" -> "F5")
// - numpad digit keys (e.g., "numpad3" -> "NUMPAD3")
// - named special keys via shared mapping (e.g., "enter" -> "ENTER")
// - fallback: return as uppercased word.
func toIJKey(kc keymapv1.KeyCode) string {
	// Named specials via shared mapping first
	if v, ok := strokeMapping.GetInverse(kc); ok {
		return v
	}

	n, ok := keycode.ToString(kc)
	if !ok {
		return ""
	}

	// Single char letters/digits
	if len(n) == 1 {
		r := rune(n[0])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return strings.ToUpper(n)
		}
	}

	// Function key
	if strings.HasPrefix(n, "f") && len(n) > 1 && isDigits(n[1:]) {
		return strings.ToUpper(n)
	}

	// numpad digit
	if strings.HasPrefix(n, "numpad") && len(n) > len("numpad") && isDigits(n[len("numpad"):]) {
		return "NUMPAD" + n[len("numpad"):]
	}

	// Fallback: return as uppercased word
	return strings.ToUpper(n)
}

var digitsRe = regexp.MustCompile(`^[0-9]+$`)

func isDigits(s string) bool { return digitsRe.MatchString(s) }
