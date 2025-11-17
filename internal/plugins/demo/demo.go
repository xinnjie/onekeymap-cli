package demo

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type demoPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi2.PluginImporter
	exporter      pluginapi2.PluginExporter
	logger        *slog.Logger
}

func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi2.Plugin {
	return &demoPlugin{
		mappingConfig: mappingConfig,
		importer:      newImporter(logger),
		exporter:      newExporter(logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

func (p *demoPlugin) EditorType() pluginapi2.EditorType { return pluginapi2.EditorType("demo") }

func (p *demoPlugin) Importer() (pluginapi2.PluginImporter, error) { return p.importer, nil }

func (p *demoPlugin) Exporter() (pluginapi2.PluginExporter, error) { return p.exporter, nil }
