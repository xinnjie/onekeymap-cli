package vscode

import (
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

// vsCodePlugin implements the plugins.Plugin interface for the VSCode editor.
type vsCodePlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new VSCodePlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi.Plugin {
	return &vsCodePlugin{
		mappingConfig: mappingConfig,
		importer:      newImporter(mappingConfig, logger),
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
