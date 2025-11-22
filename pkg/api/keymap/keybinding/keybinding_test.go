package keybinding_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

func TestNewKeybinding(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		opt         keybinding.ParseOption
		expected    keybinding.Keybinding
		expectError bool
	}{
		{
			name:  "Single chord(ctrl+s)",
			input: "ctrl+s",
			opt:   keybinding.ParseOption{Separator: "+"},
			expected: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   keycode.KeyCodeS,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
					},
				},
			},
		},
		{
			name:  "Multi-chord(ctrl+k ctrl+s)",
			input: "ctrl+k ctrl+s",
			opt:   keybinding.ParseOption{Separator: "+"},
			expected: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   keycode.KeyCodeK,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
					},
					{
						KeyCode:   keycode.KeyCodeS,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
					},
				},
			},
		},
		{
			name:  "Multi-chord(shift shift)",
			input: "shift shift",
			opt:   keybinding.ParseOption{Separator: "+"},
			expected: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   "", // Match behavior of KEY_CODE_UNSPECIFIED
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
					},
					{
						KeyCode:   "", // Match behavior of KEY_CODE_UNSPECIFIED
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
					},
				},
			},
		},
		{
			name:        "Invalid chord",
			input:       "ctrl+invalidkey",
			opt:         keybinding.ParseOption{Separator: "+"},
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			opt:         keybinding.ParseOption{Separator: "+"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := keybinding.NewKeybinding(tc.input, tc.opt)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestKeybinding_String(t *testing.T) {
	testCases := []struct {
		name        string
		input       keybinding.Keybinding
		opt         keybinding.FormatOption
		expected    string
		expectPanic bool // String() might panic or not handle nil/empty gracefully if implemented poorly, but here we check output
	}{
		{
			name: "Single chord(cmd+s)",
			input: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   keycode.KeyCodeS,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierMeta},
					},
				},
			},
			opt: keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			},
			expected: "cmd+s",
		},
		{
			name: "Multi-chord(shift shift)",
			input: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   "",
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
					},
					{
						KeyCode:   "",
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierShift},
					},
				},
			},
			opt: keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "-",
			},
			expected: "shift shift",
		},
		{
			name: "Multi-chord(ctrl+k ctrl+s)",
			input: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{
					{
						KeyCode:   keycode.KeyCodeK,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
					},
					{
						KeyCode:   keycode.KeyCodeS,
						Modifiers: []keycode.KeyModifier{keycode.KeyModifierCtrl},
					},
				},
			},
			opt: keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "-",
			},
			expected: "ctrl-k ctrl-s",
		},
		{
			name: "Empty Chords",
			input: keybinding.Keybinding{
				KeyChords: []keychord.KeyChord{},
			},
			opt: keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			},
			expected: "", // Or expect error/panic depending on implementation desired. Old test expected Error.
			// The String() method signature returns string, not (string, error).
			// So it probably shouldn't error, but return empty string.
			// The old test `input: keymap.NewKeyBinding(...)` returned error for empty chords.
			// But the new signature is `func (kb Keybinding) String(...) string`.
			// So we assume it returns empty string.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				require.Panics(t, func() {
					tc.input.String(tc.opt)
				})
			} else {
				actual := tc.input.String(tc.opt)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
