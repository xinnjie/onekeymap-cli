package xcode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
)

// Xcode uses specific symbols for modifier keys:
// @ = Cmd (Command)
// ^ = Ctrl (Control)
// ~ = Alt/Option
// $ = Shift
// No separator between modifier keys and main key

// parseKeybinding converts Xcode format (@k, ^g, ~@l) to internal format (cmd+k, ctrl+g, alt+cmd+l)
func parseKeybinding(keybind string) (*keymap.KeyBinding, error) {
	normalized, err := normalizeXcodeKeybind(keybind)
	if err != nil {
		return nil, err
	}
	return keymap.ParseKeyBinding(normalized, "+")
}

// formatKeybinding converts internal format back to Xcode format
func formatKeybinding(keybind *keymap.KeyBinding) (string, error) {
	formatted, err := keybind.Format(platform.PlatformMacOS, "+")
	if err != nil {
		return "", err
	}
	return denormalizeXcodeKeybind(formatted)
}

// normalizeXcodeKeybind converts Xcode format to standard format
// @k -> cmd+k, ^g -> ctrl+g, ~@l -> alt+cmd+l
func normalizeXcodeKeybind(xcodeKeybind string) (string, error) {
	if xcodeKeybind == "" {
		return "", errors.New("empty keybind")
	}

	var modifiers []string
	runes := []rune(xcodeKeybind)

	// Find the main key (last character that's not a modifier symbol)
	var mainKey string
	i := len(runes) - 1

	// Handle special cases like tab character
	if i >= 0 {
		lastRune := runes[i]
		switch {
		case lastRune == '\t':
			mainKey = "tab"
		case lastRune >= 'a' && lastRune <= 'z':
			mainKey = string(lastRune)
		case lastRune >= 'A' && lastRune <= 'Z':
			mainKey = strings.ToLower(string(lastRune))
		default:
			mainKey = string(lastRune)
		}
		i--
	}

	// Parse modifiers from left to right
	for j := 0; j <= i; j++ {
		switch runes[j] {
		case '@':
			modifiers = append(modifiers, "cmd")
		case '^':
			modifiers = append(modifiers, "ctrl")
		case '~':
			modifiers = append(modifiers, "alt")
		case '$':
			modifiers = append(modifiers, "shift")
		default:
			return "", fmt.Errorf("unknown modifier symbol: %c", runes[j])
		}
	}

	if mainKey == "" {
		return "", fmt.Errorf("no main key found in: %s", xcodeKeybind)
	}

	if len(modifiers) == 0 {
		return mainKey, nil
	}

	return strings.Join(modifiers, "+") + "+" + mainKey, nil
}

// denormalizeXcodeKeybind converts standard format back to Xcode format
// cmd+k -> @k, ctrl+g -> ^g, alt+cmd+l -> ~@l
func denormalizeXcodeKeybind(standardKeybind string) (string, error) {
	parts := strings.Split(standardKeybind, "+")
	if len(parts) == 0 {
		return "", errors.New("empty keybind")
	}

	var xcodeFormat strings.Builder
	mainKey := parts[len(parts)-1]

	// Convert modifiers
	for _, part := range parts[:len(parts)-1] {
		switch strings.ToLower(part) {
		case "cmd", "meta":
			xcodeFormat.WriteRune('@')
		case "ctrl":
			xcodeFormat.WriteRune('^')
		case "alt":
			xcodeFormat.WriteRune('~')
		case "shift":
			xcodeFormat.WriteRune('$')
		}
	}

	// Add main key
	if mainKey == "tab" {
		xcodeFormat.WriteRune('\t')
	} else {
		xcodeFormat.WriteString(mainKey)
	}

	return xcodeFormat.String(), nil
}
