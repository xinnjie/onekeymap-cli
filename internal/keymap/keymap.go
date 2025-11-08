package keymap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	configVersion = "1.0"
)

var (
	errInvalidConfig = errors.New("invalid config format: 'keymaps' field is missing")
)

// LoadOptions provides advanced options for loading a OneKeymap config.
type LoadOptions struct {
	MappingConfig *mappings.MappingConfig
}

// Load reads from reader and builds a keymap.
func Load(reader io.Reader, opt LoadOptions) (*keymapv1.Keymap, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &keymapv1.Keymap{}, nil
	}

	friendlyData, err := parseOneKeymapSetting(data)
	if err != nil {
		return nil, err
	}

	return buildKeymapFromFriendly(friendlyData, opt.MappingConfig)
}

type SaveOptions struct {
	Platform platform.Platform
}

func Save(writer io.Writer, setting *keymapv1.Keymap, opt SaveOptions) error {
	friendlyData := OneKeymapSetting{}
	friendlyData.Keymaps = make([]OneKeymapConfig, 0)
	friendlyData.Version = configVersion
	// Group keybindings by a composite key of Id and Comment.
	type groupKey struct {
		ID          string
		Comment     string
		Description string
	}
	groupedKeybindings := make(map[groupKey]*OneKeymapConfig)

	for _, k := range setting.GetActions() {
		var description, displayName string
		if k.GetActionConfig() != nil {
			description = k.GetActionConfig().GetDescription()
			displayName = k.GetActionConfig().GetDisplayName()
		}
		key := groupKey{ID: k.GetName(), Comment: k.GetComment(), Description: description}
		config, ok := groupedKeybindings[key]
		if !ok {
			config = &OneKeymapConfig{
				ID:          k.GetName(),
				Comment:     k.GetComment(),
				Description: description,
				Name:        displayName,
			}
			groupedKeybindings[key] = config
		}

		for _, b := range k.GetBindings() {
			if b == nil || len(b.GetKeyChords().GetChords()) == 0 {
				continue
			}
			binding := NewKeyBinding(b)
			p := opt.Platform
			if p == "" {
				p = platform.PlatformMacOS
			}
			keys, err := binding.Format(p, "+")
			if err != nil {
				return err
			}
			config.Keybinding = append(config.Keybinding, keys)
		}
	}

	for _, config := range groupedKeybindings {
		friendlyData.Keymaps = append(friendlyData.Keymaps, *config)
	}

	sort.Slice(friendlyData.Keymaps, func(i, j int) bool {
		if friendlyData.Keymaps[i].ID != friendlyData.Keymaps[j].ID {
			return friendlyData.Keymaps[i].ID < friendlyData.Keymaps[j].ID
		}
		return false
	})

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Use 2 spaces for indentation
	return encoder.Encode(friendlyData)
}

// DecorateSetting applies metadata (Description, Name, Category) to each
// KeyBinding in the provided KeymapSetting using the given MappingConfig.
// It also fills the KeyChordsReadable field when chords are present.
func DecorateSetting(
	setting *keymapv1.Keymap,
	config *mappings.MappingConfig,
) *keymapv1.Keymap {
	if setting == nil || config == nil {
		return setting
	}

	for _, ab := range setting.GetActions() {
		if cfg := config.Get(ab.GetName()); cfg != nil {
			if ab.GetActionConfig() == nil {
				ab.ActionConfig = &keymapv1.ActionConfig{}
			}
			ab.ActionConfig.Description = cfg.Description
			ab.ActionConfig.DisplayName = cfg.Name
			ab.ActionConfig.Category = cfg.Category
		}

		for _, b := range ab.GetBindings() {
			if b != nil && b.GetKeyChords() != nil && len(b.GetKeyChords().GetChords()) > 0 {
				kb := NewKeyBinding(b)
				if formatted, err := kb.Format(platform.PlatformMacOS, "+"); err == nil {
					b.KeyChordsReadable = formatted
				}
			}
		}
	}

	return setting
}

func newAction(fk OneKeymapConfig, mc *mappings.MappingConfig) *keymapv1.Action {
	ab := &keymapv1.Action{
		Name:    fk.ID,
		Comment: fk.Comment,
	}
	// Only create ActionConfig if there's actual data or editor support
	needsConfig := fk.Description != "" || fk.Name != "" || mc != nil
	if needsConfig {
		ab.ActionConfig = &keymapv1.ActionConfig{
			Description: fk.Description,
			DisplayName: fk.Name,
		}
		// Populate editor support from mapping config
		if mc != nil {
			if mapping := mc.Get(fk.ID); mapping != nil {
				ab.ActionConfig.EditorSupport = BuildEditorSupportFromMapping(mapping)
			}
		}
	}
	return ab
}

