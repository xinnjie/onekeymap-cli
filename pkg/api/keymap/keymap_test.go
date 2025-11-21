package keymap_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

// Helper function to create a keybinding from a string
func mustNewKeybinding(s string) keybinding.Keybinding {
	kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Separator: "+"})
	if err != nil {
		panic(err)
	}
	return kb
}

// Helper function to create an action with a single keybinding
func newAction(name, keybind string) keymap.Action {
	return keymap.Action{
		Name:     name,
		Bindings: []keybinding.Keybinding{mustNewKeybinding(keybind)},
	}
}

// Helper function to create an action with multiple keybindings
func newActionWithBindings(name string, keybinds ...string) keymap.Action {
	action := keymap.Action{
		Name:     name,
		Bindings: make([]keybinding.Keybinding, 0, len(keybinds)),
	}
	for _, kb := range keybinds {
		action.Bindings = append(action.Bindings, mustNewKeybinding(kb))
	}
	return action
}

func TestSave(t *testing.T) {
	testCases := []struct {
		name               string
		input              keymap.Keymap
		expectedKeymaps    []expectedKeymap
		expectedNumKeymaps int
	}{
		{
			name: "Single keybinding",
			input: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.copy", "ctrl+c"),
				},
			},
			expectedKeymaps: []expectedKeymap{
				{ID: "actions.copy", Keybinding: []string{"ctrl+c"}},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for the same action",
			input: keymap.Keymap{
				Actions: []keymap.Action{
					newActionWithBindings("actions.find", "ctrl+f", "cmd+f"),
				},
			},
			expectedKeymaps: []expectedKeymap{
				{ID: "actions.find", Keybinding: []string{"ctrl+f", "cmd+f"}},
			},
			expectedNumKeymaps: 1,
		},
		{
			name: "Multiple keybindings for different actions",
			input: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.copy", "ctrl+c"),
					newActionWithBindings("actions.find", "ctrl+f", "cmd+f", "shift+f"),
				},
			},
			expectedKeymaps: []expectedKeymap{
				{ID: "actions.copy", Keybinding: []string{"ctrl+c"}},
				{ID: "actions.find", Keybinding: []string{"ctrl+f", "cmd+f", "shift+f"}},
			},
			expectedNumKeymaps: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := keymap.Save(&buf, tc.input, keymap.SaveOptions{Platform: platform.PlatformMacOS})
			require.NoError(t, err)

			var result struct {
				Version string `json:"version"`
				Keymaps []struct {
					ID         string      `json:"id,omitempty"`
					Keybinding interface{} `json:"keybinding,omitempty"`
					Comment    string      `json:"comment,omitempty"`
				} `json:"keymaps"`
			}
			err = json.Unmarshal(buf.Bytes(), &result)
			require.NoError(t, err)

			assert.Len(t, result.Keymaps, tc.expectedNumKeymaps)

			for _, expectedMap := range tc.expectedKeymaps {
				found := false
				for _, actualMap := range result.Keymaps {
					if actualMap.ID == expectedMap.ID && actualMap.Comment == expectedMap.Comment {
						// Handle both string and []string keybinding formats
						var actualBindings []string
						switch v := actualMap.Keybinding.(type) {
						case string:
							actualBindings = []string{v}
						case []interface{}:
							for _, item := range v {
								if s, ok := item.(string); ok {
									actualBindings = append(actualBindings, s)
								}
							}
						}
						assert.ElementsMatch(t, expectedMap.Keybinding, actualBindings)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected keymap not found: %+v", expectedMap)
			}
		})
	}
}

type expectedKeymap struct {
	ID         string
	Keybinding []string
	Comment    string
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		name          string
		jsonInput     string
		expected      keymap.Keymap
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.copy", "ctrl+c"),
					newAction("actions.find", "ctrl+shift+f"),
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					newActionWithBindings("actions.copy", "ctrl+c", "cmd+c"),
				},
			},
			expectErr: false,
		},
		{
			name: "Valid configuration with multiple keybindings (separate entries)",
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					newActionWithBindings("actions.copy", "ctrl+c", "cmd+c"),
				},
			},
			expectErr: false,
		},
		{
			name:      "Empty input",
			jsonInput: "",
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name:      "Empty keymaps array",
			jsonInput: `{"keymaps": []}`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name:      "Null keymaps array",
			jsonInput: `{"keymaps": null }`,
			expected:  keymap.Keymap{},
			expectErr: false,
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
			loadedSetting, err := keymap.Load(reader, keymap.LoadOptions{})

			if tc.expectErr {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				
				// Custom comparison for Keymap
				require.Equal(t, len(tc.expected.Actions), len(loadedSetting.Actions), "Number of actions should match")
				
				for i, expectedAction := range tc.expected.Actions {
					actualAction := loadedSetting.Actions[i]
					assert.Equal(t, expectedAction.Name, actualAction.Name, "Action name should match")
					
					require.Equal(t, len(expectedAction.Bindings), len(actualAction.Bindings), "Number of bindings should match for action %s", expectedAction.Name)
					
					for j, expectedBinding := range expectedAction.Bindings {
						actualBinding := actualAction.Bindings[j]
						expectedStr := expectedBinding.String(keybinding.FormatOption{
							Platform:  platform.PlatformMacOS,
							Separator: "+",
						})
						actualStr := actualBinding.String(keybinding.FormatOption{
							Platform:  platform.PlatformMacOS,
							Separator: "+",
						})
						assert.Equal(t, expectedStr, actualStr, "Keybinding should match")
					}
				}
			}
		})
	}
}

// TestLoadAndSaveRoundTrip tests that loading and saving preserves the data
func TestLoadAndSaveRoundTrip(t *testing.T) {
	originalJSON := `{
  "version": "1.0",
  "keymaps": [
    {
      "id": "actions.clipboard.copy",
      "keybinding": "cmd+c"
    },
    {
      "id": "actions.clipboard.paste",
      "keybinding": ["cmd+v", "shift+insert"]
    }
  ]
}`

	// Load
	km, err := keymap.Load(strings.NewReader(originalJSON), keymap.LoadOptions{})
	require.NoError(t, err)

	// Save
	var buf bytes.Buffer
	err = keymap.Save(&buf, km, keymap.SaveOptions{Platform: platform.PlatformMacOS})
	require.NoError(t, err)

	// Load again
	km2, err := keymap.Load(&buf, keymap.LoadOptions{})
	require.NoError(t, err)

	// Compare
	require.Equal(t, len(km.Actions), len(km2.Actions))
	for i := range km.Actions {
		assert.Equal(t, km.Actions[i].Name, km2.Actions[i].Name)
		assert.Equal(t, len(km.Actions[i].Bindings), len(km2.Actions[i].Bindings))
	}
}

// TestSaveEmptyKeymap tests saving an empty keymap
func TestSaveEmptyKeymap(t *testing.T) {
	km := keymap.Keymap{Actions: []keymap.Action{}}
	var buf bytes.Buffer
	err := keymap.Save(&buf, km, keymap.SaveOptions{})
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "1.0", result["version"])
	keymaps := result["keymaps"].([]interface{})
	assert.Len(t, keymaps, 0)
}
