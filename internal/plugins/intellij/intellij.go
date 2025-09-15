package intellij

import (
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

type intellijPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new IntelliJ plugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi.Plugin {
	return &intellijPlugin{
		mappingConfig: mappingConfig,
		importer:      newImporter(mappingConfig, logger),
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

func (p *intellijPlugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeIntelliJ
}

func (p *intellijPlugin) Importer() (pluginapi.PluginImporter, error) { return p.importer, nil }
func (p *intellijPlugin) Exporter() (pluginapi.PluginExporter, error) { return p.exporter, nil }
