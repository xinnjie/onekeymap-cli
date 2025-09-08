package onekeymap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	vscodeplugin "github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

// testPlugin implements pluginapi.Plugin interface for testing
type testPlugin struct {
	editorType  pluginapi.EditorType
	configPath  string
	importData  *keymapv1.KeymapSetting
	importError error
}

func TestImportEndToEnd_Import_VSCode_FormatSelection_NoChange(t *testing.T) {
	// Setup mapping config according to provided YAML
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.edit.formatSelection": {
				ID:          "actions.edit.formatSelection",
				Name:        "Format selection",
				Description: "Format Selection",
				Category:    "Editor",
				VSCode: mappings.VscodeConfigs{
					{
						EditorActionMapping: mappings.EditorActionMapping{ForImport: true},
						Command:             "editor.action.formatSelection",
						When:                "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly",
					},
					{
						Command: "notebook.formatCell",
						When:    "editorHasDocumentFormattingProvider && editorTextFocus && inCompositeEditor && notebookEditable && !editorReadonly && activeEditor == 'workbench.editor.notebook'",
					},
				},
			},
		},
	}

	// Use real VSCode plugin importer for this scenario
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := plugins.NewRegistry()
	registry.Register(vscodeplugin.New(mappingConfig, logger))
	service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

	// VSCode keybindings.json content (comments stripped by importer)
	// Note that "ctrl+alt+shift+l" is different from "ctrl+shift+alt+l"
	vscodeJSON := []byte(`[
  {
    "key": "ctrl+alt+shift+l",
    "command": "editor.action.formatSelection",
    "when": "editorHasDocumentSelectionFormattingProvider && editorTextFocus && !editorReadonly"
  }
]`)

	// Base config has the same binding (order of modifiers irrelevant; parser normalizes)
	base := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.formatSelection", "ctrl+shift+alt+l"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: bytes.NewReader(vscodeJSON),
		Base:        base,
	}

	res, err := service.Import(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, res)

	expected := &importapi.ImportResult{
		Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
			{
				Id:          "actions.edit.formatSelection",
				Name:        "Format selection",
				Description: "Format Selection",
				Category:    "Editor",
				Bindings: []*keymapv1.Binding{
					{KeyChords: keymap.MustParseKeyBinding("ctrl+shift+alt+l").KeyChords, KeyChordsReadable: "ctrl+shift+alt+l"},
				},
			},
		}},
		Changes: &importapi.KeymapChanges{},
	}

	settingDiff := cmp.Diff(expected.Setting, res.Setting, protocmp.Transform())
	assert.Empty(t, settingDiff)
	changesDiff := cmp.Diff(expected.Changes, res.Changes, protocmp.Transform())
	assert.Empty(t, changesDiff)
}

func newTestPlugin(editorType pluginapi.EditorType, configPath string, importData *keymapv1.KeymapSetting, importError error) *testPlugin {
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

func (p *testPlugin) DefaultConfigPath(opts ...pluginapi.DefaultConfigPathOption) ([]string, error) {
	return []string{p.configPath}, nil
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

// testPluginImporter implements pluginapi.PluginImporter interface for testing
type testPluginImporter struct {
	importData  *keymapv1.KeymapSetting
	importError error
}

func (i *testPluginImporter) Import(ctx context.Context, source io.Reader, opts pluginapi.PluginImportOption) (*keymapv1.KeymapSetting, error) {
	return i.importData, i.importError
}

// testPluginExporter implements pluginapi.PluginExporter interface for testing
type testPluginExporter struct{}

func (e *testPluginExporter) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts pluginapi.PluginExportOption) (*pluginapi.PluginExportReport, error) {
	return &pluginapi.PluginExportReport{}, nil
}

func TestImportService_Import(t *testing.T) {
	testCases := []struct {
		name        string
		importData  *keymapv1.KeymapSetting
		baseData    *keymapv1.KeymapSetting
		importError error
		expectError bool
		expect      *importapi.ImportResult
	}{
		{
			name: "sorts imported keymaps by action ID",
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.copy",
						Name:        "Copy",
						Description: "Copy",
						Category:    "Editor",
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
						},
					},
					{
						Id:          "actions.editor.paste",
						Name:        "Paste",
						Description: "Paste",
						Category:    "Editor",
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Add: []*keymapv1.ActionBinding{
						{
							Id:          "actions.editor.copy",
							Name:        "Copy",
							Description: "Copy",
							Category:    "Editor",
							Bindings: []*keymapv1.Binding{
								{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
							},
						},
						{
							Id:          "actions.editor.paste",
							Name:        "Paste",
							Description: "Paste",
							Category:    "Editor",
							Bindings: []*keymapv1.Binding{
								{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
							},
						},
					},
				},
			},
		},
		{
			name:       "handles empty keymap list",
			importData: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{}},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{}},
				Changes: &importapi.KeymapChanges{},
			},
		},
		{
			name:       "handles nil setting from plugin",
			importData: nil,
			expect:     nil,
		},
		{
			name: "calculates no change",
			baseData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.paste",
						Name:        "Paste",
						Description: "Paste",
						Category:    "Editor",
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
			baseData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.paste",
						Name:        "Paste",
						Description: "Paste",
						Category:    "Editor",
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
			baseData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{},
			},
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.paste",
						Name:        "Paste",
						Description: "Paste",
						Category:    "Editor",
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Add: []*keymapv1.ActionBinding{
						{
							Id:          "actions.editor.paste",
							Name:        "Paste",
							Description: "Paste",
							Category:    "Editor",
							Bindings: []*keymapv1.Binding{
								{KeyChords: keymap.MustParseKeyBinding("ctrl+v").KeyChords, KeyChordsReadable: "ctrl+v"},
							},
						},
					},
				},
			},
		},
		{
			name: "calculates not removed keybindings",
			baseData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					// According to import semantics, unchanged keybindings should not be removed.
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.copy",
						Name:        "Copy",
						Description: "Copy",
						Category:    "Editor",
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
			baseData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: &keymapv1.KeymapSetting{
				Keybindings: []*keymapv1.ActionBinding{
					keymap.NewActioinBinding("actions.editor.copy", "cmd+c", "alt+c"),
				},
			},
			expect: &importapi.ImportResult{
				Setting: &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{
					{
						Id:          "actions.editor.copy",
						Name:        "Copy",
						Description: "Copy",
						Category:    "Editor",
						Bindings: []*keymapv1.Binding{
							{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
							{KeyChords: keymap.MustParseKeyBinding("cmd+c").KeyChords, KeyChordsReadable: "cmd+c"},
							{KeyChords: keymap.MustParseKeyBinding("alt+c").KeyChords, KeyChordsReadable: "alt+c"},
						},
					},
				}},
				Changes: &importapi.KeymapChanges{
					Update: []importapi.KeymapDiff{{
						Before: &keymapv1.ActionBinding{
							Id:          "actions.editor.copy",
							Name:        "Copy",
							Description: "Copy",
							Category:    "Editor",
							Bindings: []*keymapv1.Binding{
								{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
							},
						},
						After: &keymapv1.ActionBinding{
							Id:          "actions.editor.copy",
							Name:        "Copy",
							Description: "Copy",
							Category:    "Editor",
							Bindings: []*keymapv1.Binding{
								{KeyChords: keymap.MustParseKeyBinding("ctrl+c").KeyChords, KeyChordsReadable: "ctrl+c"},
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
			testFile, err := os.CreateTemp("", "test_config_*.json")
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
