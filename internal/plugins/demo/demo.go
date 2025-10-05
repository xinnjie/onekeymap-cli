package demo

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

type demoPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi.Plugin {
	return &demoPlugin{
		mappingConfig: mappingConfig,
		importer:      newImporter(logger),
		exporter:      newExporter(logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

func (p *demoPlugin) EditorType() pluginapi.EditorType { return pluginapi.EditorType("demo") }

func (p *demoPlugin) Importer() (pluginapi.PluginImporter, error) { return p.importer, nil }

func (p *demoPlugin) Exporter() (pluginapi.PluginExporter, error) { return p.exporter, nil }
