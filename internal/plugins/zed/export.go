package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/tailscale/hujson"
	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/export"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type zedExporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	differ        diff.Differ
}

func newExporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger, differ diff.Differ) *zedExporter {
	return &zedExporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		differ:        differ,
	}
}

// Export takes a universal KeymapSetting and writes it to an editor-specific
// configuration destination.
func (p *zedExporter) Export(
	_ context.Context,
	destination io.Writer,
	setting *keymapv1.Keymap,
	opts pluginapi2.PluginExportOption,
) (*pluginapi2.PluginExportReport, error) {
	// Parse existing configuration if provided
	existingConfig, err := p.parseExistingConfig(opts.ExistingConfig)
	if err != nil {
		return nil, err
	}

	// Generate managed keybindings from current setting
	marker := export.NewMarker(setting)
	managedKeymaps := p.generateManagedKeybindings(setting, marker)

	// Merge managed and existing keybindings
	finalKeymaps := p.mergeKeybindings(managedKeymaps, existingConfig)

	// Order contexts by base if provided; otherwise fallback to alphabetical for determinism
	if len(existingConfig) > 0 {
		finalKeymaps = orderByBaseContext(finalKeymaps, existingConfig)
	} else {
		sort.Slice(finalKeymaps, func(i, j int) bool {
			return finalKeymaps[i].Context < finalKeymaps[j].Context
		})
	}

	// Write JSON without HTML escaping so that '&&', '<', '>' remain as-is
	enc := json.NewEncoder(destination)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(finalKeymaps); err != nil {
		return nil, fmt.Errorf("failed to encode vscode keybindings to json: %w", err)
	}

	// Defer diff calculation to exportService. Provide structured before/after configs.
	return &pluginapi2.PluginExportReport{
		BaseEditorConfig:   existingConfig,
		ExportEditorConfig: finalKeymaps,
		SkipReport:         marker.Report(),
	}, nil
}

// parseExistingConfig reads and parses the existing configuration from the reader.
func (p *zedExporter) parseExistingConfig(reader io.Reader) (zedKeymapConfig, error) {
	var config zedKeymapConfig
	if reader == nil {
		return config, nil
	}

	rawData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing config: %w", err)
	}

	if len(rawData) == 0 {
		return config, nil
	}

	standardizedData, err := hujson.Standardize(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to standardize JSON: %w", err)
	}

	if err := json.Unmarshal(standardizedData, &config); err != nil {
		return nil, fmt.Errorf("failed to decode existing config: %w", err)
	}

	return config, nil
}

// orderByBaseContext reorders exported contexts following the order present
// in the base config. Contexts not present in base keep their relative order
// after those that do, with an alphabetical fallback for determinism.
func orderByBaseContext(final zedKeymapConfig, base zedKeymapConfig) zedKeymapConfig {
	if len(final) == 0 || len(base) == 0 {
		return final
	}
	baseOrder := make(map[string]int, len(base))
	next := 0
	for _, ctx := range base {
		if ctx.Context == "" {
			continue
		}
		if _, ok := baseOrder[ctx.Context]; !ok {
			baseOrder[ctx.Context] = next
			next++
		}
	}
	if len(baseOrder) == 0 {
		return final
	}
	sort.SliceStable(final, func(i, j int) bool {
		oi, okI := baseOrder[final[i].Context]
		oj, okJ := baseOrder[final[j].Context]
		if okI && okJ {
			return oi < oj
		}
		if okI && !okJ {
			return true
		}
		if !okI && okJ {
			return false
		}
		// Neither in base: fallback to alphabetical for determinism
		return final[i].Context < final[j].Context
	})
	return final
}

