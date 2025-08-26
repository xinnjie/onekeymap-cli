package zed

import (
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var (
	_ pluginapi.Plugin = (*zedPlugin)(nil)
)

// zedPlugin implements the plugins.Plugin interface for the Zed editor.
type zedPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new ZedPlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi.Plugin {
	return &zedPlugin{
		mappingConfig: mappingConfig,
		importer:      newImporter(mappingConfig, logger),
		exporter:      newExporter(mappingConfig, logger, diff.NewJsonAsciiDiffer()),
		logger:        logger,
	}
}

// EditorType returns the name of the plugin.
func (p *zedPlugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeZed
}

func (p *zedPlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.importer, nil
}

func (p *zedPlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.exporter, nil
}
