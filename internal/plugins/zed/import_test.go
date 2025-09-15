package zed

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

func TestImportZedKeymap(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	if err != nil {
		t.Fatal(err)
	}
	plugin := New(mappingConfig, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	testCases := []struct {
		name      string
		input     string
		expected  *keymapv1.KeymapSetting
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
					keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
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
			expected:  nil,
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.edit.copy", "meta+c"),
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+1"),
					keymap.NewActioinBinding("actions.test.mutipleActions", "alt+3"),
				},
			},
		},
		{
			name:      "Invalid keychord",
			input:     `[{"context": "Editor", "bindings": {"invalid-key": "editor::Save"}}]`,
			expected:  &keymapv1.KeymapSetting{},
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
			expected: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.test.withArgs", "meta+shift+t"),
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
			expected:  &keymapv1.KeymapSetting{},
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
			expected:  &keymapv1.KeymapSetting{},
			expectErr: false,
		},
		{
			name:      "Malformed JSON",
			input:     `[{"context": "Editor", "bindings": {"cmd-s": "editor::Save"}`,
			expected:  nil,
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
				assert.Truef(t, proto.Equal(tc.expected, result), "Expected and actual KeymapSetting should be equal, expect %s, got %s", tc.expected.String(), result.String())
			}
		})
	}
}
