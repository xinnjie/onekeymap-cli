package vscode

import (
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
	"google.golang.org/protobuf/proto"
)

func TestImporter_Import(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	importer, err := plugin.Importer()
	require.NoError(t, err)

	tests := []struct {
		name        string
		jsonContent string
		expected    *keymapv1.KeymapSetting
		expectError bool
	}{
		{
			name: "Standard keybindings array",
			jsonContent: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
				},
			},
		},
		{
			name: "Bind one action to multiple keys",
			jsonContent: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				},
				{
					"key": "ctrl+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
					keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
				},
			},
		},
		{
			name: "multiple key chord(cmd+k up)",
			jsonContent: `// This is a file-level comment
			[
			    // This is a keybinding comment
			    {
			        "key": "cmd+k up",
			        "command": "editor.action.clipboardCopyAction",
			        "when": "editorTextFocus"
			    }
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+k up"),
				},
			},
		},
		{
			name: "multiple key chord(shift shift)",
			jsonContent: `// This is a file-level comment
			[
			    // This is a keybinding comment
			    {
			        "key": "shift shift",
			        "command": "editor.action.clipboardCopyAction",
			        "when": "editorTextFocus"
			    }
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "shift shift"),
				},
			},
		},

		{
			name: "Single keybinding object",
			jsonContent: `{
				"key": "cmd+c",
				"command": "editor.action.clipboardCopyAction",
				"when": "editorTextFocus"
			}`,
			expectError: true,
		},
		{
			name: "correctly imports and deduplicates actions",
			jsonContent: `[
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
		},
		{
			// TODO(xinnjie): Need to deduplicate, and show conflict report
			name: "Import conflict mapping with multiple actions",
			jsonContent: `[
			{
					"key": "alt+1",
					"command": "command1",
					"when": "condition1"
			},
			{
					"key": "alt+3",
					"command": "command2",
					"when": "condition2"
			}
	]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+1"),
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
		},
		{
			name: "Command with args",
			jsonContent: `[
				{
					"key": "cmd+end",
					"command": "cursorEnd",
					"args": {
						"sticky": false
					}
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.withArgs", "meta+end"),
				},
			},
		},
		{
			name: "Mixed commands with and without args",
			jsonContent: `[
				{
					"key": "cmd+c",
					"command": "editor.action.clipboardCopyAction",
					"when": "editorTextFocus"
				},
				{
					"key": "cmd+end",
					"command": "cursorEnd",
					"args": {
						"sticky": false
					}
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
					keymap.NewActioinBinding("actions.test.withArgs", "meta+end"),
				},
			},
		},
		// ForImport behavior tests using importer.Import
		{
			name: "ForImport preference with args (ignoring when)",
			jsonContent: `[
				{
					"key": "alt+4",
					"command": "cmd.forimp.args",
					"when": "UNMATCHED",
					"args": {"x": 1}
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.forimport.withArgs.B", "alt+4"),
				},
			},
		},
		{
			name: "ForImport preference with no args (ignoring when)",
			jsonContent: `[
				{
					"key": "alt+5",
					"command": "cmd.forimp.noargs",
					"when": "MISMATCHED"
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.forimport.cmdOnly.B", "alt+5"),
				},
			},
		},
		{
			name: "Single-entry mapping implicitly forImport=true",
			jsonContent: `[
				{
					"key": "alt+6",
					"command": "cmd.forimp.single",
					"when": "UNMATCHED"
				}
			]`,
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.forimport.singleImplicit", "alt+6"),
				},
			},
		},
		{
			name: "Non-forImport entry should not import even when when matches",
			jsonContent: `[
				{
					"key": "alt+7",
					"command": "should.not.import.command",
					"when": "OtherCondition"
				}
			]`,
			expected: &keymapv1.KeymapSetting{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.jsonContent)
			result, err := importer.Import(context.Background(), reader, pluginapi.PluginImportOption{})

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expected, result), "Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
