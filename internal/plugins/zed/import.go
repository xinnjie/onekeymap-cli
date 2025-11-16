package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/tailscale/hujson"
	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type zedImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	reporter      *metrics.UnknownActionReporter
}

func newImporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) *zedImporter {
	return &zedImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		reporter:      metrics.NewUnknownActionReporter(recorder),
	}
}

// Import reads the editor's configuration source and converts it into the
// universal onekeymap KeymapSetting format.
func (p *zedImporter) Import(
	ctx context.Context,
	source io.Reader,
	_ pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	jsonData, err := io.ReadAll(source)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	cleanedJSON, err := hujson.Standardize(jsonData)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to standardize JSON: %w", err)
	}

	var zedKeymaps zedKeymapConfig
	if err := json.Unmarshal(cleanedJSON, &zedKeymaps); err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to parse zed keymap json: %w", err)
	}

	setting := &keymapv1.Keymap{}
	marker := imports.NewMarker()
	for _, zk := range zedKeymaps {
		// ensure deterministic order: iterate bindings by sorted keys
		keys := make([]string, 0, len(zk.Bindings))
		for k := range zk.Bindings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			action := zk.Bindings[key]
			kb, err := ParseZedKeybind(key)
			if err != nil {
				p.logger.WarnContext(ctx, "failed to parse keychord", "key", key, "error", err)
				marker.MarkSkippedForReason(action.Action, fmt.Errorf("failed to parse keychord '%s': %w", key, err))
				continue
			}

			// Use strong-typed action value parsed via UnmarshalJSON
			actionStr := action.Action
			actionArgs := action.Args
			if actionStr == "" {
				// Unsupported or empty action; skip
				p.logger.WarnContext(ctx, "unsupported or empty action", "key", key)
				continue
			}

			actionID, err := p.actionIDFromZedWithArgs(actionStr, zk.Context, actionArgs)
			if err != nil {
				// If a mapping is not found, we simply skip it for now.
				// In the future, this could be logged or added to a report.
				p.logger.WarnContext(
					ctx,
					"failed to find action",
					"action",
					actionStr,
					"context",
					zk.Context,
					"args",
					actionArgs,
					"error",
					err,
				)
				p.reporter.ReportUnknownCommand(ctx, pluginapi.EditorTypeZed, actionStr)
				marker.MarkSkippedForReason(actionStr, err)
				continue
			}
			keymapEntry := &keymapv1.Action{
				Name: actionID,
				Bindings: []*keymapv1.KeybindingReadable{
					{KeyChords: kb.KeyChords},
				},
			}

			setting.Actions = append(setting.Actions, keymapEntry)
			marker.MarkImported(actionStr)
		}
	}
	setting.Actions = internal.DedupKeyBindings(setting.GetActions())
	result := pluginapi.PluginImportResult{Keymap: setting}
	result.Report.SkipReport = marker.Report()
	return result, nil
}
