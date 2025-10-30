package internal_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// testExportPlugin is a minimal plugin implementation for export tests.
type testExportPlugin struct {
	editorType pluginapi.EditorType
	exporter   pluginapi.PluginExporter
}

func (p *testExportPlugin) EditorType() pluginapi.EditorType { return p.editorType }
func (p *testExportPlugin) ConfigDetect(_ pluginapi.ConfigDetectOptions) ([]string, bool, error) {
	return nil, false, pluginapi.ErrNotSupported
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

func (e *testExporter) Export(
	_ context.Context,
	destination io.Writer,
	_ *keymapv1.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	if e.writeContent != "" {
		_, _ = io.Copy(destination, strings.NewReader(e.writeContent))
	}
	// Simulate plugins that consume the base stream if provided
	if opts.ExistingConfig != nil {
		_, _ = io.Copy(io.Discard, opts.ExistingConfig)
	}
	return &pluginapi.PluginExportReport{
		Diff:               e.reportDiff,
		BaseEditorConfig:   e.baseEditorConfig,
		ExportEditorConfig: e.exportEditorConfig,
	}, nil
}

func newTestExportService(
	t *testing.T,
	exporter pluginapi.PluginExporter,
	editorType pluginapi.EditorType,
) exportapi.Exporter {
	t.Helper()
	r := plugins.NewRegistry()
	r.Register(&testExportPlugin{editorType: editorType, exporter: exporter})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	recorder := metrics.NewNoop()
	return internal.NewExportService(
		r,
		&mappings.MappingConfig{Mappings: map[string]mappings.ActionMappingConfig{}},
		logger,
		recorder,
	)
}

func TestExportService_Diff_Unified(t *testing.T) {
	before := "line1\n"
	after := "line1\nline2\n"
	exp := &testExporter{writeContent: after}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.Keymap{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		Base:       strings.NewReader(before),
		DiffType:   keymapv1.ExportKeymapRequest_UNIFIED_DIFF,
		FilePath:   "test.txt",
	})
	require.NoError(t, err)
	assert.Equal(t, after, out.String())
	require.NotNil(t, report)
	want := "diff --git a/test.txt b/test.txt\n--- a/test.txt\n+++ b/test.txt\n@@ -1 +1,2 @@\n line1\n+line2\n"
	assert.Equal(t, want, report.Diff)
}

func TestExportService_Diff_JSONASCII_FromStructuredConfigs(t *testing.T) {
	baseCfg := map[string]any{"k": "v1"}
	afterCfg := map[string]any{"k": "v2"}
	exp := &testExporter{baseEditorConfig: baseCfg, exportEditorConfig: afterCfg}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.Keymap{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		DiffType:   keymapv1.ExportKeymapRequest_ASCII_DIFF,
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	want := " {\n\x1b[30;41m-  \"k\": \"v1\"\x1b[0m\n\x1b[30;42m+  \"k\": \"v2\"\x1b[0m\n }\n"
	assert.Equal(t, want, report.Diff)
}

func TestExportService_Diff_FallbackFromPlugin(t *testing.T) {
	fallback := "fallback-diff"
	exp := &testExporter{reportDiff: &fallback}
	service := newTestExportService(t, exp, pluginapi.EditorType("test"))

	var out bytes.Buffer
	report, err := service.Export(context.Background(), &out, &keymapv1.Keymap{}, exportapi.ExportOptions{
		EditorType: pluginapi.EditorType("test"),
		DiffType:   keymapv1.ExportKeymapRequest_DIFF_TYPE_UNSPECIFIED,
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, fallback, report.Diff)
}
