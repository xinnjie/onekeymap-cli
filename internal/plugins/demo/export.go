package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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

func (e *demoExporter) Export(
	_ context.Context,
	destination io.Writer,
	setting keymap.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	var out []demoBinding
	for _, km := range setting.Actions {
		for _, b := range km.Bindings {
			if len(b.KeyChords) == 0 {
				continue
			}
			keys := b.String(keybinding.FormatOption{Separator: "+"})
			out = append(out, demoBinding{Keys: keys, Action: km.Name})
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal demo bindings: %w", err)
	}
	if _, err := destination.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to destination: %w", err)
	}

	d, err := e.calculateDiff(opts.ExistingConfig, out)
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
