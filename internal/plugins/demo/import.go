package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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
) (*keymapv1.Keymap, error) {
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	clean := internal.StripJSONComments(data)
	var bindings []struct {
		Keys   string `json:"keys"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(clean, &bindings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal demo keybindings: %w", err)
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
		setting.Keybindings = append(
			setting.Keybindings,
			&keymapv1.ActionBinding{Id: b.Action, Bindings: []*keymapv1.Binding{{KeyChords: kb.KeyChords}}},
		)
	}
	return setting, nil
}
