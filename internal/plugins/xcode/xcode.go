package xcode

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

// xcodePlugin implements the plugins.Plugin interface for the Xcode editor.
type xcodePlugin struct {
	mappingConfig *mappings.MappingConfig
	importer      pluginapi.PluginImporter
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new XcodePlugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) pluginapi.Plugin {
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
func (p *xcodePlugin) EditorType() pluginapi.EditorType {
	return pluginapi.EditorTypeXcode
}

// ConfigDetect returns the default path to the editor's configuration file based on the platform.
func (p *xcodePlugin) ConfigDetect(opts pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	return detectXcodeConfig(opts)
}

// Importer returns the importer for this plugin.
func (p *xcodePlugin) Importer() (pluginapi.PluginImporter, error) {
	return p.importer, nil
}

// Exporter returns the exporter for this plugin.
func (p *xcodePlugin) Exporter() (pluginapi.PluginExporter, error) {
	return p.exporter, nil
}
