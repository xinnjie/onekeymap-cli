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
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func TestExporter_Export(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	tests := []struct {
		name           string
		keymapSetting  *keymapv1.Keymap
		expectedJSON   string
		existingConfig string
	}{
		// Basic destructive export tests
		{
			name: "correctly exports a standard action",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			expectedJSON: `[
			  {
			    "key": "cmd+c",
			    "command": "editor.action.clipboardCopyAction",
			    "when": "editorTextFocus && condition > 0"
			  }
			]`,
		},
		{
			name: "correctly exports multiple actions",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
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
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
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
					"when": "editorTextFocus && condition > 0"
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
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
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
					"when": "editorTextFocus && condition > 0"
				}
			]`,
		},
		{
			name: "multiple user keybindings with mixed conflicts",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
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
					"when": "editorTextFocus && condition > 0"
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
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus && condition > 0"
				}
			]`,
		},
		{
			name: "existing config with trailing commas and comments",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: `[
				// User custom keybinding
				{
					"key": "cmd+x",
					"command": "custom.user.command",
					"when": "editorTextFocus", // trailing comma here
				}, // trailing comma after object
				{
					// Another user keybinding
					"key": "cmd+v",
					"command": "custom.paste.command",
					"when": "editorTextFocus",
				}, // final trailing comma
			]`,
			expectedJSON: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus && condition > 0"
				},
				{
					"key": "cmd+x",
					"command": "custom.user.command",
					"when": "editorTextFocus"
				},
				{
					"key": "cmd+v",
					"command": "custom.paste.command",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name: "handles empty existing config file",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
			existingConfig: ``, // Represents an empty file
			expectedJSON: `[
      {
        "key": "cmd+c",
        "command": "editor.action.clipboardCopyAction",
        "when": "editorTextFocus && condition > 0"
      }
    ]`,
		},
		// Base order preservation tests
		{
			name: "preserves order by base config using command as key",
			keymapSetting: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
			// Base config defines the desired order by command: command2, copy, command1, custom.undo
			existingConfig: `[
				{"key":"x","command":"command2"},
				{"key":"x","command":"editor.action.clipboardCopyAction"},
				{"key":"x","command":"command1"},
				{"key":"x","command":"custom.undo.command"}
			]`,
			expectedJSON: `[
				{
					"key": "alt+3",
					"command": "command2",
					"when": "condition2"
				},
				{
					"key": "x",
					"command": "command2"
				},
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus && condition > 0"
				},
				{
					"key": "x",
					"command": "editor.action.clipboardCopyAction"
				},
				{
					"key": "alt+3",
					"command": "command1",
					"when": "condition1"
				},
				{
					"key": "x",
					"command": "command1"
				},
				{
					"key": "x",
					"command": "custom.undo.command"
				}
			]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())
			exporter, err := plugin.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{ExistingConfig: nil}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.keymapSetting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.expectedJSON, buf.String())
		})
	}
}
