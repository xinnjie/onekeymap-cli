package vscode

import (
	"fmt"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

type vsCodeVariantPlugin struct {
	editorType    pluginapi.EditorType
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

func newVSCodeVariantPlugin(
	editorType pluginapi.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return &vsCodeVariantPlugin{
		editorType:    editorType,
		mappingConfig: mappingConfig,
		importer:      newImporterWithEditorType(editorType, mappingConfig, logger, recorder),
		exporter:      newExporterWithEditorType(editorType, mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

// EditorType implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) EditorType() pluginapi.EditorType {
	return p.editorType
}

// Importer implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.importer, nil
}

// Exporter implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.exporter, nil
}

// ConfigDetect implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) ConfigDetect(
	opts pluginapi.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	switch p.editorType {
	case pluginapi.EditorTypeWindsurf:
		return detectConfigForVSCodeVariant("Windsurf", "windsurf", opts)
	case pluginapi.EditorTypeWindsurfNext:
		return detectConfigForVSCodeVariant("Windsurf-Next", "windsurf-next", opts)
	case pluginapi.EditorTypeCursor:
		return detectConfigForVSCodeVariant("Cursor", "cursor", opts)
	case pluginapi.EditorTypeVSCode:
		return detectConfigForVSCodeVariant("Code", "code", opts)
	default:
		return nil, false, fmt.Errorf("unknown editor type: %s", p.editorType)
	}
}

// NewWindsurf creates a Windsurf plugin instance.
func NewWindsurf(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newVSCodeVariantPlugin(pluginapi.EditorTypeWindsurf, mappingConfig, logger, recorder)
}

// NewWindsurfNext creates a Windsurf-Next plugin instance.
func NewWindsurfNext(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newVSCodeVariantPlugin(pluginapi.EditorTypeWindsurfNext, mappingConfig, logger, recorder)
}

// NewCursor creates a Cursor plugin instance.
func NewCursor(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return newVSCodeVariantPlugin(pluginapi.EditorTypeCursor, mappingConfig, logger, recorder)
}
