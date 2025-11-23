package zed

import (
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

func TestImportZedKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	if err != nil {
		t.Fatal(err)
	}
	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)), metrics.NewNoop())

	parseKB := func(s string) keybinding.Keybinding {
		kb, err := keybinding.NewKeybinding(s, keybinding.ParseOption{Platform: platform.PlatformMacOS, Separator: "+"})
		if err != nil {
			panic(err)
		}
		return kb
	}

	testCases := []struct {
		name      string
		input     string
		expected  keymap.Keymap
		expectErr bool
	}{
		{
			name: "Standard JSON array",
			input: `[
				{
					"context": "Editor",
					"bindings": {
						"cmd-c": "editor::Copy"
					}
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Bind one action to multiple keys",
			input: `[
                {
                    "context": "Editor",
                    "bindings": {
                        "cmd-c": "editor::Copy",
                        "ctrl-c": "editor::Copy"
                    }
                }
            ]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
							parseKB("ctrl+c"),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Single JSON object should fail",
			input: `{
					"context": "Workspace",
					"bindings": {
						"cmd-s": "workspace::Save"
					}
				}`,
			expected:  keymap.Keymap{},
			expectErr: true,
		},
		{
			name: "JSON with comments",
			input: `[
				// This is a comment
				{
					"context": "Editor",
					"bindings": {
						"cmd-c": "editor::Copy" // another comment
					}
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.edit.copy",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+c"),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "correctly imports and deduplicates actions",
			input: `[
				{
					"context": "context1",
					"bindings": {
						"alt-3": "command1"
					}
				},
				{
					"context": "context2",
					"bindings": {
						"alt-3": "command2"
					}
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{
							parseKB("alt+3"),
						},
					},
				},
			},
		},
		{
			// TODO(xinnjie): Need to deduplicate, and show conflict report
			name: "Import conflict mapping with multiple actions",
			input: `[
			{
					"context": "context1",
					"bindings": {
						"alt-1": "command1"
					}
			},
			{
					"context": "context2",
					"bindings": {
						"alt-3": "command2"
					}
			}
	]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.mutipleActions",
						Bindings: []keybinding.Keybinding{
							parseKB("alt+1"),
							parseKB("alt+3"),
						},
					},
				},
			},
		},
		{
			name:      "Invalid keychord",
			input:     `[{"context": "Editor", "bindings": {"invalid-key": "editor::Save"}}]`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name: "Action with args (array format)",
			input: `[
				{
					"context": "Editor",
					"bindings": {
						"cmd-shift-t": ["test::ActionWithArgs", {"test_param": true}]
					}
				}
			]`,
			expected: keymap.Keymap{
				Actions: []keymap.Action{
					{
						Name: "actions.test.withArgs",
						Bindings: []keybinding.Keybinding{
							parseKB("meta+shift+t"),
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid array format (empty array)",
			input: `[
				{
					"context": "Editor",
					"bindings": {
						"cmd-x": []
					}
				}
			]`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name: "Invalid array format (non-string action)",
			input: `[
				{
					"context": "Editor",
					"bindings": {
						"cmd-x": [123, {"arg": "value"}]
					}
				}
			]`,
			expected:  keymap.Keymap{},
			expectErr: false,
		},
		{
			name:      "Malformed JSON",
			input:     `[{"context": "Editor", "bindings": {"cmd-s": "editor::Save"}`,
			expected:  keymap.Keymap{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			importer, err := plugin.Importer()
			require.NoError(t, err)
			result, err := importer.Import(context.Background(), reader, pluginapi.PluginImportOption{})

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result.Keymap)
			}
		})
	}
}
