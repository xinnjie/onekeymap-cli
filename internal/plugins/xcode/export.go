package xcode

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"

	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"howett.net/plist"
)

const (
	xcodeVersion = 3
)

var (
	_ pluginapi.PluginExporter = (*xcodeExporter)(nil)
)

type xcodeExporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	differ        diff.Differ
}

// orderByBaseCommand reorders exported keybindings following the order of commands
// present in the base config. Items whose command is not present in base keep
// their original relative order and are placed after those that do.
func orderByBaseCommand(final []xcodeKeybinding, base xcodeKeybindingConfig) []xcodeKeybinding {
	if len(final) == 0 || len(base) == 0 {
		return final
	}
	// Build first-seen order for each action in base
	baseOrder := make(map[string]int, len(base))
	next := 0
	for _, kb := range base {
		if kb.Action == "" {
			continue
		}
		if _, ok := baseOrder[kb.Action]; !ok {
			baseOrder[kb.Action] = next
			next++
		}
	}
	if len(baseOrder) == 0 {
		return final
	}
	// Only reorder when both items have actions present in baseOrder.
	// Otherwise, keep their original relative order to avoid disrupting
	// non-destructive merges where managed items should precede unmanaged ones
	// unless the base explicitly defines ordering for both.
	sort.SliceStable(final, func(i, j int) bool {
		oi, okI := baseOrder[final[i].Action]
		oj, okJ := baseOrder[final[j].Action]
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
	return &xcodeExporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		differ:        differ,
	}
}

func (e *xcodeExporter) Export(
	_ context.Context,
	destination io.Writer,
	setting *keymapv1.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	// Decode existing config for non-destructive merge
	var existingKeybindings []xcodeKeybinding
	var existingTextKeybindings xcodeTextKeybinding
	if opts.ExistingConfig != nil {
		// Read all content first
		rawData, err := io.ReadAll(opts.ExistingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing config: %w", err)
		}

		plistData, err := parseXcodeConfig(rawData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode existing config: %w", err)
		}
		existingKeybindings = plistData.MenuKeyBindings.KeyBindings
		existingTextKeybindings = plistData.TextKeyBindings.KeyBindings
	}

	var unmanagedKeybindings []xcodeKeybinding
	if len(existingKeybindings) > 0 {
		unmanagedKeybindings = e.identifyUnmanagedKeybindings(existingKeybindings)
	}

	managedKeybindings := e.generateManagedKeybindings(setting)

	finalKeybindings := e.mergeKeybindings(managedKeybindings, unmanagedKeybindings)

	// Reorder according to base command order if provided
	finalKeybindings = orderByBaseCommand(finalKeybindings, existingKeybindings)

	// Generate text key bindings
	finalTextKeybindings := e.generateTextKeyBindings(setting, existingTextKeybindings)

	// Use plist library to generate the XML
	plistData := xcodeKeybindingsPlist{
		MenuKeyBindings: menuKeyBindings{
			KeyBindings: finalKeybindings,
			Version:     xcodeVersion,
		},
		TextKeyBindings: textKeyBindings{
			KeyBindings: finalTextKeybindings,
			Version:     xcodeVersion,
		},
	}

	encoder := plist.NewEncoder(destination)
	encoder.Indent("\t")
	if err := encoder.Encode(plistData); err != nil {
		return nil, fmt.Errorf("failed to write plist XML: %w", err)
	}

	// Defer diff calculation to exportService
	return &pluginapi.PluginExportReport{
		BaseEditorConfig:   existingKeybindings,
		ExportEditorConfig: finalKeybindings,
	}, nil
}

// identifyUnmanagedKeybindings performs reverse lookup to identify keybindings
// that are not managed by onekeymap.
func (e *xcodeExporter) identifyUnmanagedKeybindings(existingKeybindings []xcodeKeybinding) []xcodeKeybinding {
	unmanaged := make([]xcodeKeybinding, 0)

	for _, kb := range existingKeybindings {
		// Try to find this keybinding in action_mappings via reverse lookup
		mapping := e.findMappingByXcodeKeybinding(kb)
		if mapping == nil {
			// This keybinding is not managed by onekeymap
			unmanaged = append(unmanaged, kb)
		}
	}

	return unmanaged
}

// findMappingByXcodeKeybinding performs reverse lookup to find if an Xcode keybinding
// corresponds to any action in our mappings.
// It checks both Action (for Menu Key Bindings) and CommandID.
func (e *xcodeExporter) findMappingByXcodeKeybinding(kb xcodeKeybinding) *mappings.ActionMappingConfig {
	for _, mapping := range e.mappingConfig.Mappings {
		for _, xcodeConfig := range mapping.Xcode {
			if xcodeConfig.MenuAction.Action == kb.Action &&
				xcodeConfig.MenuAction.CommandID == kb.CommandID {
				return &mapping
			}
		}
	}
	return nil
}

// generateManagedKeybindings generates Xcode keybindings from KeymapSetting.
func (e *xcodeExporter) generateManagedKeybindings(setting *keymapv1.Keymap) []xcodeKeybinding {
	var xcodeKeybindings []xcodeKeybinding

	for _, km := range setting.GetActions() {
		mapping := e.mappingConfig.Get(km.GetName())
		if mapping == nil {
			continue
		}

		xcodeConfigs := mapping.Xcode
		if len(xcodeConfigs) == 0 {
			continue
		}

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			binding := keymap.NewKeyBinding(b)
			keys, err := FormatKeybinding(binding)
			if err != nil {
				e.logger.Warn("Skipping keybinding with un-formattable key", "action", km.GetName(), "error", err)
				continue
			}
			for _, xcodeConfig := range xcodeConfigs {
				if xcodeConfig.MenuAction.Action == "" {
					continue
				}
				xcodeKeybindings = append(xcodeKeybindings, xcodeKeybinding{
					Action:           xcodeConfig.MenuAction.Action,
					CommandID:        xcodeConfig.MenuAction.CommandID,
					CommandGroupID:   xcodeConfig.MenuAction.CommandGroupID,
					KeyboardShortcut: keys,
					Title:            xcodeConfig.MenuAction.Title,
					Alternate:        xcodeConfig.MenuAction.Alternate,
					Group:            xcodeConfig.MenuAction.Group,
					GroupID:          xcodeConfig.MenuAction.GroupID,
					GroupedAlternate: xcodeConfig.MenuAction.GroupedAlternate,
					Navigation:       xcodeConfig.MenuAction.Navigation,
				})
			}
		}
	}

	return xcodeKeybindings
}

