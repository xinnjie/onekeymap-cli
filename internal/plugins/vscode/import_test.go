package vscode

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestImporter_Import(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())
	importer, err := plugin.Importer()
	require.NoError(t, err)

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name        string
		jsonContent string
		expected    keymap.Keymap
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c"), parseKB("ctrl+c")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+k up")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("shift shift")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("alt+3")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("alt+1"), parseKB("alt+3")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.withArgs",
						Bindings: []keybinding.Keybinding{parseKB("meta+end")},
					},
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
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
					{
						Name:     "actions.test.withArgs",
						Bindings: []keybinding.Keybinding{parseKB("meta+end")},
					},
				},
			},
		},
		// DisableImport behavior tests using importer.Import
		{
			name: "DisableImport: prefer enabled config with args (ignoring when)",
			jsonContent: `[
				{
					"key": "alt+4",
					"command": "cmd.forimp.args",
					"when": "UNMATCHED",
					"args": {"x": 1}
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.forimport.withArgs.B",
						Bindings: []keybinding.Keybinding{parseKB("alt+4")},
					},
				},
			},
		},
		{
			name: "DisableImport: prefer enabled config with no args (ignoring when)",
			jsonContent: `[
				{
					"key": "alt+5",
					"command": "cmd.forimp.noargs",
					"when": "MISMATCHED"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.forimport.cmdOnly.B",
						Bindings: []keybinding.Keybinding{parseKB("alt+5")},
					},
				},
			},
		},
		{
			name: "Single-entry mapping implicitly enabled for import",
			jsonContent: `[
				{
					"key": "alt+6",
					"command": "cmd.forimp.single",
					"when": "UNMATCHED"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.forimport.singleImplicit",
						Bindings: []keybinding.Keybinding{parseKB("alt+6")},
					},
				},
			},
		},
		{
			name: "DisableImport entry should not import even when when matches",
			jsonContent: `[
				{
					"key": "alt+7",
					"command": "should.not.import.command",
					"when": "OtherCondition"
				}
			]`,
			expected: keymap.Keymap{},
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
				assert.True(t, reflect.DeepEqual(tt.expected, result.Keymap), "Expected %v, got %v", tt.expected, result.Keymap)
			}
		})
	}
}

func TestImporter_Import_VSCodeVariant(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	// Inline the variant-specific test actions instead of sourcing them from the YAML fixture.
	mappingConfig.Mappings["actions.test.variantConfig"] = mappings.ActionMappingConfig{
		ID:          "actions.test.variantConfig",
		Description: "Test VSCode variant config priority",
		Category:    "Testing",
		VSCode: mappings.VscodeConfigs{
			{Command: "vscode.default.command", When: "editorTextFocus"},
		},
		Windsurf: mappings.VscodeConfigs{
			{Command: "windsurf.specific.command", When: "windsurf.cascadePanel.focused"},
		},
		Cursor: mappings.VscodeConfigs{
			{Command: "cursor.specific.command", When: "cursorContext"},
		},
	}

	mappingConfig.Mappings["actions.test.variantFallback"] = mappings.ActionMappingConfig{
		ID:          "actions.test.variantFallback",
		Description: "Test VSCode variant fallback to vscode config",
		Category:    "Testing",
		VSCode: mappings.VscodeConfigs{
			{Command: "vscode.fallback.command", When: "editorTextFocus"},
		},
	}

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name        string
		editorType  pluginapi.EditorType
		jsonContent string
		expected    keymap.Keymap
	}{
		{
			name:       "windsurf uses windsurf-specific config",
			editorType: pluginapi.EditorTypeWindsurf,
			jsonContent: `[
				{
					"key": "cmd+m",
					"command": "windsurf.specific.command",
					"when": "windsurf.cascadePanel.focused"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
		},
		{
			name:       "cursor uses cursor-specific config",
			editorType: pluginapi.EditorTypeCursor,
			jsonContent: `[
				{
					"key": "cmd+m",
					"command": "cursor.specific.command",
					"when": "cursorContext"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
		},
		{
			name:       "vscode uses vscode config (no variant override)",
			editorType: pluginapi.EditorTypeVSCode,
			jsonContent: `[
				{
					"key": "cmd+m",
					"command": "vscode.default.command",
					"when": "editorTextFocus"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
		},
		{
			name:       "windsurf falls back to vscode config when no windsurf config",
			editorType: pluginapi.EditorTypeWindsurf,
			jsonContent: `[
				{
					"key": "cmd+f",
					"command": "vscode.fallback.command",
					"when": "editorTextFocus"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantFallback",
						Bindings: []keybinding.Keybinding{parseKB("meta+f")},
					},
				},
			},
		},
		{
			name:       "cursor falls back to vscode config when no cursor config",
			editorType: pluginapi.EditorTypeCursor,
			jsonContent: `[
				{
					"key": "cmd+f",
					"command": "vscode.fallback.command",
					"when": "editorTextFocus"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantFallback",
						Bindings: []keybinding.Keybinding{parseKB("meta+f")},
					},
				},
			},
		},
		{
			name:       "windsurf-next uses windsurf config",
			editorType: pluginapi.EditorTypeWindsurfNext,
			jsonContent: `[
				{
					"key": "cmd+m",
					"command": "windsurf.specific.command",
					"when": "windsurf.cascadePanel.focused"
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			recorder := metrics.NewNoop()

			var plugin pluginapi.Plugin
			switch tt.editorType {
			case pluginapi.EditorTypeWindsurf:
				plugin = NewWindsurf(mappingConfig, logger, recorder)
			case pluginapi.EditorTypeWindsurfNext:
				plugin = NewWindsurfNext(mappingConfig, logger, recorder)
			case pluginapi.EditorTypeCursor:
				plugin = NewCursor(mappingConfig, logger, recorder)
			default:
				plugin = New(mappingConfig, logger, recorder)
			}

			importer, err := plugin.Importer()
			require.NoError(t, err)

			reader := strings.NewReader(tt.jsonContent)
			result, err := importer.Import(context.Background(), reader, pluginapi.PluginImportOption{})
			require.NoError(t, err)

			assert.True(
				t,
				reflect.DeepEqual(tt.expected, result.Keymap),
				"Expected %v, got %v",
				tt.expected,
				result.Keymap,
			)
		})
	}
}
