package onekeymap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// testExportPlugin is a minimal plugin implementation for export tests.
type testExportPlugin struct {
	editorType pluginapi.EditorType
	exporter   pluginapi.PluginExporter
}

func (p *testExportPlugin) EditorType() pluginapi.EditorType { return p.editorType }
func (p *testExportPlugin) DefaultConfigPath(opts ...pluginapi.DefaultConfigPathOption) ([]string, error) {
	return nil, pluginapi.ErrNotSupported
}
func (p *testExportPlugin) Importer() (pluginapi.PluginImporter, error) {
	return nil, pluginapi.ErrNotSupported
}
func (p *testExportPlugin) Exporter() (pluginapi.PluginExporter, error) { return p.exporter, nil }

// testExporter returns a configurable PluginExportReport and optionally writes to destination.
type testExporter struct {
	writeContent       string
	baseEditorConfig   any
	exportEditorConfig any
	reportDiff         *string
}

func (e *testExporter) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts pluginapi.PluginExportOption) (*pluginapi.PluginExportReport, error) {
	if e.writeContent != "" {
		_, _ = io.Copy(destination, strings.NewReader(e.writeContent))
	}
	return &pluginapi.PluginExportReport{
		Diff:               e.reportDiff,
		BaseEditorConfig:   e.baseEditorConfig,
		ExportEditorConfig: e.exportEditorConfig,
	}, nil
}

func newTestExportService(t *testing.T, exporter pluginapi.PluginExporter, editorType pluginapi.EditorType) exportapi.Exporter {
	t.Helper()
	r := plugins.NewRegistry()
	r.Register(&testExportPlugin{editorType: editorType, exporter: exporter})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewExportService(r, &mappings.MappingConfig{Mappings: map[string]mappings.ActionMappingConfig{}}, logger)
}

func TestExportService_Diff_Unified(t *testing.T) {
	before := "line1\n"
	after := "line1\nline2\n"
	exp := &testExporter{writeContent: after}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.KeymapSetting{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		Base:       strings.NewReader(before),
		DiffType:   keymapv1.ExportKeymapRequest_UNIFIED_DIFF,
	})
	require.NoError(t, err)
	assert.Equal(t, after, out.String())
	require.NotNil(t, report)
	assert.NotEmpty(t, report.Diff)
	// Unified patch text should include additions marker or hunk header
	assert.True(t, strings.Contains(report.Diff, "@@") || strings.Contains(report.Diff, "\n+"))
}

func TestExportService_Diff_JSONASCII_FromStructuredConfigs(t *testing.T) {
	baseCfg := map[string]any{"k": "v1"}
	afterCfg := map[string]any{"k": "v2"}
	exp := &testExporter{baseEditorConfig: baseCfg, exportEditorConfig: afterCfg}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.KeymapSetting{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		DiffType:   keymapv1.ExportKeymapRequest_ASCII_DIFF,
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.NotEmpty(t, report.Diff)
	assert.Contains(t, report.Diff, "k")
	// JSON ascii differ denotes removals and additions
	assert.True(t, strings.Contains(report.Diff, "+") || strings.Contains(report.Diff, "-"))
}

func TestExportService_Diff_FallbackFromPlugin(t *testing.T) {
	fallback := "fallback-diff"
	exp := &testExporter{reportDiff: &fallback}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.KeymapSetting{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		DiffType:   keymapv1.ExportKeymapRequest_ASCII_DIFF,
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, fallback, report.Diff)
}
