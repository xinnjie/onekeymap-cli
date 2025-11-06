package intellij

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
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
) (*keymapv1.Keymap, error) {
	_ = ctx
	_ = opts

	raw, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	var doc KeymapXML
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse intellij keymap xml: %w", err)
	}

	setting := &keymapv1.Keymap{}

	for _, act := range doc.Actions {
		actionID, err := p.ActionIDFromIntelliJ(act.ID)
		if err != nil {
			// Not found in mapping, skip quietly
			p.logger.DebugContext(ctx, "no universal mapping for intellij action", "action", act.ID)
			p.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeIntelliJ, act.ID)
			continue
		}

		for _, ks := range act.KeyboardShortcuts {
			if ks.First == "" {
				continue
			}
			kb, err := ParseKeyBinding(ks)
			if err != nil {
				p.logger.WarnContext(ctx, "failed to parse key binding", "binding", ks, "error", err)
				continue
			}
			newBinding := &keymapv1.Action{
				Name:     actionID,
				Bindings: []*keymapv1.KeybindingReadable{{KeyChords: kb.KeyChords}},
			}
			setting.Actions = append(setting.Actions, newBinding)
		}
	}

	setting.Actions = internal.DedupKeyBindings(setting.GetActions())
	return setting, nil
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
