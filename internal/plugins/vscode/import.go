package vscode

import (
	"context"
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

// vscodeImporter handles importing keybindings from VSCode.
type vscodeImporter struct {
	mappingConfig *mappings.MappingConfig
	editorType    pluginapi.EditorType
	logger        *slog.Logger
	reporter      *metrics.UnknownActionReporter
}

func newImporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *vscodeImporter {
	return newImporterWithEditorType(pluginapi.EditorTypeVSCode, mappingConfig, logger, recorder)
}

func newImporterWithEditorType(
	editorType pluginapi.EditorType,
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) *vscodeImporter {
	return &vscodeImporter{
		mappingConfig: mappingConfig,
		editorType:    editorType,
		logger:        logger,
		reporter:      metrics.NewUnknownActionReporter(recorder),
	}
}

// Import reads a VSCode keybindings.json file and converts it to a universal KeymapSetting.
func (i *vscodeImporter) Import(
	ctx context.Context,
	source io.Reader,
	_ pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	vscodeKeybindings, err := parseConfig(source)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to parse existing config: %w", err)
	}

	setting := keymap.Keymap{}
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
				i.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeVSCode, binding.Command)
			}
			marker.MarkSkippedForReason(binding.Command, pluginapi.ErrActionNotSupported)
			continue
		}

		kb, err := ParseKeybinding(binding.Key)
		if err != nil {
			i.logger.WarnContext(ctx, "Skipping keybinding with unparsable key", "key", binding.Key, "error", err)
			marker.MarkSkippedForReason(binding.Command, fmt.Errorf("unparsable key '%s': %w", binding.Key, err))
			continue
		}

		newKeymap := keymap.Action{
			Name: mapping.ID,
			Bindings: []keybinding.Keybinding{
				*kb,
			},
		}
		setting.Actions = append(setting.Actions, newKeymap)

		marker.MarkImported(binding.Command)
	}

	setting.Actions = dedup.Actions(setting.Actions)

	result := pluginapi.PluginImportResult{Keymap: setting}
	result.Report.SkipReport = marker.Report()
	return result, nil
}
