package vscode

import (
	"fmt"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

type vsCodeVariantPlugin struct {
	*vsCodePlugin

	editorType pluginapi.EditorType
}

func newVSCodeVariantPlugin(
	editorType pluginapi.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi.Plugin {
	return &vsCodeVariantPlugin{
		vsCodePlugin: newVSCodePlugin(mappingConfig, logger, recorder),
		editorType:   editorType,
	}
}

// EditorType implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) EditorType() pluginapi.EditorType {
	return p.editorType
}

// Importer implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.vsCodePlugin.Importer()
}

// Exporter implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.vsCodePlugin.Exporter()
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
