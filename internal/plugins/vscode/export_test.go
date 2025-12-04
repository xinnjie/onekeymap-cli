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
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestExporter_Export(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	tests := []struct {
		name           string
		keymapSetting  keymap.Keymap
		expectedJSON   string
		existingConfig string
	}{
		// Basic destructive export tests
		{
			name: "correctly exports a standard action",
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("alt+3")},
					},
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
		{
			name: "falls back to child action when parent not supported",
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.parentNotSupported",
						Bindings: []keybinding.Keybinding{parseKB("meta+shift+h")},
					},
				},
			},
			expectedJSON: `[
			  {
			    "key": "cmd+shift+h",
			    "command": "child.supported.command",
			    "when": "editorTextFocus"
			  }
			]`,
		},
		{
			name: "both parent and child have keybindings - each exports to child command",
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.parentNotSupported",
						Bindings: []keybinding.Keybinding{parseKB("meta+shift+h")},
					},
					{
						Name:     "actions.test.childSupported",
						Bindings: []keybinding.Keybinding{parseKB("meta+shift+j")},
					},
				},
			},
			expectedJSON: `[
			  {
			    "key": "cmd+shift+h",
			    "command": "child.supported.command",
			    "when": "editorTextFocus"
			  },
			  {
			    "key": "cmd+shift+j",
			    "command": "child.supported.command",
			    "when": "editorTextFocus"
			  }
			]`,
		},
		// Non-destructive export tests
		{
			name: "non-destructive export preserves user keybindings",
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("alt+3")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
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
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.edit.copy",
						Bindings: []keybinding.Keybinding{parseKB("meta+c")},
					},
					{
						Name:     "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{parseKB("alt+3")},
					},
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
			opts := pluginapi.PluginExportOption{
				ExistingConfig: nil,
				TargetPlatform: platform.PlatformMacOS, // Use macOS for consistent test results across platforms
			}

			if tt.existingConfig != "" {
				opts.ExistingConfig = strings.NewReader(tt.existingConfig)
			}

			_, err = exporter.Export(context.Background(), &buf, tt.keymapSetting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.expectedJSON, buf.String())
		})
	}
}

func TestExporter_Export_VSCodeVariant(t *testing.T) {
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
		name          string
		editorType    pluginapi.EditorType
		keymapSetting keymap.Keymap
		expectedJSON  string
	}{
		{
			name:       "windsurf uses windsurf-specific config",
			editorType: pluginapi.EditorTypeWindsurf,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+m",
					"command": "windsurf.specific.command",
					"when": "windsurf.cascadePanel.focused"
				}
			]`,
		},
		{
			name:       "cursor uses cursor-specific config",
			editorType: pluginapi.EditorTypeCursor,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+m",
					"command": "cursor.specific.command",
					"when": "cursorContext"
				}
			]`,
		},
		{
			name:       "vscode uses vscode config (no variant override)",
			editorType: pluginapi.EditorTypeVSCode,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+m",
					"command": "vscode.default.command",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name:       "windsurf falls back to vscode config when no windsurf config",
			editorType: pluginapi.EditorTypeWindsurf,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantFallback",
						Bindings: []keybinding.Keybinding{parseKB("meta+f")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+f",
					"command": "vscode.fallback.command",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name:       "cursor falls back to vscode config when no cursor config",
			editorType: pluginapi.EditorTypeCursor,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantFallback",
						Bindings: []keybinding.Keybinding{parseKB("meta+f")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+f",
					"command": "vscode.fallback.command",
					"when": "editorTextFocus"
				}
			]`,
		},
		{
			name:       "windsurf-next uses windsurf config",
			editorType: pluginapi.EditorTypeWindsurfNext,
			keymapSetting: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name:     "actions.test.variantConfig",
						Bindings: []keybinding.Keybinding{parseKB("meta+m")},
					},
				},
			},
			expectedJSON: `[
				{
					"key": "cmd+m",
					"command": "windsurf.specific.command",
					"when": "windsurf.cascadePanel.focused"
				}
			]`,
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

			exporter, err := plugin.Exporter()
			require.NoError(t, err)

			var buf bytes.Buffer
			opts := pluginapi.PluginExportOption{
				ExistingConfig: nil,
				TargetPlatform: platform.PlatformMacOS, // Use macOS for consistent test results across platforms
			}

			_, err = exporter.Export(context.Background(), &buf, tt.keymapSetting, opts)
			require.NoError(t, err)

			assert.JSONEq(t, tt.expectedJSON, buf.String())
		})
	}
}
