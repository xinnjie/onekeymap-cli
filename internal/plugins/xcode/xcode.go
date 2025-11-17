package xcode

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

// xcodePlugin implements the plugins.Plugin interface for the Xcode editor.
type xcodePlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi2.PluginImporter
	exporter      pluginapi2.PluginExporter
	logger        *slog.Logger
}

// New creates a new XcodePlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi2.Plugin {
	return newXcodePlugin(mappingConfig, logger, recorder)
}

func newXcodePlugin(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *xcodePlugin {
	importer := newImporter(mappingConfig, logger, recorder)

	return &xcodePlugin{
		mappingConfig: mappingConfig,
		importer:      importer,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

// EditorType returns the name of the plugin.
func (p *xcodePlugin) EditorType() pluginapi2.EditorType {
	return pluginapi2.EditorTypeXcode
}

// ConfigDetect returns the default path to the editor's configuration file based on the platform.
func (p *xcodePlugin) ConfigDetect(opts pluginapi2.ConfigDetectOptions) (paths []string, installed bool, err error) {
	return detectXcodeConfig(opts)
}

// Importer returns the importer for this plugin.
func (p *xcodePlugin) Importer() (pluginapi2.PluginImporter, error) {
	return p.importer, nil
}

// Exporter returns the exporter for this plugin.
func (p *xcodePlugin) Exporter() (pluginapi2.PluginExporter, error) {
	return p.exporter, nil
}
