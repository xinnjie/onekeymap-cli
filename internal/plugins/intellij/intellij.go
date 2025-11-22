package intellij

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

type intellijPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new IntelliJ plugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi.Plugin {
	return newIntellijPlugin(mappingConfig, logger, recorder)
}

func newIntellijPlugin(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *intellijPlugin {
	importer := newImporter(mappingConfig, logger, recorder)

	return &intellijPlugin{
		mappingConfig: mappingConfig,
		importer:      importer,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

func (p *intellijPlugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeIntelliJ
}

func (p *intellijPlugin) Importer() (pluginapi.PluginImporter, error) { return p.importer, nil }
func (p *intellijPlugin) Exporter() (pluginapi.PluginExporter, error) { return p.exporter, nil }
