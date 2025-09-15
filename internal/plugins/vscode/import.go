package vscode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// vscodeImporter handles importing keybindings from VSCode.
type vscodeImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
}

func newImporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger) *vscodeImporter {
	return &vscodeImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
	}
}

// Import reads a VSCode keybindings.json file and converts it to a universal KeymapSetting.
func (i *vscodeImporter) Import(
	ctx context.Context,
	source io.Reader,
	option pluginapi.PluginImportOption,
) (*keymapv1.KeymapSetting, error) {
	// VSCode's keybindings.json can contain comments, so we need to strip them.
	jsonData, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	cleanedJSON := internal.StripJSONComments(jsonData)

	var vscodeKeybindings []vscodeKeybinding
	if err := json.Unmarshal(cleanedJSON, &vscodeKeybindings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vscode keybindings: %w", err)
	}

	setting := &keymapv1.KeymapSetting{}
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
			continue
		}

		kb, err := parseKeybinding(binding.Key)
		if err != nil {
			i.logger.WarnContext(ctx, "Skipping keybinding with unparsable key", "key", binding.Key, "error", err)
			continue
		}

		newKeymap := &keymapv1.ActionBinding{
			Id:       mapping.ID,
			Bindings: []*keymapv1.Binding{{KeyChords: kb.KeyChords}},
		}
		setting.Keybindings = append(setting.Keybindings, newKeymap)
	}

	return setting, nil
}
