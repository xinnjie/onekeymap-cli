package helix

import (
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var _ pluginapi.Plugin = (*helixPlugin)(nil)

type helixPlugin struct {
	mappingConfig *mappings.MappingConfig
	exporter      pluginapi.PluginExporter
	logger        *slog.Logger
}

// New creates a new Helix plugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi.Plugin {
	return &helixPlugin{
		mappingConfig: mappingConfig,
		exporter:      newExporter(mappingConfig, logger, diff.NewJsonAsciiDiffer()),
		logger:        logger,
	}
}

// EditorType returns the unique identifier for Helix.
func (p *helixPlugin) EditorType() pluginapi.EditorType { return pluginapi.EditorTypeHelix }

// Importer returns the importer for this plugin.
func (p *helixPlugin) Importer() (pluginapi.PluginImporter, error) {
	return nil, pluginapi.ErrNotSupported
}

// Exporter returns the exporter for this plugin.
func (p *helixPlugin) Exporter() (pluginapi.PluginExporter, error) { return p.exporter, nil }
