package keychord_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
)

func TestNewKeyChord(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      keychord.KeyChord
		expectError   bool
		errorContains string
	}{
		{
			name:  "Simple key",
			input: "a",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodeA,
			},
			expectError: false,
		},
		{
			name:  "Single modifier",
			input: "ctrl+c",
			expected: keychord.KeyChord{
				KeyCode:   keycode.KeyCodeC,
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
			},
			expectError: false,
		},
		{
			name:  "Multiple modifiers",
			input: "ctrl+shift+f",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodeF,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierCtrl,
					keycode.KeyModifierShift,
				},
			},
			expectError: false,
		},
		{
			name:  "All modifiers",
			input: "ctrl+alt+shift+meta+enter",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodeEnter,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierCtrl,
					keycode.KeyModifierAlt,
					keycode.KeyModifierShift,
					keycode.KeyModifierMeta,
				},
			},
			expectError: false,
		},
		{
			name:  "Meta modifier",
			input: "meta+s",
			expected: keychord.KeyChord{
				KeyCode:   keycode.KeyCodeS,
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierMeta},
			},
			expectError: false,
		},
		{
			name:  "Cmd modifier",
			input: "cmd+s",
			expected: keychord.KeyChord{
				KeyCode:   keycode.KeyCodeS,
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierMeta},
			},
			expectError: false,
		},
		{
			name:  "Win modifier",
			input: "win+r",
			expected: keychord.KeyChord{
				KeyCode:   keycode.KeyCodeR,
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierMeta},
			},
			expectError: false,
		},
		{
			name:  "ctrl+alt++",
			input: "ctrl+alt++",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodePlus,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierCtrl,
					keycode.KeyModifierAlt,
				},
			},
			expectError: false,
		},
		{
			name:          "Empty string",
			input:         "",
			expectError:   true,
			errorContains: "cannot parse empty string",
		},
		{
			name:          "Multiple key codes",
			input:         "ctrl+a+b",
			expectError:   true,
			errorContains: "multiple key codes found",
		},
		{
			name:          "No key code",
			input:         "ctrl+shift",
			expectError:   true,
			errorContains: "no key code found",
		},
		{
			name:  "single modifier without key code(shift)",
			input: "shift",
			expected: keychord.KeyChord{
				KeyCode:   "", // Or some "Unspecified" constant if available, checking keycode.go
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
			},
			expectError: false,
		},
		{
			name:  "single modifier without key code(ctrl)",
			input: "ctrl",
			expected: keychord.KeyChord{
				KeyCode:   "",
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := keychord.NewKeyChord(tc.input, keychord.ParseOption{Separator: "+"})

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expected, actual)
				assert.Empty(t, diff)
			}
		})
	}
}

func TestNewKeyChord_Minus(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      keychord.KeyChord
		expectError   bool
		errorContains string
	}{
		{
			name:  "ctrl-alt-+",
			input: "ctrl-alt-+",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodePlus,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierCtrl,
					keycode.KeyModifierAlt,
				},
			},
			expectError: false,
		},

		{
			name:  "ctrl-alt--",
			input: "ctrl-alt--",
			expected: keychord.KeyChord{
				KeyCode: keycode.KeyCodeMinus,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierCtrl,
					keycode.KeyModifierAlt,
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := keychord.NewKeyChord(tc.input, keychord.ParseOption{Separator: "-"})

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expected, actual)
				assert.Empty(t, diff)
			}
		})
	}
}

func TestKeyChord_String(t *testing.T) {
	testCases := []struct {
		name        string
		input       keychord.KeyChord
		expected    string
		expectError bool
	}{
		{
			name: "Simple key(a)",
			input: keychord.KeyChord{
				KeyCode: keycode.KeyCodeA,
			},
			expected: "a",
		},
		{
			name: "Simple key([)",
			input: keychord.KeyChord{
				KeyCode: keycode.KeyCodeLeftBracket,
			},
			expected: "[",
		},
		{
			name: "Single modifier",
			input: keychord.KeyChord{
				KeyCode:   keycode.KeyCodeC,
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
			},
			expected: "ctrl+c",
		},
		{
			name: "Multiple modifiers unordered",
			input: keychord.KeyChord{
				KeyCode: keycode.KeyCodeF,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierShift,
					keycode.KeyModifierCtrl,
				},
			},
			expected: "ctrl+shift+f", // Should be formatted in canonical order: meta, ctrl, shift, alt
		},
		{
			name: "All modifiers",
			input: keychord.KeyChord{
				KeyCode: keycode.KeyCodeX,
				Modifiers: []keycode.KeyModifier{
					keycode.KeyModifierMeta,
					keycode.KeyModifierAlt,
					keycode.KeyModifierCtrl,
					keycode.KeyModifierShift,
				},
			},
			expected: "cmd+ctrl+shift+alt+x",
		},
		{
			name: "Single modifier(shift)",
			input: keychord.KeyChord{
				KeyCode:   "",
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
			},
			expected: "shift",
		},
		{
			name: "Single modifier(ctrl)",
			input: keychord.KeyChord{
				KeyCode:   "",
				Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
			},
			expected: "ctrl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.String(keychord.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			})
			assert.Equal(t, tc.expected, actual)
		})
	}
}
