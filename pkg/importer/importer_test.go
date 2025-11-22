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
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/importer"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
)

// testPlugin implements pluginapi.Plugin interface for testing.
type testPlugin struct {
	editorType  pluginapi.EditorType
	configPath  string
	importData  keymap.Keymap
	importError error
}

func newTestPlugin(
	editorType pluginapi.EditorType,
	configPath string,
	importData keymap.Keymap,
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

func (p *testPlugin) ConfigDetect(_ pluginapi.ConfigDetectOptions) ([]string, bool, error) {
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
	importData  keymap.Keymap
	importError error
}

func (i *testPluginImporter) Import(
	_ context.Context,
	_ io.Reader,
	_ pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	return pluginapi.PluginImportResult{Keymap: i.importData}, i.importError
}

// testPluginExporter implements pluginapi.PluginExporter interface for testing.
type testPluginExporter struct{}

func (e *testPluginExporter) Export(
	_ context.Context,
	_ io.Writer,
	_ keymap.Keymap,
	_ pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	return &pluginapi.PluginExportReport{}, nil
}

// Helper function to create test action with bindings
func newAction(name string, bindings ...string) keymap.Action {
	action := keymap.Action{Name: name}
	for _, b := range bindings {
		kb, err := keybinding.NewKeybinding(b, keybinding.ParseOption{Separator: "+"})
		if err != nil {
			panic(err)
		}
		action.Bindings = append(action.Bindings, kb)
	}
	return action
}

func TestImportService_Import(t *testing.T) {
	testCases := []struct {
		name        string
		importData  keymap.Keymap
		baseData    keymap.Keymap
		importError error
		expectError bool
		expect      *importerapi.ImportResult
	}{
		{
			name: "sorts imported keymaps by action ID",
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
					newAction("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.copy", "ctrl+c"),
					newAction("actions.editor.paste", "ctrl+v"),
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []keymap.Action{
						newAction("actions.editor.copy", "ctrl+c"),
						newAction("actions.editor.paste", "ctrl+v"),
					},
				},
			},
		},
		{
			name: "empty keybindings add",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{
					{Name: "actions.editor.paste", Bindings: []keybinding.Keybinding{}},
				},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []keymap.Action{
						newAction("actions.editor.paste", "ctrl+v"),
					},
				},
			},
		},

		{
			name:       "handles empty keymap list",
			importData: keymap.Keymap{Actions: []keymap.Action{}},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name:        "handles nil setting from plugin",
			importData:  keymap.Keymap{},
			expect:      nil,
			expectError: true,
		},
		{
			name: "calculates no change",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "deduplicate",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "calculates added keybindings",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.paste", "ctrl+v"),
				}},
				Changes: &importerapi.KeymapChanges{
					Add: []keymap.Action{
						newAction("actions.editor.paste", "ctrl+v"),
					},
				},
			},
		},
		{
			name: "calculates not removed keybindings",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					// According to import semantics, unchanged keybindings should not be removed.
					newAction("actions.editor.copy", "ctrl+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.copy", "ctrl+c"),
				}},
				Changes: &importerapi.KeymapChanges{},
			},
		},
		{
			name: "calculates updated keybindings",
			baseData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.copy", "ctrl+c"),
				},
			},
			importData: keymap.Keymap{
				Actions: []keymap.Action{
					newAction("actions.editor.copy", "cmd+c", "alt+c"),
				},
			},
			expect: &importerapi.ImportResult{
				Setting: keymap.Keymap{Actions: []keymap.Action{
					newAction("actions.editor.copy", "ctrl+c", "cmd+c", "alt+c"),
				}},
				Changes: &importerapi.KeymapChanges{
					Update: []importerapi.KeymapDiff{{
						Before: newAction("actions.editor.copy", "ctrl+c"),
						After:  newAction("actions.editor.copy", "ctrl+c", "cmd+c", "alt+c"),
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
			registry := registry.NewRegistry()
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

			settingDiff := cmp.Diff(tc.expect.Setting, res.Setting)
			assert.Empty(t, settingDiff, "Setting mismatch: %s", settingDiff)

			changesDiff := cmp.Diff(tc.expect.Changes, res.Changes)
			assert.Empty(t, changesDiff, "Changes mismatch: %s", changesDiff)
		})
	}
}
