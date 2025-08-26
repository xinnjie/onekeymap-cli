package onekeymap

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// testPlugin implements pluginapi.Plugin interface for testing
type testPlugin struct {
	editorType  pluginapi.EditorType
	configPath  string
	importData  *keymapv1.KeymapSetting
	importError error
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

func (p *testPlugin) DefaultConfigPath() ([]string, error) {
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

func TestImportService_Import_SortsByAction(t *testing.T) {
	// Create test data with unsorted actions
	unsortedKeymaps := []*keymapv1.KeyBinding{
		keymap.NewBinding("actions.editor.paste", "ctrl+v"),
		keymap.NewBinding("actions.editor.copy", "ctrl+c"),
		keymap.NewBinding("actions.file.save", "ctrl+s"),
		keymap.NewBinding("actions.editor.cut", "ctrl+x"),
	}

	unsortedSetting := &keymapv1.KeymapSetting{
		Keybindings: unsortedKeymaps,
	}

	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_config_*.json")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(testFile.Name())
		_ = testFile.Close()
	}()

	// Write some dummy content to the test file
	_, err = testFile.WriteString(`{"test": "data"}`)
	require.NoError(t, err)

	// Setup test plugin
	testPlug := newTestPlugin(pluginapi.EditorTypeVSCode, testFile.Name(), unsortedSetting, nil)

	registry := plugins.NewRegistry()
	registry.Register(testPlug)

	// Create mapping config and logger
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create import service
	service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

	// Test import with sorting
	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: testFile,
	}

	res, err := service.Import(context.Background(), opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Setting)
	require.Len(t, res.Setting.Keybindings, 4)

	// Verify sorting by action
	expectedActions := []string{
		"actions.editor.copy",
		"actions.editor.cut",
		"actions.editor.paste",
		"actions.file.save",
	}

	actualActions := make([]string, len(res.Setting.Keybindings))
	for i, keymap := range res.Setting.Keybindings {
		actualActions[i] = keymap.Action
	}

	assert.Equal(t, expectedActions, actualActions, "Keymaps should be sorted by action")
}

func TestImportService_Import_EmptyKeymaps(t *testing.T) {
	// Test with empty keymaps
	emptySetting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{},
	}

	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_config_*.json")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(testFile.Name())
		_ = testFile.Close()
	}()

	// Setup test plugin
	testPlug := newTestPlugin(pluginapi.EditorTypeVSCode, testFile.Name(), emptySetting, nil)

	registry := plugins.NewRegistry()
	registry.Register(testPlug)

	// Create mapping config and logger
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create import service
	service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

	// Test import with empty keymaps
	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: testFile,
	}

	res, err := service.Import(context.Background(), opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Setting)
	assert.Len(t, res.Setting.Keybindings, 0)
}

func TestImportService_Import_NilSetting(t *testing.T) {
	// Create a temporary test file
	testFile, err := os.CreateTemp("", "test_config_*.json")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(testFile.Name())
		_ = testFile.Close()
	}()

	// Setup test plugin with nil setting
	testPlug := newTestPlugin(pluginapi.EditorTypeVSCode, testFile.Name(), nil, nil)

	registry := plugins.NewRegistry()
	registry.Register(testPlug)

	// Create mapping config and logger
	mappingConfig, err := mappings.NewTestMappingConfig()
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create import service
	service := NewImportService(registry, mappingConfig, logger, metrics.NewNoop())

	// Test import with nil setting
	opts := importapi.ImportOptions{
		EditorType:  pluginapi.EditorTypeVSCode,
		InputStream: testFile,
	}

	res, err := service.Import(context.Background(), opts)

	// Assertions
	require.NoError(t, err)
	assert.Nil(t, res)
}
