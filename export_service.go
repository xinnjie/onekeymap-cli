package onekeymap

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// exportService is the default implementation of the Exporter interface.
type exportService struct {
	registry      *plugins.Registry
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
}

// NewExportService creates a new default export service.
func NewExportService(registry *plugins.Registry, config *mappings.MappingConfig, logger *slog.Logger) exportapi.Exporter {
	return &exportService{
		registry:      registry,
		mappingConfig: config,
		logger:        logger,
	}
}

// Export is the method implementation for the default service.
func (s *exportService) Export(destination io.Writer, setting *keymapv1.KeymapSetting, opts exportapi.ExportOptions) (*exportapi.ExportReport, error) {
	plugin, ok := s.registry.Get(opts.EditorType)
	if !ok {
		return nil, fmt.Errorf("no plugin found for editor type '%s'", opts.EditorType)
	}

	exporter, err := plugin.Exporter()
	if err != nil {
		return nil, fmt.Errorf("failed to get exporter for %s: %w", opts.EditorType, err)
	}

	report, err := exporter.Export(context.TODO(), destination, setting, pluginapi.PluginExportOption{Base: opts.Base})
	if err != nil {
		return nil, fmt.Errorf("failed to export config: %w", err)
	}

	diff := func() string {
		if report.Diff == nil {
			return ""
		}
		return *report.Diff
	}()

	return &exportapi.ExportReport{Diff: diff}, nil
}
