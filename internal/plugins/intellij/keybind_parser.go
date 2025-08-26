package intellij

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal"
)

// Shared well-known key name mappings between IntelliJ naming and our normalized names.
// IntelliJ side uses upper-case tokens like OPEN_BRACKET, BACK_QUOTE, etc.
// Our normalized side uses lower-case names or single-character symbols.
// Source: https://github.com/kasecato/vscode-intellij-idea-keybindings/blob/master/resource/KeystrokeKeyMapping.json
var (
	strokeMapping = internal.NewBiMapFromMap(map[string]string{
		// Alphanumeric and punctuation
		"BACK_QUOTE":        "`",
		"MINUS":             "-",
		"EQUALS":            "=",
		"OPEN_BRACKET":      "[",
		"CLOSE_BRACKET":     "]",
		"BACK_SLASH":        `\`,
		"SEMICOLON":         ";",
		"QUOTEDBL":          `"`,
		"QUOTE":             "'",
		"COMMA":             ",",
		"PERIOD":            ".",
		"SLASH":             "/",
		"LEFT_PARENTHESIS":  "(",
		"RIGHT_PARENTHESIS": ")",
		"EXCLAMATION_MARK":  "!",
		"NUMBER_SIGN":       "#",
		"DOLLAR":            "$",
		"CIRCUMFLEX":        "^",
		"AMPERSAND":         "&",
		"ASTERISK":          "*",
		"UNDERSCORE":        "_",
		"PLUS":              "+",
		"COLON":             ":",
		"LESS":              "<",
		"GREATER":           ">",

		// Navigation and control keys
		"LEFT":       "left",
		"UP":         "up",
		"RIGHT":      "right",
		"DOWN":       "down",
		"END":        "end",
		"HOME":       "home",
		"PAGE_UP":    "pageup",
		"PAGE_DOWN":  "pagedown",
		"TAB":        "tab",
		"ENTER":      "enter",
		"ESCAPE":     "escape",
		"BACK_SPACE": "backspace",
		"SPACE":      "space",
		"DELETE":     "delete",
		"PAUSE":      "pause",
		"CAPS_LOCK":  "capslock",
		"INSERT":     "insert",

		// Numpad digits and ops
		"NUMPAD0":      "numpad0",
		"NUMPAD1":      "numpad1",
		"NUMPAD2":      "numpad2",
		"NUMPAD3":      "numpad3",
		"NUMPAD4":      "numpad4",
		"NUMPAD5":      "numpad5",
		"NUMPAD6":      "numpad6",
		"NUMPAD7":      "numpad7",
		"NUMPAD8":      "numpad8",
		"NUMPAD9":      "numpad9",
		"DIVIDE":       "numpad_divide",
		"MULTIPLY":     "numpad_multiply",
		"SUBTRACT":     "numpad_subtract",
		"ADD":          "numpad_add",
		"DECIMAL":      "numpad_decimal",
		"NUMPAD_ENTER": "numpad_enter",
	})
)

// parseKeyStroke parses an IntelliJ keystroke string (space separated, e.g., "control alt S")
// into a normalized internal representation (e.g., ["ctrl","alt","s"]).
func parseKeyStroke(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("empty keystroke")
	}

	tokens := strings.Fields(raw)
	var result []string
	var keySeen bool

	for _, tok := range tokens {
		lt := strings.ToLower(tok)

		switch lt {
		case "control", "ctrl":
			result = append(result, "ctrl")
			continue
		case "shift":
			result = append(result, "shift")
			continue
		case "alt":
			result = append(result, "alt")
			continue
		case "meta":
			result = append(result, "meta")
			continue
		}

		// Non-modifier -> key. Only one key allowed
		if keySeen {
			return nil, fmt.Errorf("invalid keystroke: multiple keys in '%s'", raw)
		}
		keySeen = true

		norm, ok := normalizeIJKey(tok)
		if !ok {
			return nil, fmt.Errorf("invalid key code: '%s'", tok)
		}
		result = append(result, norm)
	}

	return result, nil
}

// normalizeIJKey converts an IntelliJ key token (possibly uppercase word) into
// our normalized token (lowercase, real character where applicable).
func normalizeIJKey(tok string) (string, bool) {
	t := strings.TrimSpace(tok)
	if t == "" {
		return "", false
	}

	// Accept single ASCII character keys as-is (lowercased)
	if len(t) == 1 {
		r := rune(t[0])
		// letter or digit or punctuation we explicitly handle below
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("`-=[].", r) {
			return strings.ToLower(t), true
		}
	}

	u := strings.ToUpper(t)

	// Function keys F1..F24
	if strings.HasPrefix(u, "F") && len(u) > 1 {
		if isDigits(u[1:]) {
			return strings.ToLower(u), true // e.g., F5 -> f5
		}
	}

	// NUMPAD digit keys
	if strings.HasPrefix(u, "NUMPAD") && len(u) > len("NUMPAD") {
		rest := u[len("NUMPAD"):]
		if isDigits(rest) {
			return "numpad" + strings.ToLower(rest), true // NUMPAD3 -> numpad3
		}
	}

	// Well-known named keys via shared mapping
	if v, ok := strokeMapping.Get(u); ok {
		return v, true
	}

	// Accept already-normalized names present in our key set
	switch strings.ToLower(t) {
	case "enter", "tab", "space", "backspace", "delete", "insert", "home", "end", "pageup", "pagedown", "escape":
		return strings.ToLower(t), true
	case "numpad_divide", "numpad_multiply", "numpad_subtract", "numpad_add", "numpad_enter", "numpad_decimal":
		return strings.ToLower(t), true
	}

	return "", false
}

// formatKeyChord formats normalized key parts (e.g., ["ctrl","alt","s"]) into IntelliJ form
// (e.g., "control alt S").
func formatKeyChord(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Separate modifiers and key
	var hasCtrl, hasShift, hasAlt, hasMeta bool
	var key string
	for _, p := range parts {
		switch strings.ToLower(p) {
		case "ctrl":
			hasCtrl = true
		case "shift":
			hasShift = true
		case "alt":
			hasAlt = true
		case "meta":
			hasMeta = true
		default:
			key = p
		}
	}

	var out []string
	if hasCtrl {
		out = append(out, "control")
	}
	if hasShift {
		out = append(out, "shift")
	}
	if hasAlt {
		out = append(out, "alt")
	}
	if hasMeta {
		out = append(out, "meta")
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
// - fallback: return as uppercased word
func toIJKey(norm string) string {
	n := strings.ToLower(strings.TrimSpace(norm))
	if n == "" {
		return ""
	}

	// Single char letters/digits
	if len(n) == 1 {
		r := rune(n[0])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return strings.ToUpper(n)
		}
		if v, ok := strokeMapping.GetInverse(n); ok {
			return v
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

	// Named specials via shared mapping
	if v, ok := strokeMapping.GetInverse(n); ok {
		return v
	}
	// Fallback: return as uppercased word
	return strings.ToUpper(n)
}

var digitsRe = regexp.MustCompile(`^[0-9]+$`)

func isDigits(s string) bool { return digitsRe.MatchString(s) }
