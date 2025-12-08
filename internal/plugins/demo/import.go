package demo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/tailscale/hujson"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type demoImporter struct {
	logger *slog.Logger
}

func newImporter(logger *slog.Logger) pluginapi.PluginImporter {
	return &demoImporter{logger: logger}
}

// Import reads a minimal demo keybindings JSON array and converts it to a universal KeymapSetting.
// Expected input shape (comments allowed):
// [
//
//	{ "keys": "ctrl+c", "action": "actions.edit.copy" },
//	{ "keys": "ctrl+v", "action": "actions.edit.paste" }
//
// ].
func (i *demoImporter) Import(
	ctx context.Context,
	source io.Reader,
	_ pluginapi.PluginImportOption,
) (pluginapi.PluginImportResult, error) {
	data, err := io.ReadAll(source)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to read from reader: %w", err)
	}

	clean, err := hujson.Standardize(data)
	if err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to standardize JSON: %w", err)
	}
	var bindings []struct {
		Keys   string `json:"keys"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(clean, &bindings); err != nil {
		return pluginapi.PluginImportResult{}, fmt.Errorf("failed to unmarshal demo keybindings: %w", err)
	}

	setting := keymap.Keymap{}
	marker := imports.NewMarker()
	for _, b := range bindings {
		if b.Keys == "" || b.Action == "" {
			marker.MarkSkipped(b.Action, nil, errors.New("missing keys or action"))
			continue
		}
		kb, err := keybinding.NewKeybinding(b.Keys, keybinding.ParseOption{Separator: "+"})
		if err != nil {
			i.logger.WarnContext(ctx, "Skipping unparsable keybinding", "keys", b.Keys, "error", err)
			marker.MarkSkipped(b.Action, nil, fmt.Errorf("unparsable key '%s': %w", b.Keys, err))
			continue
		}
		setting.Actions = append(
			setting.Actions,
			keymap.Action{Name: b.Action, Bindings: []keybinding.Keybinding{kb}},
		)
		marker.MarkImported(b.Action, b.Action, kb, kb)
	}
	result := pluginapi.PluginImportResult{Keymap: setting}
	result.Report.SkipReport = marker.Report()
	return result, nil
}