// generateManagedKeybindings creates keybindings from the current setting.
func (p *zedExporter) generateManagedKeybindings(setting *keymapv1.Keymap, marker *export.Marker) zedKeymapConfig {
	keymapsByContext := make(map[string]map[string]zedActionValue)

	for _, km := range setting.GetActions() {
		actionConfig, err := p.actionIDToZed(km.GetName())
		if err != nil {
			p.logger.Info("no mapping found for action", "action", km.GetName())
			for _, b := range km.GetBindings() {
				if b != nil && b.GetKeyChords() != nil {
					marker.MarkSkippedForReason(km.GetName(), b.GetKeyChords(), pluginapi2.ErrActionNotSupported)
				}
			}
			continue
		}

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			keys, err := FormatZedKeybind(keymap.NewKeyBinding(b))
			if err != nil {
				p.logger.Warn("failed to format key binding", "error", err)
				marker.MarkSkippedForReason(
					km.GetName(),
					b.GetKeyChords(),
					&pluginapi2.UnsupportedExportActionError{Note: err.Error()},
				)
				continue
			}
			marker.MarkExported(km.GetName(), b.GetKeyChords())

			// For each Zed mapping config, create a binding under its context
			for _, zconf := range *actionConfig {
				if zconf.Action == "" {
					continue
				}
				if _, ok := keymapsByContext[zconf.Context]; !ok {
					keymapsByContext[zconf.Context] = make(map[string]zedActionValue)
				}

				// Create action value - either string or array with args
				var actionValue zedActionValue
				if len(zconf.Args) > 0 {
					// Use array format: [action, args]
					actionValue = zedActionValue{Action: zconf.Action, Args: zconf.Args}
				} else {
					// Use simple string format
					actionValue = zedActionValue{Action: zconf.Action}
				}

				keymapsByContext[zconf.Context][keys] = actionValue
			}
		}
	}

	result := make(zedKeymapConfig, 0, len(keymapsByContext))
	for context, bindings := range keymapsByContext {
		result = append(result, zedKeymapOfContext{
			Context:  context,
			Bindings: bindings,
		})
	}

	return result
}

// mergeKeybindings merges managed and existing keybindings, with managed taking priority.
func (p *zedExporter) mergeKeybindings(managed, existing zedKeymapConfig) zedKeymapConfig {
	// Create a map for quick lookup of existing contexts
	existingByContext := make(map[string]map[string]zedActionValue)
	for _, contextConfig := range existing {
		existingByContext[contextConfig.Context] = contextConfig.Bindings
	}

	// Create a map for managed contexts
	managedByContext := make(map[string]map[string]zedActionValue)
	for _, contextConfig := range managed {
		managedByContext[contextConfig.Context] = contextConfig.Bindings
	}

	// Merge all contexts
	allContexts := make(map[string]bool)
	for context := range existingByContext {
		allContexts[context] = true
	}
	for context := range managedByContext {
		allContexts[context] = true
	}

	result := make(zedKeymapConfig, 0, len(allContexts))
	for context := range allContexts {
		mergedBindings := make(map[string]zedActionValue)

		// Start with existing bindings
		if existingBindings, exists := existingByContext[context]; exists {
			for key, action := range existingBindings {
				mergedBindings[key] = action
			}
		}

		// Override with managed bindings (managed takes priority)
		if managedBindings, exists := managedByContext[context]; exists {
			for key, action := range managedBindings {
				if existingAction, actionExists := mergedBindings[key]; actionExists {
					p.logger.Debug(
						"Conflict resolved: managed keybinding takes priority",
						"context",
						context,
						"key",
						key,
						"existing_action",
						existingAction.Action,
						"managed_action",
						action.Action,
					)
				}
				mergedBindings[key] = action
			}
		}

		// Only include contexts that have bindings
		if len(mergedBindings) > 0 {
			result = append(result, zedKeymapOfContext{
				Context:  context,
				Bindings: mergedBindings,
			})
		}
	}

	return result
}

// actionIDToZed converts a universal action ID to Zed action and context.
func (p *zedExporter) actionIDToZed(actionID string) (*mappings.ZedConfigs, error) {
	if mapping, exists := p.mappingConfig.Mappings[actionID]; exists {
		return &mapping.Zed, nil
	}
	return nil, fmt.Errorf("no mapping found for action ID '%s'", actionID)
}
