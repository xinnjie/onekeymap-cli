package helix

import (
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

var _ pluginapi2.Plugin = (*helixPlugin)(nil)

type helixPlugin struct {
	mappingConfig *mappings.MappingConfig
	exporter      pluginapi2.PluginExporter
	logger        *slog.Logger
}

// New creates a new Helix plugin instance.
func New(mappingConfig *mappings.MappingConfig, logger *slog.Logger) pluginapi2.Plugin {
	return &helixPlugin{
		mappingConfig: mappingConfig,
		exporter:      newExporter(mappingConfig, logger, diff.NewJSONASCIIDiffer()),
		logger:        logger,
	}
}

// EditorType returns the unique identifier for Helix.
func (p *helixPlugin) EditorType() pluginapi2.EditorType { return pluginapi2.EditorTypeHelix }

// Importer returns the importer for this plugin.
func (p *helixPlugin) Importer() (pluginapi2.PluginImporter, error) {
	return nil, pluginapi2.ErrNotSupported
}

// Exporter returns the exporter for this plugin.
func (p *helixPlugin) Exporter() (pluginapi2.PluginExporter, error) { return p.exporter, nil }
