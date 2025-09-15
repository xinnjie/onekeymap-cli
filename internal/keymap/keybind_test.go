package keymap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap/keycode"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/proto"
)

func TestParseKeyBinding(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    *keymapv1.KeyChordSequence
		expectError bool
	}{
		{
			name:  "Single chord(ctrl+s)",
			input: "ctrl+s",
			expected: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keycode.MustKeyCode("s"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
					},
				},
			},
		},
		{
			name:  "Multi-chord(ctrl+k ctrl+s)",
			input: "ctrl+k ctrl+s",
			expected: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keycode.MustKeyCode("k"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
					},
					{
						KeyCode:   keycode.MustKeyCode("s"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
					},
				},
			},
		},
		{
			name:  "Multi-chord(shift shift)",
			input: "shift shift",
			expected: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
					},
					{
						KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
					},
				},
			},
		},
		{
			name:        "Invalid chord",
			input:       "ctrl+invalidkey",
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseKeyBinding(tc.input, "+")
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tc.expected, actual.KeyChords), "Expected %v, got %v", tc.expected, actual.KeyChords)
			}
		})
	}
}

func TestKeyBinding_Format(t *testing.T) {
	testCases := []struct {
		name        string
		input       *KeyBinding
		separator   string
		expected    string
		expectError bool
	}{
		{
			name: "Single chord(cmd+s)",
			input: NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keycode.MustKeyCode("s"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_META},
					},
				},
			}}),
			separator: "+",
			expected:  "cmd+s",
		},
		{
			name: "Multi-chord(shift shift)",
			input: NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
					},
					{
						KeyCode:   keymapv1.KeyCode_KEY_CODE_UNSPECIFIED,
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_SHIFT},
					},
				},
			}}),
			separator: "-",
			expected:  "shift shift",
		},
		{
			name: "Multi-chord(ctrl+k ctrl+s)",
			input: NewKeyBinding(&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{
				Chords: []*keymapv1.KeyChord{
					{
						KeyCode:   keycode.MustKeyCode("k"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
					},
					{
						KeyCode:   keycode.MustKeyCode("s"),
						Modifiers: []keymapv1.KeyModifier{keymapv1.KeyModifier_KEY_MODIFIER_CTRL},
					},
				},
			}}),
			separator: "-",
			expected:  "ctrl-k ctrl-s",
		},
		{
			name:        "Nil KeyBinding",
			input:       nil,
			separator:   "+",
			expectError: true,
		},
		{
			name: "Empty Chords",
			input: NewKeyBinding(
				&keymapv1.Binding{KeyChords: &keymapv1.KeyChordSequence{Chords: []*keymapv1.KeyChord{}}},
			),
			separator:   "+",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.input.Format(platform.PlatformMacOS, tc.separator)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
