package helix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/pelletier/go-toml/v2"
	"github.com/xinnjie/onekeymap-cli/internal/diff"
	"github.com/xinnjie/onekeymap-cli/internal/export"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

type helixExporter struct {
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	differ        diff.Differ
}

func newExporter(
	mappingConfig *mappings.MappingConfig,
	logger *slog.Logger,
	differ diff.Differ,
) pluginapi.PluginExporter {
	return &helixExporter{mappingConfig: mappingConfig, logger: logger, differ: differ}
}

func (e *helixExporter) Export(
	ctx context.Context,
	destination io.Writer,
	setting keymap.Keymap,
	opts pluginapi.PluginExportOption,
) (*pluginapi.PluginExportReport, error) {
	// Read existing configuration if provided for non-destructive export
	var existingKeys helixKeys
	var existingFullConfig map[string]interface{}
	if opts.ExistingConfig != nil {
		existingKeys, existingFullConfig = e.parseExistingConfig(ctx, opts.ExistingConfig)
	}

	// Identify unmanaged keybindings from existing config
	var unmanagedKeys helixKeys
	if opts.ExistingConfig != nil {
		unmanagedKeys = e.identifyUnmanagedKeybindings(existingKeys)
	}

	// Generate managed keybindings from current setting
	marker := export.NewMarker(&setting)
	managedKeys := e.generateManagedKeybindings(ctx, &setting, marker)

	// Merge managed and unmanaged keybindings
	finalKeys := e.mergeKeybindings(ctx, managedKeys, unmanagedKeys)

	// Create final configuration preserving other sections
	finalConfig := make(map[string]interface{})

	// Copy all existing sections except keys
	for k, v := range existingFullConfig {
		if k != "keys" {
			finalConfig[k] = v
		}
	}

	// Add the merged keys section
	finalConfig["keys"] = e.convertHelixKeysToMap(finalKeys)

	// Encode the final configuration
	if err := toml.NewEncoder(destination).Encode(finalConfig); err != nil {
		return nil, fmt.Errorf("failed to encode helix toml: %w", err)
	}

	return &pluginapi.PluginExportReport{
		BaseEditorConfig:   existingKeys,
		ExportEditorConfig: finalKeys,
		SkipReport:         marker.Report(),
	}, nil
}

func (e *helixExporter) parseExistingConfig(
	ctx context.Context,
	existingConfig io.Reader,
) (helixKeys, map[string]interface{}) {
	var existingKeys helixKeys
	var existingFullConfig map[string]interface{}

	// Parse as generic map to preserve all sections
	if err := toml.NewDecoder(existingConfig).Decode(&existingFullConfig); err != nil {
		e.logger.WarnContext(
			ctx,
			"Failed to parse existing config, proceeding with destructive export",
			"error",
			err,
		)
		return existingKeys, existingFullConfig
	}

	// Extract keys section if it exists
	if keysSection, ok := existingFullConfig["keys"]; ok {
		if keysMap, ok := keysSection.(map[string]interface{}); ok {
			existingKeys = e.convertMapToHelixKeys(keysMap)
		}
	}

	return existingKeys, existingFullConfig
}

// identifyUnmanagedKeybindings performs reverse lookup to identify keybindings
// that are not managed by onekeymap.
func (e *helixExporter) identifyUnmanagedKeybindings(existingKeys helixKeys) helixKeys {
	unmanaged := helixKeys{}

	// Check each mode
	if existingKeys.Normal != nil {
		unmanaged.Normal = make(map[string]string)
		for key, command := range existingKeys.Normal {
			if !e.isManagedKeybinding(command, HelixModeNormal) {
				unmanaged.Normal[key] = command
			}
		}
		if len(unmanaged.Normal) == 0 {
			unmanaged.Normal = nil
		}
	}

	if existingKeys.Insert != nil {
		unmanaged.Insert = make(map[string]string)
		for key, command := range existingKeys.Insert {
			if !e.isManagedKeybinding(command, HelixModeInsert) {
				unmanaged.Insert[key] = command
			}
		}
		if len(unmanaged.Insert) == 0 {
			unmanaged.Insert = nil
		}
	}

	if existingKeys.Select != nil {
		unmanaged.Select = make(map[string]string)
		for key, command := range existingKeys.Select {
			if !e.isManagedKeybinding(command, HelixModeSelect) {
				unmanaged.Select[key] = command
			}
		}
		if len(unmanaged.Select) == 0 {
			unmanaged.Select = nil
		}
	}

	return unmanaged
}

