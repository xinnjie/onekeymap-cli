package intellij

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/dedup"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

type intellijImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	reporter      *metrics.UnknownActionReporter
}

func newImporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *intellijImporter {
	return &intellijImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		reporter:      metrics.NewUnknownActionReporter(recorder),
	}
}

// Import converts IntelliJ keymap XML into the universal KeymapSetting.
func (p *intellijImporter) Import(
	ctx context.Context,
	source io.Reader,
	opts pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	_ = ctx
	_ = opts

	raw, err := io.ReadAll(source)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	var doc KeymapXML
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to parse intellij keymap xml: %w", err)
	}

	setting := keymap.Keymap{}
	marker := imports.NewMarker()
	for _, act := range doc.Actions {
		actionID, err := p.ActionIDFromIntelliJ(act.ID)
		if err != nil {
			// Not found in mapping, skip quietly
			p.logger.DebugContext(ctx, "no universal mapping for intellij action", "action", act.ID)
			p.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeIntelliJ, act.ID)
			marker.MarkSkippedForReason(act.ID, err)
			continue
		}

		for _, ks := range act.KeyboardShortcuts {
			if ks.First == "" {
				continue
			}
			kb, err := ParseKeyBinding(ks)
			if err != nil {
				p.logger.WarnContext(ctx, "failed to parse key binding", "binding", ks, "error", err)
				marker.MarkSkippedForReason(act.ID, fmt.Errorf("failed to parse key binding %v: %w", ks, err))
				continue
			}
			newBinding := keymap.Action{
				Name: actionID,
				Bindings: []keybinding.Keybinding{
					kb,
				},
			}
			setting.Actions = append(setting.Actions, newBinding)
			marker.MarkImported(act.ID)
		}
	}

	setting.Actions = dedup.Actions(setting.Actions)
	result := pluginapi.PluginImportResult{Keymap: setting}
	result.Report.SkipReport = marker.Report()
	return result, nil
}

// ActionIDFromIntelliJ converts an IntelliJ action and optional context to a universal action ID.
func (p *intellijImporter) ActionIDFromIntelliJ(action string) (string, error) {
	for _, mapping := range p.mappingConfig.Mappings {
		if mapping.IntelliJ.Action == action {
			return mapping.ID, nil
		}
	}
	return "", fmt.Errorf("no mapping found for intellij action: %s", action)
}
