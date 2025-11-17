package vscode

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/dedup"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// vscodeImporter handles importing keybindings from VSCode.
type vscodeImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	reporter      *metrics.UnknownActionReporter
}

func newImporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *vscodeImporter {
	return &vscodeImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		reporter:      metrics.NewUnknownActionReporter(recorder),
	}
}

// Import reads a VSCode keybindings.json file and converts it to a universal KeymapSetting.
func (i *vscodeImporter) Import(
	ctx context.Context,
	source io.Reader,
	_ pluginapi2.PluginImportOption,
) (pluginapi2.PluginImportResult, error) {
	vscodeKeybindings, err := parseExistingConfig(source)
	if err != nil {
		return pluginapi2.PluginImportResult{}, fmt.Errorf("failed to parse existing config: %w", err)
	}

	setting := &keymapv1.Keymap{}
	marker := imports.NewMarker()
	for _, binding := range vscodeKeybindings {
		mapping := i.FindByVSCodeActionWithArgs(binding.Command, binding.When, binding.Args)
		if mapping == nil {
			i.logger.DebugContext(
				ctx,
				"Skipping keybinding with unknown action",
				"action",
				binding.Command,
				"when",
				binding.When,
				"args",
				binding.Args,
			)
			// Report unknown command metric
			if i.reporter != nil {
				i.reporter.ReportUnknownCommand(ctx, pluginapi2.EditorTypeVSCode, binding.Command)
			}
			marker.MarkSkippedForReason(binding.Command, pluginapi2.ErrActionNotSupported)
			continue
		}

		kb, err := ParseKeybinding(binding.Key)
		if err != nil {
			i.logger.WarnContext(ctx, "Skipping keybinding with unparsable key", "key", binding.Key, "error", err)
			marker.MarkSkippedForReason(binding.Command, fmt.Errorf("unparsable key '%s': %w", binding.Key, err))
			continue
		}

		newKeymap := &keymapv1.Action{
			Name:     mapping.ID,
			Bindings: []*keymapv1.KeybindingReadable{{KeyChords: kb.KeyChords}},
		}
		setting.Actions = append(setting.Actions, newKeymap)

		marker.MarkImported(binding.Command)
	}

	setting.Actions = dedup.DedupKeyBindings(setting.GetActions())

	result := pluginapi2.PluginImportResult{Keymap: setting}
	result.Report.SkipReport = marker.Report()
	return result, nil
}
