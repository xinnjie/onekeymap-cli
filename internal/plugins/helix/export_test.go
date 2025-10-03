package helix

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func decodeHelixTOMLMap(t *testing.T, s string) map[string]any {
	var got map[string]any
	require.NoError(t, toml.NewDecoder(bytes.NewBufferString(s)).Decode(&got))
	return got
}

func TestExportHelixKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	tests := []struct {
		name           string
		setting        *keymapv1.Keymap
		wantTOML       string
		existingConfig string
	}{
		// Basic destructive export tests
		{
			name: "export copy keymap",
			setting: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			wantTOML: `
[keys.insert]
"M-c" = "yank"
`,
		},
		// Non-destructive export tests
		{
			name: "non-destructive export preserves user keybindings",
			setting: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `
[keys.normal]
"C-x" = "custom_user_command"

[keys.insert]
"C-v" = "custom_paste_command"
`,
			wantTOML: `
[keys.normal]
"C-x" = "custom_user_command"

[keys.insert]
"M-c" = "yank"
"C-v" = "custom_paste_command"
`,
		},
		{
			name: "managed keybinding takes priority over conflicting user keybinding",
			setting: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `
[keys.insert]
"M-c" = "custom_conflicting_command"
"C-v" = "custom_paste_command"
`,
			wantTOML: `
[keys.insert]
"M-c" = "yank"
"C-v" = "custom_paste_command"
`,
		},
		{
			name: "multiple modes with mixed conflicts",
			setting: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `
[keys.normal]
"C-x" = "custom_cut_command"
"C-z" = "custom_undo_command"

[keys.insert]
"M-c" = "conflicting_copy_command"
"C-v" = "custom_paste_command"

[keys.select]
"C-a" = "custom_select_all_command"
`,
			wantTOML: `
[keys.normal]
"C-x" = "custom_cut_command"
"C-z" = "custom_undo_command"

[keys.insert]
"M-c" = "yank"
"C-v" = "custom_paste_command"

[keys.select]
"C-a" = "custom_select_all_command"
`,
		},
		{
			name: "empty existing config behaves as destructive export",
			setting: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
				},
			},
			existingConfig: `[keys.normal]`,
			wantTOML: `
[keys.insert]
"C-c" = "yank"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(mappingConfig, slog.New(slog.NewTextHandler(io.Discard, nil)))
			exporter, err := p.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{ExistingConfig: nil}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.setting, opts)
			require.NoError(t, err)

			gotMap := decodeHelixTOMLMap(t, buf.String())
			wantMap := decodeHelixTOMLMap(t, tt.wantTOML)
			assert.Equal(t, wantMap, gotMap)
		})
	}
}

func TestExportHelixKeymap_PreservesOtherSections(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	// Existing user configuration to be preserved
	existingConfig := `theme = "onedark"

[editor]
line-number = "relative"
mouse = false

[editor.cursor-shape]
insert = "bar"
normal = "block"

[keys.normal]
"C-x" = "extend_line_below"`

	// Managed keymap setting to apply
	setting := &keymapv1.Keymap{
		Keybindings: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
		},
	}

	p := New(mappingConfig, slog.New(slog.NewTextHandler(io.Discard, nil)))
	exporter, err := p.Exporter()
	require.NoError(t, err)

	var buf bytes.Buffer
	opts := pluginapi.PluginExportOption{ExistingConfig: nil}
	if existingConfig != "" {
		opts.ExistingConfig = strings.NewReader(existingConfig)
	}

	report, err := exporter.Export(context.TODO(), &buf, setting, opts)
	require.NoError(t, err)
	require.NotNil(t, report)

	result := buf.String()
	t.Logf("Actual output:\n%s", result)

	// Parse as generic map to check all sections
	var fullConfig map[string]interface{}
	err = toml.Unmarshal([]byte(result), &fullConfig)
	require.NoError(t, err)

	t.Logf("Parsed full config: %+v", fullConfig)

	// Check that keys are correctly managed
	keysSection, ok := fullConfig["keys"].(map[string]interface{})
	require.True(t, ok, "keys section should exist")

	normalKeys, ok := keysSection["normal"].(map[string]interface{})
	require.True(t, ok, "normal keys should exist")
	assert.Len(t, normalKeys, 1) // only unmanaged
	assert.Equal(t, "extend_line_below", normalKeys["C-x"])

	insertKeys, ok := keysSection["insert"].(map[string]interface{})
	require.True(t, ok, "insert keys should exist")
	assert.Len(t, insertKeys, 1) // only managed
	assert.Equal(t, "yank", insertKeys["C-c"])

	// Check that other sections are preserved
	// Verify theme is preserved
	assert.Equal(t, "onedark", fullConfig["theme"])

	// Verify editor section is preserved
	editor, ok := fullConfig["editor"].(map[string]interface{})
	require.True(t, ok, "editor should be a map")
	assert.Equal(t, "relative", editor["line-number"])
	assert.Equal(t, false, editor["mouse"])

	// Check nested cursor-shape section
	cursorShape, ok := editor["cursor-shape"].(map[string]interface{})
	require.True(t, ok, "cursor-shape should be a map")
	assert.Equal(t, "bar", cursorShape["insert"])
	assert.Equal(t, "block", cursorShape["normal"])
}
