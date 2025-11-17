package vscode

import (
	"fmt"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type vsCodeVariantPlugin struct {
	*vsCodePlugin

	editorType pluginapi2.EditorType
}

func newVSCodeVariantPlugin(
	editorType pluginapi2.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return &vsCodeVariantPlugin{
		vsCodePlugin: newVSCodePlugin(mappingConfig, logger, recorder),
		editorType:   editorType,
	}
}

// EditorType implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) EditorType() pluginapi2.EditorType {
	return p.editorType
}

// Importer implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Importer() (pluginapi2.PluginImporter, error) {
	return p.vsCodePlugin.Importer()
}

// Exporter implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) Exporter() (pluginapi2.PluginExporter, error) {
	return p.vsCodePlugin.Exporter()
}

// ConfigDetect implements pluginapi.Plugin.
func (p *vsCodeVariantPlugin) ConfigDetect(
	opts pluginapi2.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	switch p.editorType {
	case pluginapi2.EditorTypeWindsurf:
		return detectConfigForVSCodeVariant("Windsurf", "windsurf", opts)
	case pluginapi2.EditorTypeWindsurfNext:
		return detectConfigForVSCodeVariant("Windsurf-Next", "windsurf-next", opts)
	case pluginapi2.EditorTypeCursor:
		return detectConfigForVSCodeVariant("Cursor", "cursor", opts)
	case pluginapi2.EditorTypeVSCode:
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
) pluginapi2.Plugin {
	return newVSCodeVariantPlugin(pluginapi2.EditorTypeWindsurf, mappingConfig, logger, recorder)
}

// NewWindsurfNext creates a Windsurf-Next plugin instance.
func NewWindsurfNext(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newVSCodeVariantPlugin(pluginapi2.EditorTypeWindsurfNext, mappingConfig, logger, recorder)
}

// NewCursor creates a Cursor plugin instance.
func NewCursor(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) pluginapi2.Plugin {
	return newVSCodeVariantPlugin(pluginapi2.EditorTypeCursor, mappingConfig, logger, recorder)
}
