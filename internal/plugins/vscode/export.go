package vscode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

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

func newExporter(mappingConfig *mappings.MappingConfig, logger *slog.Logger, differ diff.Differ) pluginapi.PluginExporter {
	return &vscodeExporter{
		mappingConfig: mappingConfig,
		logger:        logger,
		differ:        differ,
	}
}

func (e *vscodeExporter) Export(ctx context.Context, destination io.Writer, setting *keymapv1.KeymapSetting, opts pluginapi.PluginExportOption) (*pluginapi.PluginExportReport, error) {
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

	jsonData, err := json.MarshalIndent(finalKeybindings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vscode keybindings to json: %w", err)
	}

	_, err = destination.Write(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to write to destination: %w", err)
	}

	diff, err := e.calculateDiff(opts.Base, finalKeybindings)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate diff: %w", err)
	}
	return &pluginapi.PluginExportReport{Diff: &diff}, nil
}

// identifyUnmanagedKeybindings performs reverse lookup to identify keybindings
// that are not managed by onekeymap
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
// corresponds to any action in our mappings
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

// generateManagedKeybindings generates VSCode keybindings from KeymapSetting
func (e *vscodeExporter) generateManagedKeybindings(setting *keymapv1.KeymapSetting) []vscodeKeybinding {
	var vscodeKeybindings []vscodeKeybinding

	for _, km := range setting.Keybindings {
		mapping := e.mappingConfig.FindByUniversalAction(km.Action)
		if mapping == nil {
			continue
		}

		vscodeConfigs := mapping.VSCode
		if len(vscodeConfigs) == 0 {
			continue
		}

		binding := keymap.NewKeyBinding(km)
		keys, err := formatKeybinding(binding)
		if err != nil {
			e.logger.Warn("Skipping keybinding with un-formattable key", "key", km.KeyChords, "error", err)
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

	return vscodeKeybindings
}

// mergeKeybindings merges managed and unmanaged keybindings, with managed taking priority
func (e *vscodeExporter) mergeKeybindings(managed, unmanaged []vscodeKeybinding) []vscodeKeybinding {
	// Create a map to track managed keybindings by their key combination
	managedKeys := make(map[string]bool)
	for _, kb := range managed {
		managedKeys[kb.Key] = true
	}

	// Start with all managed keybindings
	result := make([]vscodeKeybinding, len(managed))
	copy(result, managed)

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

// equalVSCodeArgs compares VSCode args, handling the conversion between different types
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

func (e *vscodeExporter) calculateDiff(base io.Reader, vscodeKeybindings vscodeKeybindingConfig) (string, error) {
	var before vscodeKeybindingConfig
	if base == nil {
		before = vscodeKeybindingConfig{}
	} else {
		if err := json.NewDecoder(base).Decode(&before); err != nil {
			return "", fmt.Errorf("failed to decode base: %w", err)
		}
	}

	d, err := e.differ.Diff(before, vscodeKeybindings)
	if err != nil {
		return "", fmt.Errorf("failed to calculate diff: %w", err)
	}

	return d, nil
}
