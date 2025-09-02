package keymap

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSave(t *testing.T) {
	testCases := []struct {
		name               string
		input              *keymapv1.KeymapSetting
		expectedKeymaps    []OneKeymapConfig
		expectedNumKeymaps int
	}{
		{
			name: "Single keybinding",
			input: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBindingWithDescription("actions.copy", "ctrl+c", "copy"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{Id: "actions.copy", Keybinding: KeybindingStrings{"ctrl+c"}, Description: "copy"},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Single keybinding with comment",
			input: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBindingWithComment("actions.find", "shift+f", "with comment"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{Id: "actions.find", Keybinding: KeybindingStrings{"shift+f"}, Comment: "with comment"},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for the same action",
			input: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBinding("actions.find", "ctrl+f"),
					NewBinding("actions.find", "cmd+f"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{Id: "actions.find", Keybinding: KeybindingStrings{"ctrl+f", "cmd+f"}},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for different actions",
			input: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBinding("actions.copy", "ctrl+c"),
					NewBinding("actions.find", "ctrl+f"),
					NewBinding("actions.find", "cmd+f"),
					NewBindingWithComment("actions.find", "shift+f", "with comment"),
				},
			},
			expectedKeymaps: []OneKeymapConfig{
				{Id: "actions.copy", Keybinding: KeybindingStrings{"ctrl+c"}},
				{Id: "actions.find", Keybinding: KeybindingStrings{"ctrl+f", "cmd+f"}},
				{Id: "actions.find", Keybinding: KeybindingStrings{"shift+f"}, Comment: "with comment"},
			},
			expectedNumKeymaps: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Save(&buf, tc.input)
			assert.NoError(t, err)

			var result OneKeymapSetting
			err = json.Unmarshal(buf.Bytes(), &result)
			assert.NoError(t, err)

			assert.Len(t, result.Keymaps, tc.expectedNumKeymaps)

			for _, expectedMap := range tc.expectedKeymaps {
				found := false
				for _, actualMap := range result.Keymaps {
					if actualMap.Id == expectedMap.Id && actualMap.Comment == expectedMap.Comment {
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
		expected      *keymapv1.KeymapSetting
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBindingWithComment("actions.copy", "ctrl+c", "Standard copy command"),
					NewBinding("actions.find", "ctrl+shift+f"),
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.KeyBinding{
					NewBindingWithComment("actions.copy", "ctrl+c", "Standard copy command"),
					NewBindingWithComment("actions.copy", "cmd+c", "Standard copy command"),
				},
			},
			expectErr: false,
		},
		{
			name:      "Empty keymaps array",
			jsonInput: `{"keymaps": []}`,
			expected:  &keymapv1.KeymapSetting{},
			expectErr: false,
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
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				diff := cmp.Diff(tc.expected, loadedSetting, protocmp.Transform())
				assert.Empty(t, diff, "The loaded setting should match the expected one")
			}
		})
	}
}