// BuildEditorSupportFromMapping constructs EditorSupport slice from an ActionMappingConfig.
func BuildEditorSupportFromMapping(mapping *mappings.ActionMappingConfig) []*keymapv1.EditorSupport {
	if mapping == nil {
		return nil
	}

	// Define all main editor types to check
	editorTypes := []pluginapi.EditorType{
		pluginapi.EditorTypeVSCode,
		pluginapi.EditorTypeIntelliJ,
		pluginapi.EditorTypeZed,
		pluginapi.EditorTypeVim,
		pluginapi.EditorTypeXcode,
		pluginapi.EditorTypeHelix,
	}

	var result []*keymapv1.EditorSupport
	for _, editorType := range editorTypes {
		supported, notSupportedReason := mapping.IsSupported(editorType)

		result = append(result, &keymapv1.EditorSupport{
			EditorType: editorType.ToAPI(),
			Supported:  supported,
			Note:       notSupportedReason,
		})
	}

	return result
}

func mergeActionMetadata(ab *keymapv1.Action, fk OneKeymapConfig, mc *mappings.MappingConfig) {
	// Preserve first non-empty metadata.
	if ab.GetComment() == "" && fk.Comment != "" {
		ab.Comment = fk.Comment
	}
	if fk.Description != "" || fk.Name != "" {
		if ab.GetActionConfig() == nil {
			ab.ActionConfig = &keymapv1.ActionConfig{}
		}
		if ab.GetActionConfig().GetDescription() == "" && fk.Description != "" {
			ab.ActionConfig.Description = fk.Description
		}
		if ab.GetActionConfig().GetDisplayName() == "" && fk.Name != "" {
			ab.ActionConfig.DisplayName = fk.Name
		}
	}
	// Populate editor support if not already set and mapping config is available
	if mc != nil && ab.GetActionConfig() != nil && len(ab.GetActionConfig().GetEditorSupport()) == 0 {
		if mapping := mc.Get(fk.ID); mapping != nil {
			ab.ActionConfig.EditorSupport = BuildEditorSupportFromMapping(mapping)
		}
	}
}

// parseOneKeymapSetting parses the raw JSON into `OneKeymapSetting` and returns
// `errInvalidConfig` when `keymaps` is explicitly null or absent while other
// unsupported top-level fields exist.
func parseOneKeymapSetting(data []byte) (OneKeymapSetting, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return OneKeymapSetting{}, err
	}

	var friendlyData OneKeymapSetting
	if err := json.Unmarshal(data, &friendlyData); err != nil {
		return OneKeymapSetting{}, err
	}

	var (
		unknownFieldsPresent bool
	)
	for field := range raw {
		switch field {
		case "keymaps":
		case "version":
			// allowed field
		default:
			unknownFieldsPresent = true
		}
	}

	if len(friendlyData.Keymaps) == 0 && unknownFieldsPresent {
		return OneKeymapSetting{}, errInvalidConfig
	}

	return friendlyData, nil
}

func buildKeymapFromFriendly(friendlyData OneKeymapSetting, mc *mappings.MappingConfig) (*keymapv1.Keymap, error) {
	setting := &keymapv1.Keymap{}
	grouped := make(map[string]*keymapv1.Action)
	order := make([]string, 0)

	for _, fk := range friendlyData.Keymaps {
		key := fk.ID
		ab, ok := grouped[key]
		if !ok {
			ab = newAction(fk, mc)
			grouped[key] = ab
			order = append(order, key)
		} else {
			mergeActionMetadata(ab, fk, mc)
		}

		for _, keybindingStr := range fk.Keybinding {
			kb, err := ParseKeyBinding(keybindingStr, "+")
			if err != nil {
				return nil, fmt.Errorf("failed to parse keybinding '%s' for id '%s': %w", keybindingStr, fk.ID, err)
			}
			ab.Bindings = append(
				ab.Bindings,
				&keymapv1.KeybindingReadable{KeyChords: kb.KeyChords, KeyChordsReadable: keybindingStr},
			)
		}
	}

	for _, k := range order {
		setting.Actions = append(setting.Actions, grouped[k])
	}
	return setting, nil
}
