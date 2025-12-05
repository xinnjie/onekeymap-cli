package exporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
)

// exporter is the default implementation of the Exporter interface.
type exporter struct {
	registry        *registry.Registry
	mappingConfig   *mappings.MappingConfig
	logger          *slog.Logger
	recorder        metrics.Recorder
	serviceReporter *metrics.ServiceReporter
}

// NewExporter creates a new default export service.
func NewExporter(
	registry *registry.Registry,
	config *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) exporterapi.Exporter {
	return &exporter{
		registry:        registry,
		mappingConfig:   config,
		logger:          logger,
		recorder:        recorder,
		serviceReporter: metrics.NewServiceReporter(recorder),
	}
}

// Export is the method implementation for the default service.
func (s *exporter) Export(
	ctx context.Context,
	destination io.Writer,
	setting keymap.Keymap,
	opts exporterapi.ExportOptions,
) (*exporterapi.ExportReport, error) {
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
	if opts.OriginalConfig != nil {
		baseReadForPlugin = io.TeeReader(opts.OriginalConfig, &baseBuf)
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
	if opts.OriginalConfig != nil {
		baseReadForDiff = bytes.NewReader(baseBuf.Bytes())
	}
	diffStr, err := s.computeDiff(opts, baseReadForDiff, &newConfigBuf, report)
	if err != nil {
		return nil, err
	}

	// Build export coverage from plugin report
	coverage := s.buildCoverage(setting, report)

	return &exporterapi.ExportReport{
		Diff:        diffStr,
		Coverage:    coverage,
		SkipActions: report.SkipReport.SkipActions,
	}, nil
}

// computeDiff centralizes diff generation for export results based on requested options.
func (s *exporter) computeDiff(
	opts exporterapi.ExportOptions,
	originalConfig io.Reader,
	updatedConfig io.Reader,
	report *pluginapi.PluginExportReport,
) (string, error) {
	switch {
	case opts.DiffType == exporterapi.DiffTypeUnified:
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
	case opts.DiffType == exporterapi.DiffTypeASCII:
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

// buildCoverage computes export coverage by comparing requested actions with plugin export results.
func (s *exporter) buildCoverage(
	input keymap.Keymap,
	report *pluginapi.PluginExportReport,
) exporterapi.ExportCoverage {
	coverage := exporterapi.ExportCoverage{
		TotalActions: len(input.Actions),
	}

	if report == nil {
		return coverage
	}

	// Build a map of action results from plugin report
	resultMap := make(map[string]pluginapi.ActionExportResult)
	for _, r := range report.ExportedReport.Actions {
		resultMap[r.Action] = r
	}

	// Track which actions were skipped
	skippedActions := make(map[string]bool)
	for _, skip := range report.SkipReport.SkipActions {
		skippedActions[skip.Action] = true
	}

	for _, action := range input.Actions {
		// Skip actions that were completely skipped
		if skippedActions[action.Name] {
			continue
		}

		result, hasResult := resultMap[action.Name]
		if !hasResult {
			// Action was exported but plugin didn't report details
			// Assume fully exported
			coverage.FullyExported++
			continue
		}

		// Compare requested vs exported keybindings
		if len(result.Exported) >= len(result.Requested) {
			coverage.FullyExported++
		} else {
			coverage.PartiallyExported = append(coverage.PartiallyExported, exporterapi.PartialExportedAction{
				Action:    action.Name,
				Requested: result.Requested,
				Exported:  result.Exported,
				Reason:    result.Reason,
			})
		}
	}

	return coverage
}
