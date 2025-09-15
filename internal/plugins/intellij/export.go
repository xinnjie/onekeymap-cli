package intellij

import (
	"context"
	"encoding/xml"
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

type intellijExporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	differ        diff.Differ
}

func newExporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger, differ diff.Differ) *intellijExporter {
	return &intellijExporter{mappingConfig: mappingConfig, logger: logger, differ: differ}
}

// Export writes IntelliJ-specific keymap XML to destination.
func (e *intellijExporter) Export(
	ctx context.Context,
	destination io.Writer,
	setting *keymapv1.KeymapSetting,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	_ = ctx

	// Read existing configuration if provided for non-destructive export
	var existingDoc KeymapXML
	if opts.ExistingConfig != nil {
		if err := xml.NewDecoder(opts.ExistingConfig).Decode(&existingDoc); err != nil {
			e.logger.Warn("Failed to parse existing config, proceeding with destructive export", "error", err)
		}
	}

	// Identify unmanaged actions from existing config
	var unmanagedActions []ActionXML
	if opts.ExistingConfig != nil {
		unmanagedActions = e.identifyUnmanagedActions(existingDoc.Actions)
	}

	// Generate managed actions from current setting
	managedActions := e.generateManagedActions(setting)

	// Merge managed and unmanaged actions
	finalActions := e.mergeActions(managedActions, unmanagedActions)

	doc := KeymapXML{
		Name:             "Onekeymap",
		Version:          "1",
		DisableMnemonics: true,
		Actions:          finalActions,
		Parent:           "$default",
	}

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal intellij keymap xml: %w", err)
	}
	if _, err := destination.Write([]byte(xml.Header)); err != nil {
		return nil, fmt.Errorf("failed to write xml header: %w", err)
	}
	if _, err := destination.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write xml body: %w", err)
	}

	return &pluginapi.PluginExportReport{
		BaseEditorConfig:   existingDoc,
		ExportEditorConfig: doc,
	}, nil
}

// identifyUnmanagedActions performs reverse lookup to identify actions
// that are not managed by onekeymap.
func (e *intellijExporter) identifyUnmanagedActions(existingActions []ActionXML) []ActionXML {
	var unmanaged []ActionXML

	for _, action := range existingActions {
		if !e.isManagedAction(action.ID) {
			unmanaged = append(unmanaged, action)
		}
	}

	return unmanaged
}

// isManagedAction checks if an action is managed by onekeymap.
func (e *intellijExporter) isManagedAction(actionID string) bool {
	for _, mapping := range e.mappingConfig.Mappings {
		if mapping.IntelliJ.Action == actionID {
			return true
		}
	}
	return false
}

// generateManagedActions generates IntelliJ actions from KeymapSetting.
func (e *intellijExporter) generateManagedActions(setting *keymapv1.KeymapSetting) []ActionXML {
	// Group keybindings by IntelliJ action ID while preserving order of first appearance.
	actionsMap := make(map[string]*ActionXML)
	var actionOrder []string

	for _, km := range setting.GetKeybindings() {
		if km == nil || len(km.GetBindings()) == 0 {
			continue
		}
		mapping := e.mappingConfig.FindByUniversalAction(km.GetId())
		if mapping == nil || mapping.IntelliJ.Action == "" {
			e.logger.Info("no mapping found for action", "action", km.GetId())
			continue
		}
		actionID := mapping.IntelliJ.Action

		for _, b := range km.GetBindings() {
			if b == nil {
				continue
			}
			shortcutXML, err := formatKeybinding(keymap.NewKeyBinding(b))
			if err != nil {
				e.logger.Warn("failed to format keybinding", "action", km.GetId(), "error", err)
				continue
			}

			if _, exists := actionsMap[actionID]; !exists {
				actionsMap[actionID] = &ActionXML{ID: actionID}
				actionOrder = append(actionOrder, actionID)
			}
			actionsMap[actionID].KeyboardShortcuts = append(
				actionsMap[actionID].KeyboardShortcuts,
				KeyboardShortcutXML{First: shortcutXML.First, Second: shortcutXML.Second},
			)
		}
	}

	// Build Actions slice in stable order (by first appearance, then fallback to sort for determinism if empty)
	var actions []ActionXML
	if len(actionOrder) == 0 && len(actionsMap) > 0 {
		for id := range actionsMap {
			actionOrder = append(actionOrder, id)
		}
		sort.Strings(actionOrder)
	}
	for _, id := range actionOrder {
		if ax, ok := actionsMap[id]; ok {
			// Deduplicate keyboard shortcuts per action while preserving first occurrence order
			seen := make(map[string]struct{}, len(ax.KeyboardShortcuts))
			var uniq []KeyboardShortcutXML
			for _, ks := range ax.KeyboardShortcuts {
				key := ks.First + "\x00" + ks.Second
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				uniq = append(uniq, ks)
			}
			ax.KeyboardShortcuts = uniq
			actions = append(actions, *ax)
		}
	}

	return actions
}

// mergeActions merges managed and unmanaged actions, with managed taking priority.
func (e *intellijExporter) mergeActions(managed, unmanaged []ActionXML) []ActionXML {
	// Create a map of managed action IDs for quick lookup
	managedIDs := make(map[string]bool)
	for _, action := range managed {
		managedIDs[action.ID] = true
	}

	result := make([]ActionXML, 0, len(managed)+len(unmanaged))
	result = append(result, managed...)

	// Add unmanaged actions that don't conflict with managed ones
	for _, action := range unmanaged {
		if !managedIDs[action.ID] {
			result = append(result, action)
		} else {
			// Log conflict - managed action takes priority
			e.logger.Debug("Conflict resolved: managed action takes priority",
				"action_id", action.ID)
		}
	}

	return result
}
