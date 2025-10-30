package keychord_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/internal/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      *keymapv1.KeyChord
		expectError   bool
		errorContains string
	}{
		{
			name:  "Simple key",
			input: "a",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("a"),
			},
			expectError: false,
		},
		{
			name:  "Single modifier",
			input: "ctrl+c",
			expected: &keymapv1.KeyChord{
				KeyCode:   keycode.MustKeyCode("c"),
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
			},
			expectError: false,
		},
		{
			name:  "Multiple modifiers",
			input: "ctrl+shift+f",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("f"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
				},
			},
			expectError: false,
		},
		{
			name:  "All modifiers",
			input: "ctrl+alt+shift+meta+enter",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("enter"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_ALT,
					keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
					keymapv1.KeyModifier_KEY_MODIFIER_META,
				},
			},
			expectError: false,
		},
		{
			name:  "Meta modifier",
			input: "meta+s",
			expected: &keymapv1.KeyChord{
				KeyCode:   keycode.MustKeyCode("s"),
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_META},
			},
			expectError: false,
		},
		{
			name:  "Cmd modifier",
			input: "cmd+s",
			expected: &keymapv1.KeyChord{
				KeyCode:   keycode.MustKeyCode("s"),
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_META},
			},
			expectError: false,
		},
		{
			name:  "Win modifier",
			input: "win+r",
			expected: &keymapv1.KeyChord{
				KeyCode:   keycode.MustKeyCode("r"),
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_META},
			},
			expectError: false,
		},
		{
			name:  "ctrl+alt++",
			input: "ctrl+alt++",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("+"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_ALT,
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
			expected: &keymapv1.KeyChord{
				KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
			},
			expectError: false,
		},
		{
			name:  "single modifier without key code(ctrl)",
			input: "ctrl",
			expected: &keymapv1.KeyChord{
				KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := keychord.Parse(tc.input, "+")

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expected, actual, protocmp.Transform())
				assert.Empty(t, diff)
			}
		})
	}
}

func TestParseMinus(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expected      *keymapv1.KeyChord
		expectError   bool
		errorContains string
	}{
		{
			name:  "ctrl-alt-+",
			input: "ctrl-alt-+",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("+"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_ALT,
				},
			},
			expectError: false,
		},

		{
			name:  "ctrl-alt--",
			input: "ctrl-alt--",
			expected: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("-"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_ALT,
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := keychord.Parse(tc.input, "-")

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expected, actual, protocmp.Transform())
				assert.Empty(t, diff)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		name        string
		input       *keymapv1.KeyChord
		expected    []string
		expectError bool
	}{
		{
			name: "Simple key(a)",
			input: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("a"),
			},
			expected: []string{"a"},
		},
		{
			name: "Simple key([)",
			input: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("["),
			},
			expected: []string{"["},
		},
		{
			name: "Single modifier",
			input: &keymapv1.KeyChord{
				KeyCode:   keycode.MustKeyCode("c"),
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
			},
			expected: []string{"ctrl", "c"},
		},
		{
			name: "Multiple modifiers unordered",
			input: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("f"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
				},
			},
			expected: []string{"ctrl", "shift", "f"}, // Should be formatted in canonical order: meta, ctrl, shift, alt
		},
		{
			name: "All modifiers",
			input: &keymapv1.KeyChord{
				KeyCode: keycode.MustKeyCode("x"),
				Modifiers: []keymapv1.KeyModifier{
					keymapv1.KeyModifier_KEY_MODIFIER_META,
					keymapv1.KeyModifier_KEY_MODIFIER_ALT,
					keymapv1.KeyModifier_KEY_MODIFIER_CTRL,
					keymapv1.KeyModifier_KEY_MODIFIER_SHIFT,
				},
			},
			expected: []string{"cmd", "ctrl", "shift", "alt", "x"},
		},
		{
			name:        "Nil input",
			input:       nil,
			expected:    []string{},
			expectError: true,
		},
		{
			name:        "Empty key code",
			input:       &keymapv1.KeyChord{},
			expected:    []string{},
			expectError: true,
		},
		{
			name: "Single modifier(shift)",
			input: &keymapv1.KeyChord{
				KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
			},
			expected: []string{"shift"},
		},
		{
			name: "Single modifier(ctrl)",
			input: &keymapv1.KeyChord{
				KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
				Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
			},
			expected: []string{"ctrl"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyChord := keychord.NewKeyChord(tc.input)
			actual, err := keyChord.Format(platform.PlatformMacOS)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
