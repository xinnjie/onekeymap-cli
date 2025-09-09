package onekeymap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
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
func (s *exportService) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts exportapi.ExportOptions) (*exportapi.ExportReport, error) {
	plugin, ok := s.registry.Get(opts.EditorType)
	if !ok {
		return nil, fmt.Errorf("no plugin found for editor type '%s'", opts.EditorType)
	}

	exporter, err := plugin.Exporter()
	if err != nil {
		return nil, fmt.Errorf("failed to get exporter for %s: %w", opts.EditorType, err)
	}

	var newConfigBuf bytes.Buffer
	writer := destination
	if opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF && opts.Base != nil {
		writer = io.MultiWriter(destination, &newConfigBuf)
	}

	report, err := exporter.Export(ctx, writer, setting, pluginapi.PluginExportOption{Base: opts.Base})
	if err != nil {
		return nil, fmt.Errorf("failed to export config: %w", err)
	}

	var diffStr string
	if opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF && opts.Base != nil {
		gitDiffer := diff.NewUnifiedDiffFormatDiffer()
		diffString, err := gitDiffer.Diff(opts.Base, &newConfigBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to compute git diff: %w", err)
		}
		diffStr = diffString
	} else if report != nil && report.Diff != nil {
		diffStr = *report.Diff
	}

	return &exportapi.ExportReport{Diff: diffStr}, nil
}