// isManagedKeybinding checks if a keybinding is managed by onekeymap.
func (e *helixExporter) isManagedKeybinding(command string, mode Mode) bool {
	for _, mapping := range e.mappingConfig.Mappings {
		for _, hconf := range mapping.Helix {
			if hconf.Command == command {
				var mappingMode Mode
				if hconf.Mode == "" {
					mappingMode = HelixModeNormal
				} else {
					mappingMode = Mode(hconf.Mode)
				}
				if mappingMode == mode {
					return true
				}
			}
		}
	}
	return false
}

// generateManagedKeybindings generates Helix keybindings from KeymapSetting.
func (e *helixExporter) generateManagedKeybindings(
	ctx context.Context,
	setting *keymap.Keymap,
	marker *export.Marker,
) helixKeys {
	keysByMode := helixKeys{}

	for _, km := range setting.Actions {
		mapping, usedFallback := e.mappingConfig.GetExportAction(km.Name, pluginapi.EditorTypeHelix)
		if mapping == nil || len(mapping.Helix) == 0 {
			for _, b := range km.Bindings {
				if len(b.KeyChords) > 0 {
					marker.MarkSkippedForReason(km.Name, &b, pluginapi.ErrActionNotSupported)
				}
			}
			continue
		}

		if usedFallback {
			e.logger.InfoContext(ctx, "Action not directly supported, using fallback action",
				"originalAction", km.Name,
				"fallbackAction", mapping.ID,
			)
		}

		for _, b := range km.Bindings {
			if len(b.KeyChords) == 0 {
				continue
			}
			keyStr, err := formatKeybinding(b)
			if err != nil {
				// TODO(xinnjie): Add doc about this behavior: because helix do not recognize numpad keys(numpad1 is recognized as "1"), to avoid conflict with other keybindings, we skip these keybindings
				if errors.Is(err, ErrNotSupportKeyChords) {
					e.logger.DebugContext(
						ctx,
						"Skipping keybinding with unsupported key chords",
						"action",
						km.Name,
					)
					marker.MarkSkippedForReason(
						km.Name,
						&b,
						&pluginapi.UnsupportedExportActionError{Note: err.Error()},
					)
				} else {
					e.logger.WarnContext(ctx, "Skipping keybinding with un-formattable key", "action", km.Name, "error", err)
					marker.MarkSkippedForReason(km.Name, &b, &pluginapi.UnsupportedExportActionError{Note: err.Error()})
				}
				continue
			}
			marker.MarkExported(km.Name, b)

			for _, hconf := range mapping.Helix {
				if hconf.Command == "" {
					continue
				}
				var m Mode
				if hconf.Mode == "" {
					m = HelixModeNormal
				} else {
					m = Mode(hconf.Mode)
				}

				var dest *map[string]string
				switch m {
				case HelixModeNormal:
					if keysByMode.Normal == nil {
						keysByMode.Normal = make(map[string]string)
					}
					dest = &keysByMode.Normal
				case HelixModeInsert:
					if keysByMode.Insert == nil {
						keysByMode.Insert = make(map[string]string)
					}
					dest = &keysByMode.Insert
				case HelixModeSelect:
					if keysByMode.Select == nil {
						keysByMode.Select = make(map[string]string)
					}
					dest = &keysByMode.Select
				default:
					e.logger.WarnContext(
						ctx,
						"Unsupported Helix mode; skipping",
						"mode",
						string(m),
						"action",
						km.Name,
					)
					continue
				}
				(*dest)[keyStr] = hconf.Command
			}
		}
	}

	return keysByMode
}

