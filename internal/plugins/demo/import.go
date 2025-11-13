package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/tailscale/hujson"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
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

	setting := &keymapv1.Keymap{}
	for _, b := range bindings {
		if b.Keys == "" || b.Action == "" {
			continue
		}
		kb, err := keymap.ParseKeyBinding(b.Keys, "+")
		if err != nil {
			i.logger.WarnContext(ctx, "Skipping unparsable keybinding", "keys", b.Keys, "error", err)
			continue
		}
		setting.Actions = append(
			setting.Actions,
			&keymapv1.Action{Name: b.Action, Bindings: []*keymapv1.KeybindingReadable{{KeyChords: kb.KeyChords}}},
		)
	}
	return pluginapi.PluginImportResult{Keymap: setting}, nil
}
