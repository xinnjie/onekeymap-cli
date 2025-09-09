package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/diff"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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
func (p *zedExporter) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts pluginapi.PluginExportOption) (*pluginapi.PluginExportReport, error) {
	// Parse existing configuration if provided
	var existingConfig zedKeymapConfig
	if opts.ExistingConfig != nil {
		if err := json.NewDecoder(opts.ExistingConfig).Decode(&existingConfig); err != nil {
			return nil, fmt.Errorf("failed to decode existing config: %w", err)
		}
	}

	// Generate managed keybindings from current setting
	managedKeymaps := p.generateManagedKeybindings(setting)

	// Merge managed and existing keybindings
	finalKeymaps := p.mergeKeybindings(managedKeymaps, existingConfig)

	// Ensure deterministic order by sorting contexts
	sort.Slice(finalKeymaps, func(i, j int) bool {
		return finalKeymaps[i].Context < finalKeymaps[j].Context
	})

	// Write JSON without HTML escaping so that '&&', '<', '>' remain as-is
	enc := json.NewEncoder(destination)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(finalKeymaps); err != nil {
		return nil, fmt.Errorf("failed to encode vscode keybindings to json: %w", err)
	}

	// Prepare structured 'before' from opts.Base for centralized diffing
	var baseConfig zedKeymapConfig
	if opts.Base != nil {
		if err := json.NewDecoder(opts.Base).Decode(&baseConfig); err != nil {
			return nil, fmt.Errorf("failed to decode base: %w", err)
		}
	}

	// Defer diff calculation to exportService. Provide structured before/after configs.
	return &pluginapi.PluginExportReport{
		BaseEditorConfig:   baseConfig,
		ExportEditorConfig: finalKeymaps,
	}, nil
}

// generateManagedKeybindings creates keybindings from the current setting
func (p *zedExporter) generateManagedKeybindings(setting *keymapv1.KeymapSetting) zedKeymapConfig {
	keymapsByContext := make(map[string]map[string]zedActionValue)

	for _, km := range setting.Keybindings {
		actionConfig, err := p.actionIDToZed(km.Id)
		if err != nil {
			p.logger.Info("no mapping found for action", "action", km.Id)
			continue
		}

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			keys, err := formatZedKeybind(keymap.NewKeyBinding(b))
			if err != nil {
				p.logger.Warn("failed to format key binding", "error", err)
				continue
			}

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

// mergeKeybindings merges managed and existing keybindings, with managed taking priority
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
				if existingAction, exists := mergedBindings[key]; exists {
					p.logger.Debug("Conflict resolved: managed keybinding takes priority",
						"context", context, "key", key, "existing_action", existingAction.Action, "managed_action", action.Action)
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
