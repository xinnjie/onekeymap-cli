package vscode

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

var (
	_ pluginapi.PluginExporter = (*vscodeExporter)(nil)
)

type vscodeExporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	differ        diff.Differ
}

// orderByBaseCommand reorders exported keybindings following the order of commands
// present in the base config. Items whose command is not present in base keep
// their original relative order and are placed after those that do.
func orderByBaseCommand(final []vscodeKeybinding, base vscodeKeybindingConfig) []vscodeKeybinding {
	if len(final) == 0 || len(base) == 0 {
		return final
	}
	// Build first-seen order for each command in base
	baseOrder := make(map[string]int, len(base))
	next := 0
	for _, kb := range base {
		if kb.Command == "" {
			continue
		}
		if _, ok := baseOrder[kb.Command]; !ok {
			baseOrder[kb.Command] = next
			next++
		}
	}
	if len(baseOrder) == 0 {
		return final
	}
	// Only reorder when both items have commands present in baseOrder.
	// Otherwise, keep their original relative order to avoid disrupting
	// non-destructive merges where managed items should precede unmanaged ones
	// unless the base explicitly defines ordering for both.
	sort.SliceStable(final, func(i, j int) bool {
		oi, okI := baseOrder[final[i].Command]
		oj, okJ := baseOrder[final[j].Command]
		if okI && okJ {
			return oi < oj
		}
		// Keep original order when one or both are not in baseOrder
		return false
	})
	return final
}

func newExporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	differ diff.Differ,
) pluginapi.PluginExporter {
	return &vscodeExporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		differ:        differ,
	}
}

func (e *vscodeExporter) Export(
	ctx context.Context,
	destination io.Writer,
	setting *keymapv1.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	// Decode existing config for non-destructive merge
	var existingKeybindings []vscodeKeybinding
	if opts.ExistingConfig != nil {
		if err := json.NewDecoder(opts.ExistingConfig).Decode(&existingKeybindings); err != nil {
			return nil, fmt.Errorf("failed to decode existing config: %w", err)
		}
	}

	var unmanagedKeybindings []vscodeKeybinding
	if len(existingKeybindings) > 0 {
		unmanagedKeybindings = e.identifyUnmanagedKeybindings(existingKeybindings)
	}

	managedKeybindings := e.generateManagedKeybindings(setting)

	finalKeybindings := e.mergeKeybindings(managedKeybindings, unmanagedKeybindings)

	// Reorder according to base command order if provided
	finalKeybindings = orderByBaseCommand(finalKeybindings, existingKeybindings)

	// Write JSON without HTML escaping so that '&&', '<', '>' remain as-is
	enc := json.NewEncoder(destination)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(finalKeybindings); err != nil {
		return nil, fmt.Errorf("failed to encode vscode keybindings to json: %w", err)
	}

	// Defer diff calculation to exportService
	return &pluginapi.PluginExportReport{
		BaseEditorConfig:   existingKeybindings,
		ExportEditorConfig: finalKeybindings,
	}, nil
}

// identifyUnmanagedKeybindings performs reverse lookup to identify keybindings
// that are not managed by onekeymap.
func (e *vscodeExporter) identifyUnmanagedKeybindings(existingKeybindings []vscodeKeybinding) []vscodeKeybinding {
	unmanaged := make([]vscodeKeybinding, 0)

	for _, kb := range existingKeybindings {
		// Try to find this keybinding in action_mappings via reverse lookup
		mapping := e.findMappingByVSCodeKeybinding(kb)
		if mapping == nil {
			// This keybinding is not managed by onekeymap
			unmanaged = append(unmanaged, kb)
		}
	}

	return unmanaged
}

// findMappingByVSCodeKeybinding performs reverse lookup to find if a VSCode keybinding
// corresponds to any action in our mappings.
func (e *vscodeExporter) findMappingByVSCodeKeybinding(kb vscodeKeybinding) *mappings.ActionMappingConfig {
	for _, mapping := range e.mappingConfig.Mappings {
		for _, vscodeConfig := range mapping.VSCode {
			if vscodeConfig.Command == kb.Command &&
				vscodeConfig.When == kb.When &&
				equalVSCodeArgs(vscodeConfig.Args, kb.Args) {
				return &mapping
			}
		}
	}
	return nil
}

// generateManagedKeybindings generates VSCode keybindings from KeymapSetting.
func (e *vscodeExporter) generateManagedKeybindings(setting *keymapv1.Keymap) []vscodeKeybinding {
	var vscodeKeybindings []vscodeKeybinding

	for _, km := range setting.GetKeybindings() {
		mapping := e.mappingConfig.FindByUniversalAction(km.GetName())
		if mapping == nil {
			continue
		}

		vscodeConfigs := mapping.VSCode
		if len(vscodeConfigs) == 0 {
			continue
		}

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			binding := keymap.NewKeyBinding(b)
			keys, err := formatKeybinding(binding)
			if err != nil {
				e.logger.Warn("Skipping keybinding with un-formattable key", "action", km.GetName(), "error", err)
				continue
			}
			for _, vscodeConfig := range vscodeConfigs {
				if vscodeConfig.Command == "" {
					continue
				}
				vscodeKeybindings = append(vscodeKeybindings, vscodeKeybinding{
					Key:     keys,
					Command: vscodeConfig.Command,
					When:    vscodeConfig.When,
					Args:    vscodeConfig.Args,
				})
			}
		}
	}

	return vscodeKeybindings
}

// mergeKeybindings merges managed and unmanaged keybindings, with managed taking priority.
func (e *vscodeExporter) mergeKeybindings(managed, unmanaged []vscodeKeybinding) []vscodeKeybinding {
	// Create a map to track managed keybindings by their key combination
	managedKeys := make(map[string]bool)
	for _, kb := range managed {
		managedKeys[kb.Key] = true
	}

	// Start with all managed keybindings; initialize with zero length to satisfy makezero lint
	result := make([]vscodeKeybinding, 0, len(managed)+len(unmanaged))
	result = append(result, managed...)

	// Add unmanaged keybindings that don't conflict with managed ones
	for _, kb := range unmanaged {
		if !managedKeys[kb.Key] {
			result = append(result, kb)
		} else {
			// Log conflict - managed keybinding takes priority
			e.logger.Debug("Conflict resolved: managed keybinding takes priority",
				"key", kb.Key, "unmanaged_command", kb.Command)
		}
	}

	return result
}

// equalVSCodeArgs compares VSCode args, handling the conversion between different types.
func equalVSCodeArgs(a map[string]interface{}, b vscodeArgs) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	// Convert both to JSON strings for comparison to avoid type issues
	aJSON, err1 := json.Marshal(a)
	bJSON, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}
