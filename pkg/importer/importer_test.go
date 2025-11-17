package importer_test

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
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/importer"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

// testPlugin implements pluginapi.Plugin interface for testing.
type testPlugin struct {
	editorType  pluginapi2.EditorType
	configPath  string
	importData  *keymapv1.Keymap
	importError error
}

func newTestPlugin(
	editorType pluginapi2.EditorType,
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

func (p *testPlugin) EditorType() pluginapi2.EditorType {
	return p.editorType
}

func (p *testPlugin) ConfigDetect(_ pluginapi2.ConfigDetectOptions) ([]string, bool, error) {
	return []string{p.configPath}, true, nil
}

func (p *testPlugin) Importer() (pluginapi2.PluginImporter, error) {
	return &testPluginImporter{
		importData:  p.importData,
		importError: p.importError,
	}, nil
}

func (p *testPlugin) Exporter() (pluginapi2.PluginExporter, error) {
	return &testPluginExporter{}, nil
}

// testPluginImporter implements pluginapi.PluginImporter interface for testing.
type testPluginImporter struct {
	importData  *keymapv1.Keymap
	importError error
}

func (i *testPluginImporter) Import(
	_ context.Context,
	_ io.Reader,
	_ pluginapi2.PluginImportOption,
) (pluginapi2.PluginImportResult, error) {
	return pluginapi2.PluginImportResult{Keymap: i.importData}, i.importError
}

// testPluginExporter implements pluginapi.PluginExporter interface for testing.
type testPluginExporter struct{}

func (e *testPluginExporter) Export(
	_ context.Context,
	_ io.Writer,
	_ *keymapv1.Keymap,
	_ pluginapi2.PluginExportOption,
) (*pluginapi2.PluginExportReport, error) {
	return &pluginapi2.PluginExportReport{}, nil
}

func TestImportService_Import(t *testing.T) {
	testCases := []struct {
		name        string
		importData  *keymapv1.Keymap
		baseData    *keymapv1.Keymap
		importError error
		expectError bool
		expect      *importerapi.ImportResult
	}{
		{
			name: "sorts imported keymaps by action ID",
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
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
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.copy",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Copy",
								Description: "Copy",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.KeybindingReadable{
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
							Bindings: []*keymapv1.KeybindingReadable{
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
				Actions: []*keymapv1.Action{
					{
						Name:     "actions.editor.paste",
						Bindings: []*keymapv1.KeybindingReadable{},
					},
				},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.paste",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Paste",
								Description: "Paste",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.KeybindingReadable{
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
			importData: &keymapv1.Keymap{Actions: []*keymapv1.Action{}},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{}},
				Changes: &importerapi.KeymapChanges{},
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
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "deduplicate",
			baseData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "calculates added keybindings",
			baseData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.paste",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Paste",
							Description: "Paste",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []*keymapv1.Action{
						{
							Name: "actions.editor.paste",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Paste",
								Description: "Paste",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.KeybindingReadable{
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
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					// According to import semantics, unchanged keybindings should not be removed.
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "calculates updated keybindings",
			baseData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.Keymap{
				Actions: []*keymapv1.Action{
					keymap.NewActioinBinding("actions.editor.copy", "cmd+c", "alt+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: &keymapv1.Keymap{Actions: []*keymapv1.Action{
					{
						Name: "actions.editor.copy",
						ActionConfig: &keymapv1.ActionConfig{
							DisplayName: "Copy",
							Description: "Copy",
							Category:    "Editor",
						},
						Bindings: []*keymapv1.KeybindingReadable{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
							{KeyChords: keymap.MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "cmd+c"},
							{KeyChords: keymap.MustParseKeyBinding("alt+c").KeyChords, KeyChordsReadable: "alt+c"},
						},
					},
				}},
				Changes: &importerapi.KeymapChanges{
					Update: []importerapi.KeymapDiff{{
						Before: &keymapv1.Action{
							Name: "actions.editor.copy",
							ActionConfig: &keymapv1.ActionConfig{
								DisplayName: "Copy",
								Description: "Copy",
								Category:    "Editor",
							},
							Bindings: []*keymapv1.KeybindingReadable{
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
							Bindings: []*keymapv1.KeybindingReadable{
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

			testPlug := newTestPlugin(pluginapi2.EditorTypeVSCode, testFile.Name(), tc.importData, tc.importError)
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
			service := importer.NewImporter(registry, mappingConfig, logger, metrics.NewNoop())

			opts := importerapi.ImportOptions{
				EditorType:  pluginapi2.EditorTypeVSCode,
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
