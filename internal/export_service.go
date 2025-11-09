package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// exportService is the default implementation of the Exporter interface.
type exportService struct {
	registry        *plugins.Registry
	mappingConfig   *mappings.MappingConfig
	logger          *slog.Logger
	recorder        metrics.Recorder
	serviceReporter *metrics.ServiceReporter
}

// NewExportService creates a new default export service.
func NewExportService(
	registry *plugins.Registry,
	config *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) exportapi.Exporter {
	return &exportService{
		registry:        registry,
		mappingConfig:   config,
		logger:          logger,
		recorder:        recorder,
		serviceReporter: metrics.NewServiceReporter(recorder),
	}
}

// Export is the method implementation for the default service.
func (s *exportService) Export(
	ctx context.Context,
	destination io.Writer,
	setting *keymapv1.Keymap,
	opts exportapi.ExportOptions,
) (*exportapi.ExportReport, error) {
	s.serviceReporter.ReportExportCall(ctx)

	plugin, ok := s.registry.Get(opts.EditorType)
	if !ok {
		return nil, fmt.Errorf("no plugin found for editor type '%s'", opts.EditorType)
	}

	exporter, err := plugin.Exporter()
	if err != nil {
		return nil, fmt.Errorf("failed to get exporter for %s: %w", opts.EditorType, err)
	}

	var newConfigBuf bytes.Buffer
	writer := io.MultiWriter(destination, &newConfigBuf)

	// If unified diff is requested and a base is provided, buffer the base once and
	// use duplicated readers so that the plugin can consume its copy without
	// affecting our diff generation.
	var baseReadForPlugin io.Reader
	var baseBuf bytes.Buffer
	if opts.Base != nil {
		baseReadForPlugin = io.TeeReader(opts.Base, &baseBuf)
	}

	report, err := exporter.Export(
		ctx,
		writer,
		setting,
		pluginapi.PluginExportOption{ExistingConfig: baseReadForPlugin},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to export config: %w", err)
	}

	// If we used TeeReader above, ensure we drain any remaining bytes from opts.Base
	// (in case the plugin didn't read to EOF), then use the full buffered Base for diff now.
	drainBase := func() error {
		if baseReadForPlugin == nil {
			return nil
		}
		if _, err := io.Copy(io.Discard, baseReadForPlugin); err != nil {
			return fmt.Errorf("failed to drain base for unified diff: %w", err)
		}
		return nil
	}
	if err := drainBase(); err != nil {
		return nil, err
	}
	// Now that we've fully buffered the base, create a fresh reader for diffing
	var baseReadForDiff io.Reader
	if opts.Base != nil {
		baseReadForDiff = bytes.NewReader(baseBuf.Bytes())
	}
	diffStr, err := s.computeDiff(opts, baseReadForDiff, &newConfigBuf, report)
	if err != nil {
		return nil, err
	}
	return &exportapi.ExportReport{Diff: diffStr, SkipActions: report.SkipReport.SkipActions}, nil
}

// computeDiff centralizes diff generation for export results based on requested options.
func (s *exportService) computeDiff(
	opts exportapi.ExportOptions,
	originalConfig io.Reader,
	updatedConfig io.Reader,
	report *pluginapi.PluginExportReport,
) (string, error) {
	switch {
	case opts.DiffType == keymapv1.ExportKeymapRequest_UNIFIED_DIFF:
		// Unified diff over raw editor configs
		ud := diff.NewUnifiedDiffFormatDiffer()
		if originalConfig == nil {
			originalConfig = bytes.NewReader([]byte(""))
		}
		if updatedConfig == nil {
			updatedConfig = bytes.NewReader([]byte(""))
		}
		d, err := ud.Diff(originalConfig, updatedConfig, opts.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to compute unified diff: %w", err)
		}
		return d, nil
	case opts.DiffType == keymapv1.ExportKeymapRequest_ASCII_DIFF:
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