// mergeKeybindings merges managed and unmanaged keybindings, with managed taking priority.
func (e *xcodeExporter) mergeKeybindings(managed, unmanaged []xcodeKeybinding) []xcodeKeybinding {
	// Create a map to track managed keybindings by their key combination
	managedKeys := make(map[string]bool)
	for _, kb := range managed {
		managedKeys[kb.KeyboardShortcut] = true
	}

	// Start with all managed keybindings; initialize with zero length to satisfy makezero lint
	result := make([]xcodeKeybinding, 0, len(managed)+len(unmanaged))
	result = append(result, managed...)

	// Add unmanaged keybindings that don't conflict with managed ones
	for _, kb := range unmanaged {
		if !managedKeys[kb.KeyboardShortcut] {
			result = append(result, kb)
		} else {
			// Log conflict - managed keybinding takes priority
			e.logger.Debug("Conflict resolved: managed keybinding takes priority",
				"key", kb.KeyboardShortcut, "unmanaged_action", kb.Action)
		}
	}

	return result
}

// generateTextKeyBindings generates Xcode Text Key Bindings from KeymapSetting.
// It merges managed text bindings with existing ones, with managed taking priority.
func (e *xcodeExporter) generateTextKeyBindings(
	setting *keymapv1.Keymap,
	existingTextBindings xcodeTextKeybinding,
) xcodeTextKeybinding {
	// Start with existing text bindings or create new map
	result := make(xcodeTextKeybinding)
	for k, v := range existingTextBindings {
		result[k] = v
	}

	// Generate managed text bindings
	for _, km := range setting.GetActions() {
		mapping := e.mappingConfig.Get(km.GetName())
		if mapping == nil {
			continue
		}

		xcodeConfigs := mapping.Xcode
		if len(xcodeConfigs) == 0 {
			continue
		}

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			binding := keymap.NewKeyBinding(b)
			keys, err := FormatKeybinding(binding)
			if err != nil {
				e.logger.Warn("Skipping text keybinding with un-formattable key", "action", km.GetName(), "error", err)
				continue
			}

			for _, xcodeConfig := range xcodeConfigs {
				items := xcodeConfig.TextAction.TextAction.Items
				if len(items) == 0 {
					continue
				}
				result[keys] = textActionValue{Items: items}
			}
		}
	}

	return result
}
