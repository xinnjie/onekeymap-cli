package vscode

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// vsCodePlugin implements the plugins.Plugin interface for the VSCode editor.
type vsCodePlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new VSCodePlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi.Plugin {
	return newVSCodePlugin(mappingConfig, logger, recorder)
}

func newVSCodePlugin(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *vsCodePlugin {
	importer := newImporter(mappingConfig, logger, recorder)

	return &vsCodePlugin{
		mappingConfig: mappingConfig,
		importer:      importer,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

// EditorType returns the name of the plugin.
func (p *vsCodePlugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeVSCode
}

// Importer returns the importer for this plugin.
func (p *vsCodePlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.importer, nil
}

// Exporter returns the exporter for this plugin.
func (p *vsCodePlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.exporter, nil
}
