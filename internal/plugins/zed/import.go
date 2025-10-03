package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type zedImporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
}

func newImporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger) *zedImporter {
	return &zedImporter{
		mappingConfig: mappingConfig,
		logger:        logger,
	}
}

// Import reads the editor's configuration source and converts it into the
// universal onekeymap KeymapSetting format.
func (p *zedImporter) Import(
	ctx context.Context,
	source io.Reader,
	opts pluginapi.PluginImportOption,
) (*keymapv1.Keymap, error) {
	jsonData, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	cleanedJSON := internal.StripJSONComments(jsonData)

	var zedKeymaps zedKeymapConfig
	if err := json.Unmarshal(cleanedJSON, &zedKeymaps); err != nil {
		return nil, fmt.Errorf("failed to parse zed keymap json: %w", err)
	}

	setting := &keymapv1.Keymap{}
	for _, zk := range zedKeymaps {
		// ensure deterministic order: iterate bindings by sorted keys
		keys := make([]string, 0, len(zk.Bindings))
		for k := range zk.Bindings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			action := zk.Bindings[key]
			kb, err := parseZedKeybind(key)
			if err != nil {
				p.logger.WarnContext(ctx, "failed to parse keychord", "key", key, "error", err)
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
				continue
			}
			keymapEntry := &keymapv1.Action{
				Name: actionID,
				Bindings: []*keymapv1.Binding{
					{KeyChords: kb.KeyChords},
				},
			}

			setting.Keybindings = append(setting.Keybindings, keymapEntry)
		}
	}
	return setting, nil
}