// mergeKeybindings merges managed and unmanaged keybindings, with managed taking priority.
func (e *helixExporter) mergeKeybindings(ctx context.Context, managed, unmanaged helixKeys) helixKeys {
	result := helixKeys{}

	// Start with managed keybindings
	if managed.Normal != nil {
		result.Normal = make(map[string]string)
		for k, v := range managed.Normal {
			result.Normal[k] = v
		}
	}
	if managed.Insert != nil {
		result.Insert = make(map[string]string)
		for k, v := range managed.Insert {
			result.Insert[k] = v
		}
	}
	if managed.Select != nil {
		result.Select = make(map[string]string)
		for k, v := range managed.Select {
			result.Select[k] = v
		}
	}

	// Add unmanaged keybindings that don't conflict
	if unmanaged.Normal != nil {
		if result.Normal == nil {
			result.Normal = make(map[string]string)
		}
		for key, command := range unmanaged.Normal {
			if _, exists := result.Normal[key]; !exists {
				result.Normal[key] = command
			} else {
				e.logger.DebugContext(ctx, "Conflict resolved: managed keybinding takes priority",
					"mode", "normal", "key", key, "unmanaged_command", command)
			}
		}
	}

	if unmanaged.Insert != nil {
		if result.Insert == nil {
			result.Insert = make(map[string]string)
		}
		for key, command := range unmanaged.Insert {
			if _, exists := result.Insert[key]; !exists {
				result.Insert[key] = command
			} else {
				e.logger.DebugContext(ctx, "Conflict resolved: managed keybinding takes priority",
					"mode", "insert", "key", key, "unmanaged_command", command)
			}
		}
	}

	if unmanaged.Select != nil {
		if result.Select == nil {
			result.Select = make(map[string]string)
		}
		for key, command := range unmanaged.Select {
			if _, exists := result.Select[key]; !exists {
				result.Select[key] = command
			} else {
				e.logger.DebugContext(ctx, "Conflict resolved: managed keybinding takes priority",
					"mode", "select", "key", key, "unmanaged_command", command)
			}
		}
	}

	return result
}

// convertMapToHelixKeys converts a generic map to helixKeys structure.
func (e *helixExporter) convertMapToHelixKeys(keysMap map[string]interface{}) helixKeys {
	keys := helixKeys{}

	if normalMap, ok := keysMap["normal"].(map[string]interface{}); ok {
		keys.Normal = make(map[string]string)
		for k, v := range normalMap {
			if str, ok := v.(string); ok {
				keys.Normal[k] = str
			}
		}
	}

	if insertMap, ok := keysMap["insert"].(map[string]interface{}); ok {
		keys.Insert = make(map[string]string)
		for k, v := range insertMap {
			if str, ok := v.(string); ok {
				keys.Insert[k] = str
			}
		}
	}

	if selectMap, ok := keysMap["select"].(map[string]interface{}); ok {
		keys.Select = make(map[string]string)
		for k, v := range selectMap {
			if str, ok := v.(string); ok {
				keys.Select[k] = str
			}
		}
	}

	return keys
}

// convertHelixKeysToMap converts helixKeys structure to a generic map.
func (e *helixExporter) convertHelixKeysToMap(keys helixKeys) map[string]interface{} {
	result := make(map[string]interface{})

	if len(keys.Normal) > 0 {
		result["normal"] = keys.Normal
	}

	if len(keys.Insert) > 0 {
		result["insert"] = keys.Insert
	}

	if len(keys.Select) > 0 {
		result["select"] = keys.Select
	}

	return result
}
