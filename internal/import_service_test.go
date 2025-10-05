package internal

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

// testPlugin implements pluginapi.Plugin interface for testing.
type testPlugin struct {
	editorType  pluginapi.EditorType
	configPath  string
	importData  *keymapv1.Keymap
	importError error
}

func newTestPlugin(
	editorType pluginapi.EditorType,
	configPath string,
	importData *keymapv1.Keymap,
	importError error,
) *testPlugin {
	return &testPlugin{
		editorType:  editorType,
		configPath:  configPath,
		importData:  importData,
		importError: importError,
	}
}

func (p *testPlugin) EditorType() pluginapi.EditorType {
	return p.editorType
}

func (p *testPlugin) ConfigDetect(opts pluginapi.ConfigDetectOptions) ([]string, bool, error) {
	return []string{p.configPath}, true, nil
}

func (p *testPlugin) Importer() (pluginapi.PluginImporter, error) {
	return &testPluginImporter{
		importData:  p.importData,
		importError: p.importError,
	}, nil
}

func (p *testPlugin) Exporter() (pluginapi.PluginExporter, error) {
	return &testPluginExporter{}, nil
}

// testPluginImporter implements pluginapi.PluginImporter interface for testing.
type testPluginImporter struct {
	importData  *keymapv1.Keymap
	importError error
}

func (i *testPluginImporter) Import(
	ctx context.Context,
	source io.Reader,
	opts pluginapi.PluginImportOption,
) (*keymapv1.Keymap, error) {
	return i.importData, i.importError
}

// testPluginExporter implements pluginapi.PluginExporter interface for testing.
type testPluginExporter struct{}

func (e *testPluginExporter) Export(
	ctx context.Context,
	destination io.Writer,
	setting *keymapv1.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	return &pluginapi.PluginExportReport{}, nil
}

func TestImportService_Import(t *testing.T) {
	testCases := []struct {
		name        string
		importData  *keymapv1.Keymap
		baseData    *keymapv1.Keymap
		importError error
		expectError bool
		expect      *importapi.ImportResult
	}{
		{
			name: "sorts imported keymaps by action ID",
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
						},
					},
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.copy",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Copy",
								Description: "Copy",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+c").KeyChords,
									KeyChordsReadable: "ctrl+c",
								},
							},
						},
						{
							Name: "actions.editor.paste",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Paste",
								Description: "Paste",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+v").KeyChords,
									KeyChordsReadable: "ctrl+v",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty keybindings add",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					{
						Name:     "actions.editor.paste",
						Bindings: []*keymapv1.Binding{},
					},
				},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.paste",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Paste",
								Description: "Paste",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+v").KeyChords,
									KeyChordsReadable: "ctrl+v",
								},
							},
						},
					},
				},
			},
		},

		{
			name:       "handles empty keymap list",
			importData: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{}},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{}},
				Changes: &importapi.KeymapChanges{},
			},
		},
		{
			name:        "handles nil setting from plugin",
			importData:  nil,
			expect:      nil,
			expectError: true,
		},
		{
			name: "calculates no change",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{},
			},
		},
		{
			name: "deduplicate",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{},
			},
		},
		{
			name: "calculates added keybindings",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.paste",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Paste",
								Description: "Paste",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+v").KeyChords,
									KeyChordsReadable: "ctrl+v",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "calculates not removed keybindings",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					// According to import semantics, unchanged keybindings should not be removed.
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{},
			},
		},
		{
			name: "calculates updated keybindings",
			baseData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.Keymap{
				Keybindings: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "cmd+c", "alt+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.Keymap{Keybindings: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
							{KeyChords: keymap.MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "cmd+c"},
							{KeyChords: keymap.MustParseKeyBinding("alt+c").KeyChords, KeyChordsReadable: "alt+c"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Update: []importapi.KeymapDiff{{
						Before: &keymapv1.Action{
							Name: "actions.editor.copy",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Copy",
								Description: "Copy",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+c").KeyChords,
									KeyChordsReadable: "ctrl+c",
								},
							},
						},
						After: &keymapv1.Action{
							Name: "actions.editor.copy",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Copy",
								Description: "Copy",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.Binding{
								{
									KeyChords:         keymap.MustParseKeyBinding("ctrl+c").KeyChords,
									KeyChordsReadable: "ctrl+c",
								},
								{KeyChords: keymap.MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "cmd+c"},
								{KeyChords: keymap.MustParseKeyBinding("alt+c").KeyChords, KeyChordsReadable: "alt+c"},
							},
						},
					}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup common test dependencies
			testFile, err := os.CreateTemp(t.TempDir(), "test_config_*.json")
			require.NoError(t, err)
			defer func() { _ = os.Remove(testFile.Name()); _ = testFile.Close() }()
			_, err = testFile.WriteString(`{}`)
			require.NoError(t, err)

			testPlug := newTestPlugin(pluginapi.EditorTypeVSCode, testFile.Name(), tc.importData, tc.importError)
			registry := plugins.NewRegistry()
			registry.Register(testPlug)

			mappingConfig := &mappings.MappingConfig{
				Mappings: map[string]mappings.ActionMappingConfig{
					"actions.editor.copy": {
						ID:          "actions.editor.copy",
						Description: "Copy",
						Name:        "Copy",
						Category:    "Editor",
					},
					"actions.editor.paste": {
						ID:          "actions.editor.paste",
						Description: "Paste",
						Name:        "Paste",
						Category:    "Editor",
					},
					"actions.file.save": {
						ID:          "actions.file.save",
						Description: "Save",
						Name:        "Save",
						Category:    "File",
					},
					"actions.editor.cut": {
						ID:          "actions.editor.cut",
						Description: "Cut",
						Name:        "Cut",
						Category:    "Editor",
					},
				},
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

			opts := importapi.ImportOptions{
				EditorType:  pluginapi.EditorTypeVSCode,
				InputStream: testFile,
				Base:        tc.baseData,
			}

			// Execute the import
			res, err := service.Import(context.Background(), opts)

			// Assertions
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expect == nil {
				assert.Nil(t, res)
				return
			}

			require.NotNil(t, res)

			settingDiff := cmp.Diff(tc.expect.Setting, res.Setting, protocmp.Transform())
			assert.Empty(t, settingDiff)

			changesDiff := cmp.Diff(tc.expect.Changes, res.Changes, protocmp.Transform())
			assert.Empty(t, changesDiff)
		})
	}
}
