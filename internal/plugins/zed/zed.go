package zed

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
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
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi.Plugin {
	importer := newImporter(mappingConfig, logger, recorder)

	return &zedPlugin{
		mappingConfig: mappingConfig,
		importer:      importer,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
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
