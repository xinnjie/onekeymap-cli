package zed

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

var (
	_ pluginapi2.Plugin = (*zedPlugin)(nil)
)

// zedPlugin implements the plugins.Plugin interface for the Zed editor.
type zedPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi2.PluginImporter
	exporter      pluginapi2.PluginExporter
	logger        *slog.Logger
}

// New creates a new ZedPlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi2.Plugin {
	importer := newImporter(mappingConfig, logger, recorder)

	return &zedPlugin{
		mappingConfig: mappingConfig,
		importer:      importer,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

// EditorType returns the name of the plugin.
func (p *zedPlugin) EditorType() pluginapi2.EditorType {
	return pluginapi2.EditorTypeZed
}

func (p *zedPlugin) Importer() (pluginapi2.PluginImporter, error) {
	return p.importer, nil
}

func (p *zedPlugin) Exporter() (pluginapi2.PluginExporter, error) {
	return p.exporter, nil
}
