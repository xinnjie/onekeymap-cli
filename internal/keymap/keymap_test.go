package keymap

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	// 1. Define the initial KeymapSetting proto object.
	originalSetting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			NewBinding("actions.copy", "ctrl+c"),
			NewBinding("actions.find", "ctrl+shift+f"),
			NewBinding("actions.noop", "f12"),
		},
	}

	// 2. Save the setting to an in-memory buffer.
	var buf bytes.Buffer
	err := Save(&buf, originalSetting)
	assert.NoError(t, err, "Save should not produce an error")

	// 3. Load the setting back from the buffer.
	loadedSetting, err := Load(&buf)
	assert.NoError(t, err, "Load should not produce an error")

	// 4. Compare the original and loaded settings.
	// Using protocmp is the correct way to compare protobuf messages.
	diff := cmp.Diff(originalSetting, loadedSetting, protocmp.Transform())
	assert.Empty(t, diff, "The loaded setting should be identical to the original")
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
					func() *keymapv1.KeyBinding {
						binding := NewBinding("actions.copy", "ctrl+c")
						binding.Comment = "Standard copy command"
						return binding
					}(),
					NewBinding("actions.find", "ctrl+shift+f"),
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
