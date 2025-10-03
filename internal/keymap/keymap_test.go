package keymap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSave(t *testing.T) {
	testCases := []struct {
		name               string
		input              *keymapv1.Keymap
		expectedKeymaps    []OneKeymapConfig
		expectedNumKeymaps int
	}{
		{
			name: "Single keybinding",
			input: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					NewActionBindingWithDescription("actions.copy", "ctrl+c", "copy"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{ID: "actions.copy", Keybinding: KeybindingStrings{"ctrl+c"}, Description: "copy"},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Single keybinding with comment",
			input: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					NewActionBindingWithComment("actions.find", "shift+f", "with comment"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{ID: "actions.find", Keybinding: KeybindingStrings{"shift+f"}, Comment: "with comment"},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for the same action",
			input: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					NewActioinBinding("actions.find", "ctrl+f"),
					NewActioinBinding("actions.find", "cmd+f"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{ID: "actions.find", Keybinding: KeybindingStrings{"ctrl+f", "cmd+f"}},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for the same action",
			input: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					{
						Id:       "actions.find",
						Bindings: newBindingProto("ctrl+f", "cmd+f"),
					},
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{ID: "actions.find", Keybinding: KeybindingStrings{"ctrl+f", "cmd+f"}},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for different actions",
			input: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					NewActioinBinding("actions.copy", "ctrl+c"),
					NewActioinBinding("actions.find", "ctrl+f"),
					NewActioinBinding("actions.find", "cmd+f"),
					NewActionBindingWithComment("actions.find", "shift+f", "with comment"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{ID: "actions.copy", Keybinding: KeybindingStrings{"ctrl+c"}},
				{ID: "actions.find", Keybinding: KeybindingStrings{"ctrl+f", "cmd+f"}},
				{ID: "actions.find", Keybinding: KeybindingStrings{"shift+f"}, Comment: "with comment"},
			},
			expectedNumKeymaps: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Save(&buf, tc.input)
			require.NoError(t, err)

			var result OneKeymapSetting
			err = json.Unmarshal(buf.Bytes(), &result)
			require.NoError(t, err)

			assert.Len(t, result.Keymaps, tc.expectedNumKeymaps)

			for _, expectedMap := range tc.expectedKeymaps {
				found := false
				for _, actualMap := range result.Keymaps {
					if actualMap.ID == expectedMap.ID && actualMap.Comment == expectedMap.Comment {
						assert.ElementsMatch(t, expectedMap.Keybinding, actualMap.Keybinding)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected keymap not found: %+v", expectedMap)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		name          string
		jsonInput     string
		expected      *keymapv1.Keymap
		expectErr     bool
		errorContains string
	}{
		{
			name: "Valid configuration with comment",
			jsonInput: `
{
  "keymaps": [
    {
      "id": "actions.copy",
      "keybinding": "Ctrl+C",
      "comment": "Standard copy command"
    },
    {
      "id": "actions.find",
      "keybinding": "Ctrl+Shift+F"
    }
  ]
}
`,
			expected: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					{
						Id:      "actions.copy",
						Comment: "Standard copy command",
						Bindings: []*keymapv1.Binding{
							{KeyChords: MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "Ctrl+C"},
						},
					},
					{
						Id: "actions.find",
						Bindings: []*keymapv1.Binding{
							{
								KeyChords:         MustParseKeyBinding("ctrl+shift+f").KeyChords,
								KeyChordsReadable: "Ctrl+Shift+F",
							},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Valid configuration with multiple keybindings",
			jsonInput: `
{
  "keymaps": [
    {
      "id": "actions.copy",
      "keybinding": ["Ctrl+C", "Cmd+C"],
      "comment": "Standard copy command"
    }
  ]
}
`,
			expected: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					{
						Id:      "actions.copy",
						Comment: "Standard copy command",
						Bindings: []*keymapv1.Binding{
							{KeyChords: MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "Ctrl+C"},
							{KeyChords: MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "Cmd+C"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Valid configuration with multiple keybindings",
			jsonInput: `
{
  "keymaps": [
    {
      "id": "actions.copy",
      "keybinding": "Ctrl+C",
      "comment": "Standard copy command"
    },
		{
      "id": "actions.copy",
      "keybinding": "Cmd+C",
      "comment": "Standard copy command"
    }
  ]
}
`,
			expected: &keymapv1.Keymap{
				Keybindings: []*keymapv1.ActionBinding{
					{
						Id:      "actions.copy",
						Comment: "Standard copy command",
						Bindings: []*keymapv1.Binding{
							{KeyChords: MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "Ctrl+C"},
							{KeyChords: MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "Cmd+C"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name:      "Empty input",
			jsonInput: "",
			expected:  &keymapv1.Keymap{},
			expectErr: false,
		},
		{
			name:      "Empty keymaps array",
			jsonInput: `{"keymaps": []}`,
			expected:  &keymapv1.Keymap{},
			expectErr: false,
		},
		{
			name:      "Empty keymaps array",
			jsonInput: `{"keymaps": null }`,
			expectErr: true,
		},
		{
			name:      "Unknown field",
			jsonInput: `{"keybinding": null }`,
			expectErr: true,
		},
		{
			name:      "Malformed JSON",
			jsonInput: `{"keymaps": [}`,
			expectErr: true,
		},
		{
			name: "Invalid key chord",
			jsonInput: `
{
  "keymaps": [
    {
      "id": "actions.bad",
      "keybinding": "Ctrl+Alt+Oops"
    }
  ]
}
`,
			expectErr:     true,
			errorContains: "invalid key code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.jsonInput)
			loadedSetting, err := Load(reader)

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				diff := cmp.Diff(tc.expected, loadedSetting, protocmp.Transform())
				assert.Empty(t, diff, "The loaded setting should match the expected one")
			}
		})
	}
}
