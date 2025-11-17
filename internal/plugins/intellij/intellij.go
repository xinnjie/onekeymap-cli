package intellij

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type intellijPlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi2.PluginImporter
	exporter      pluginapi2.PluginExporter
	logger        *slog.Logger
}

// New creates a new IntelliJ plugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi2.Plugin {
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

func (p *intellijPlugin) EditorType() pluginapi2.EditorType {
	return pluginapi2.EditorTypeIntelliJ
}

func (p *intellijPlugin) Importer() (pluginapi2.PluginImporter, error) { return p.importer, nil }
func (p *intellijPlugin) Exporter() (pluginapi2.PluginExporter, error) { return p.exporter, nil }
