package helix

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/bimap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap/keycode"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// see https://github.com/helix-editor/helix/blob/22a3b10dd8ab907367ae1fe57d9703e22b30d391/book/src/remapping.md?plain=1#L92
var helixWellKnownKeyStroke = bimap.NewBiMapFromMap(map[string]string{
	// Special keys from Helix docs
	"backspace": "backspace",
	"space":     "space",
	"ret":       "enter",
	"esc":       "escape",
	"del":       "delete",
	"ins":       "insert",
	"left":      "left",
	"right":     "right",
	"up":        "up",
	"down":      "down",
	"home":      "home",
	"end":       "end",
	"pageup":    "pageup",
	"pagedown":  "pagedown",
	"tab":       "tab",
	// null is to set keybinding to nothing
	"null": "null",
	// Angle brackets with modifiers
	"lt": "<",
	"gt": ">",
	// Function keys
	"F1":  "f1",
	"F2":  "f2",
	"F3":  "f3",
	"F4":  "f4",
	"F5":  "f5",
	"F6":  "f6",
	"F7":  "f7",
	"F8":  "f8",
	"F9":  "f9",
	"F10": "f10",
	"F11": "f11",
	"F12": "f12",
})

func formatKeybinding(kb *keymap.KeyBinding) (string, error) {
	// helix do not recognize numpad keys, numpad1 is recognized as "1"
	hasNumpadStroke := slices.ContainsFunc(kb.GetKeyChords().GetChords(), func(chord *keymapv1.KeyChord) bool {
		keyStr, ok := keycode.ToString(chord.KeyCode)
		if !ok {
			return false
		}
		return strings.Contains(keyStr, "numpad")
	})
	if hasNumpadStroke {
		return "", ErrNotSupportKeyChords
	}

	// Use linux to normalize meta to "meta" instead of "cmd"/"win".
	s, err := kb.Format(platform.PlatformLinux, "-")
	if err != nil {
		return "", err
	}

	var out []string
	for _, chord := range strings.Fields(s) {
		tokens := strings.Split(chord, "-")
		var mods []string
		var key string
		for _, t := range tokens {
			switch t {
			case "ctrl":
				mods = append(mods, "C")
			case "alt":
				mods = append(mods, "A")
			case "shift":
				mods = append(mods, "S")
			case "meta":
				mods = append(mods, "M")
			default:
				if v, ok := helixWellKnownKeyStroke.GetInverse(t); ok {
					key = v
				} else {
					key = t
				}
			}
		}
		if key != "" {
			if len(mods) > 0 {
				out = append(out, strings.Join(append(mods, key), "-"))
			} else {
				out = append(out, key)
			}
		} else {
			// Handle '-' key specially: when the chord ends with '-', it indicates the '-' key.
			// Expected Helix format for ctrl + '-' is "C--".
			if strings.HasSuffix(chord, "-") && len(mods) > 0 {
				out = append(out, strings.Join(mods, "-")+"--")
			} else if chord == "-" && len(mods) == 0 {
				out = append(out, "-")
			} else {
				out = append(out, strings.Join(mods, "-"))
			}
		}
	}
	return strings.Join(out, " "), nil
}

func parseKeybinding(s string) (*keymap.KeyBinding, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty keybinding")
	}

	var chords []string
	for _, chord := range strings.Fields(s) {
		c := chord
		// Normalize to lowercase for prefix parsing
		lc := strings.ToLower(c)
		var parts []string

		// Consume leading modifier prefixes.
		// Supports short forms: c-, a-, s-, m-
		// And long forms (Helix synonyms for super/meta): cmd-, win-, meta-
		for {
			consumed := false
			// Long-form first
			switch {
			case strings.HasPrefix(lc, "cmd-"):
				parts = append(parts, "meta")
				lc = lc[4:]
				consumed = true
			case strings.HasPrefix(lc, "win-"):
				parts = append(parts, "meta")
				lc = lc[4:]
				consumed = true
			case strings.HasPrefix(lc, "meta-"):
				parts = append(parts, "meta")
				lc = lc[5:]
				consumed = true
			}
			if consumed {
				continue
			}
			// Short-form single-letter prefixes
			if len(lc) >= 2 && lc[1] == '-' {
				switch lc[0] {
				case 'c':
					parts = append(parts, "ctrl")
					lc = lc[2:]
					consumed = true
				case 'a':
					parts = append(parts, "alt")
					lc = lc[2:]
					consumed = true
				case 's':
					parts = append(parts, "shift")
					lc = lc[2:]
					consumed = true
				case 'm':
					parts = append(parts, "meta")
					lc = lc[2:]
					consumed = true
				}
			}
			if !consumed {
				break
			}
		}
		key := strings.ToLower(lc)
		if key != "" {
			if v, ok := helixWellKnownKeyStroke.Get(key); ok {
				key = v
			}
			parts = append(parts, key)
		}
		chords = append(chords, strings.Join(parts, "-"))
	}

	return keymap.ParseKeyBinding(strings.Join(chords, " "), "-")
}
