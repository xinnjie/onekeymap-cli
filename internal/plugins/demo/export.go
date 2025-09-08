package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type demoExporter struct {
	logger *slog.Logger
	differ diff.Differ
}

type demoBinding struct {
	Keys   string `json:"keys"`
	Action string `json:"action"`
}

func newExporter(logger *slog.Logger, differ diff.Differ) pluginapi.PluginExporter {
	return &demoExporter{logger: logger, differ: differ}
}

func (e *demoExporter) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts pluginapi.PluginExportOption) (*pluginapi.PluginExportReport, error) {
	var out []demoBinding
	for _, km := range setting.GetKeybindings() {
		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			binding := keymap.NewKeyBinding(b)
			keys, err := binding.Format(platform.PlatformMacOS, "+")
			if err != nil {
				e.logger.Warn("Skipping un-formattable keybinding", "action", km.GetId(), "error", err)
				continue
			}
			out = append(out, demoBinding{Keys: keys, Action: km.GetId()})
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal demo bindings: %w", err)
	}
	if _, err := destination.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to destination: %w", err)
	}

	d, err := e.calculateDiff(opts.Base, out)
	if err != nil {
		return nil, err
	}
	return &pluginapi.PluginExportReport{Diff: &d}, nil
}

func (e *demoExporter) calculateDiff(base io.Reader, out []demoBinding) (string, error) {
	var before any
	if base == nil {
		before = []any{}
	} else {
		if err := json.NewDecoder(base).Decode(&before); err != nil {
			return "", fmt.Errorf("failed to decode base: %w", err)
		}
	}

	raw, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("failed to marshal out: %w", err)
	}
	var after any
	if err := json.Unmarshal(raw, &after); err != nil {
		return "", fmt.Errorf("failed to unmarshal out: %w", err)
	}

	return e.differ.Diff(before, after)
}
