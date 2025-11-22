package intellij

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/xinnjie/onekeymap-cli/internal/bimap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
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
	strokeMapping = bimap.NewBiMapFromMap(map[string]keycode.KeyCode{
		// Alphanumeric and punctuation
		"BACK_QUOTE":    keycode.KeyCodeBacktick,
		"MINUS":         keycode.KeyCodeMinus,
		"EQUALS":        keycode.KeyCodeEqual,
		"OPEN_BRACKET":  keycode.KeyCodeLeftBracket,
		"CLOSE_BRACKET": keycode.KeyCodeRightBracket,
		"BACK_SLASH":    keycode.KeyCodeBackslash,
		"SEMICOLON":     keycode.KeyCodeSemicolon,
		"QUOTE":         keycode.KeyCodeQuote,
		"COMMA":         keycode.KeyCodeComma,
		"PERIOD":        keycode.KeyCodePeriod,
		"SLASH":         keycode.KeyCodeSlash,
		"PLUS":          keycode.KeyCodePlus,

		// Navigation and control keys
		"LEFT":       keycode.KeyCodeLeft,
		"UP":         keycode.KeyCodeUp,
		"RIGHT":      keycode.KeyCodeRight,
		"DOWN":       keycode.KeyCodeDown,
		"END":        keycode.KeyCodeEnd,
		"HOME":       keycode.KeyCodeHome,
		"PAGE_UP":    keycode.KeyCodePageUp,
		"PAGE_DOWN":  keycode.KeyCodePageDown,
		"TAB":        keycode.KeyCodeTab,
		"ENTER":      keycode.KeyCodeEnter,
		"ESCAPE":     keycode.KeyCodeEscape,
		"BACK_SPACE": keycode.KeyCodeBackspace, // This is 'backspace' on most keyboards
		"SPACE":      keycode.KeyCodeSpace,
		"DELETE":     keycode.KeyCodeDelete,
		"CAPS_LOCK":  keycode.KeyCodeCapsLock,
		"INSERT":     keycode.KeyCodeInsert,

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
		"F13": keycode.KeyCodeF13,
		"F14": keycode.KeyCodeF14,
		"F15": keycode.KeyCodeF15,
		"F16": keycode.KeyCodeF16,
		"F17": keycode.KeyCodeF17,
		"F18": keycode.KeyCodeF18,
		"F19": keycode.KeyCodeF19,
		"F20": keycode.KeyCodeF20,

		// Numpad digits and ops
		"NUMPAD0":      keycode.KeyCodeNumpad0,
		"NUMPAD1":      keycode.KeyCodeNumpad1,
		"NUMPAD2":      keycode.KeyCodeNumpad2,
		"NUMPAD3":      keycode.KeyCodeNumpad3,
		"NUMPAD4":      keycode.KeyCodeNumpad4,
		"NUMPAD5":      keycode.KeyCodeNumpad5,
		"NUMPAD6":      keycode.KeyCodeNumpad6,
		"NUMPAD7":      keycode.KeyCodeNumpad7,
		"NUMPAD8":      keycode.KeyCodeNumpad8,
		"NUMPAD9":      keycode.KeyCodeNumpad9,
		"DIVIDE":       keycode.KeyCodeNumpadDivide,
		"MULTIPLY":     keycode.KeyCodeNumpadMultiply,
		"SUBTRACT":     keycode.KeyCodeNumpadSubtract,
		"ADD":          keycode.KeyCodeNumpadAdd,
		"DECIMAL":      keycode.KeyCodeNumpadDecimal,
		"NUMPAD_ENTER": keycode.KeyCodeNumpadEnter,
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
		result = append(result, string(norm))
	}

	return result, nil
}

// normalizeIJKey converts an IntelliJ key token (possibly uppercase word) into
// our normalized token (lowercase, real character where applicable).
func normalizeIJKey(tok string) (keycode.KeyCode, bool) {
	t := strings.TrimSpace(tok)
	if t == "" {
		return "", false
	}

	// Accept single ASCII character keys as-is (lowercased)
	if len(t) == 1 {
		r := rune(t[0])
		// letter or digit or punctuation we explicitly handle below
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("`-=[].'", r) {
			return keycode.KeyCode(strings.ToLower(t)), true
		}
	}

	u := strings.ToUpper(t)

	// Well-known named keys via shared mapping
	if v, ok := strokeMapping.Get(u); ok {
		return v, true
	}

	// Accept already-normalized names present in our key set
	k := keycode.KeyCode(strings.ToLower(t))
	if k.IsValid() {
		return k, true
	}
	return "", false
}

// formatKeyChord formats normalized key parts (e.g., ["ctrl","alt","s"]) into IntelliJ form
// (e.g., "control alt S").
func formatKeyChord(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	var modifiers []string
	var key keycode.KeyCode

	for _, p := range parts {
		lp := strings.ToLower(p)
		if _, ok := modifierMapping.GetInverse(lp); ok {
			modifiers = append(modifiers, lp)
		} else {
			key = keycode.KeyCode(lp)
		}
	}

	var out []string
	for _, mod := range modifiers {
		if ijMod, ok := modifierMapping.GetInverse(mod); ok {
			out = append(out, ijMod)
		}
	}

	if key != "" {
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
func toIJKey(kc keycode.KeyCode) string {
	// Named specials via shared mapping first
	if v, ok := strokeMapping.GetInverse(kc); ok {
		return v
	}

	n := string(kc)
	if n == "" {
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
