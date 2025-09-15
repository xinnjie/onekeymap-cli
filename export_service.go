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
func NewExportService(
	registry *plugins.Registry,
	config *mappings.MappingConfig,
	logger *slog.Logger,
) exportapi.Exporter {
	return &exportService{
		registry:      registry,
		mappingConfig: config,
		logger:        logger,
	}
}

// Export is the method implementation for the default service.
func (s *exportService) Export(
	ctx context.Context,
	destination io.Writer,
	setting *keymapv1.KeymapSetting,
	opts exportapi.ExportOptions,
) (*exportapi.ExportReport, error) {
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

	// If unified diff is requested and a base is provided, buffer the base once and
	// use duplicated readers so that the plugin can consume its copy without
	// affecting our diff generation.
	var pluginBase = opts.Base
	var diffBase = opts.Base
	var baseBuf bytes.Buffer
	if opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF && opts.Base != nil {
		// Stream Base to plugin while teeing into baseBuf for later diff computation.
		pluginBase = io.TeeReader(opts.Base, &baseBuf)
		// Defer setting diffBase until after plugin export completes so baseBuf is fully populated.
		writer = io.MultiWriter(destination, &newConfigBuf)
	}

	report, err := exporter.Export(ctx, writer, setting, pluginapi.PluginExportOption{ExistingConfig: pluginBase})
	if err != nil {
		return nil, fmt.Errorf("failed to export config: %w", err)
	}

	// If we used TeeReader above, ensure we drain any remaining bytes from opts.Base
	// (in case the plugin didn't read to EOF), then use the full buffered Base for diff now.
	if opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF && opts.Base != nil {
		drainBase := func() error {
			if _, err := io.Copy(&baseBuf, opts.Base); err != nil {
				return fmt.Errorf("failed to drain base for unified diff: %w", err)
			}
			return nil
		}
		if err := drainBase(); err != nil {
			return nil, err
		}
		diffBase = bytes.NewReader(baseBuf.Bytes())
	}
	diffStr, err := s.computeDiff(opts, diffBase, &newConfigBuf, report)
	if err != nil {
		return nil, err
	}
	return &exportapi.ExportReport{Diff: diffStr}, nil
}

// computeDiff centralizes diff generation for export results based on requested options.
func (s *exportService) computeDiff(
	opts exportapi.ExportOptions,
	originalConfig io.Reader,
	updateConfig *bytes.Buffer,
	report *pluginapi.PluginExportReport,
) (string, error) {
	switch {
	case opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF && originalConfig != nil:
		// Unified diff over raw editor configs
		ud := diff.NewUnifiedDiffFormatDiffer()
		d, err := ud.Diff(originalConfig, updateConfig, opts.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to compute unified diff: %w", err)
		}
		return d, nil
	case report != nil && report.BaseEditorConfig != nil && report.ExportEditorConfig != nil:
		// JSON ASCII diff over structured editor configs supplied by plugin
		jd := diff.NewJSONASCIIDiffer()
		d, err := jd.Diff(report.BaseEditorConfig, report.ExportEditorConfig)
		if err != nil {
			return "", fmt.Errorf("failed to compute ascii json diff: %w", err)
		}
		return d, nil
	case report != nil && report.Diff != nil:
		// Backward compatibility with legacy plugin-provided diffs
		return *report.Diff, nil
	default:
		return "", nil
	}
}
