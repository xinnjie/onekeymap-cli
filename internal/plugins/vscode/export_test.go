package vscode

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestExporter_Export(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.KeymapSetting
		expectedJSON   string
		existingConfig string
	}{
		// Basic destructive export tests
		{
			name: "correctly exports a standard action",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			expectedJSON: `[
			  {
			    "key": "cmd+c",
			    "command": "editor.action.clipboardCopyAction",
			    "when": "editorTextFocus"
			  }
			]`,
		},
		{
			name: "correctly exports multiple actions",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
			expectedJSON: `[
			  {
			    "key": "alt+3",
			    "command": "command1",
			    "when": "condition1"
			  },
			  {
			    "key": "alt+3",
			    "command": "command2",
			    "when": "condition2"
			  }
			]`,
		},
		// Non-destructive export tests
		{
			name: "non-destructive export preserves user keybindings",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
				{
					"key": "cmd+x",
					"command": "custom.user.command",
					"when": "editorTextFocus"
				}
			]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				},
				{
					"key": "cmd+x",
					"command": "custom.user.command",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name: "managed keybinding takes priority over conflicting user keybinding",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
				{
					"key": "cmd+c",
					"command": "custom.user.command",
					"when": "editorTextFocus"
				}
			]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name: "multiple user keybindings with mixed conflicts",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
			existingConfig: `[
				{
					"key": "cmd+c",
					"command": "custom.conflicting.command"
				},
				{
					"key": "cmd+v",
					"command": "custom.paste.command",
					"when": "editorTextFocus"
				},
				{
					"key": "cmd+z",
					"command": "custom.undo.command"
				}
			]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				},
				{
					"key": "alt+3",
					"command": "command1",
					"when": "condition1"
				},
				{
					"key": "alt+3",
					"command": "command2",
					"when": "condition2"
				},
				{
					"key": "cmd+v",
					"command": "custom.paste.command",
					"when": "editorTextFocus"
				},
				{
					"key": "cmd+z",
					"command": "custom.undo.command"
				}
			]`,
		},
		{
			name: "empty existing config behaves as destructive export",
			keymapSetting: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				}
			]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)))
			exporter, err := plugin.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{Base: nil}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.keymapSetting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.expectedJSON, buf.String())
		})
	}
}
