package keymap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

const (
	configVersion = "1.0"
)

var (
	errInvalidConfig = errors.New("invalid config format: 'keymaps' field is missing or null")
)

// OneKeymapConfig is a struct that matches the user config file format.
type OneKeymapConfig struct {
	ID          string            `json:"id,omitempty"`
	Keybinding  KeybindingStrings `json:"keybinding,omitempty"`
	Comment     string            `json:"comment,omitempty"`
	Description string            `json:"description,omitempty"`
	Name        string            `json:"name,omitempty"`
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

	for _, ab := range setting.GetKeybindings() {
		if cfg := config.FindByUniversalAction(ab.GetName()); cfg != nil {
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

// OneKeymapSetting is the root struct for the user config file.
type OneKeymapSetting struct {
	// Version default to 1.0. When having breaking changes, increment the version.
	Version string            `json:"version"`
	Keymaps []OneKeymapConfig `json:"keymaps"`
}

// KeybindingStrings is a custom type to handle single or multiple keybindings.
type KeybindingStrings []string

// UnmarshalJSON allows KeybindingStrings to be unmarshalled from either a single string or an array of strings.
func (ks *KeybindingStrings) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a single string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*ks = []string{s}
		return nil
	}

	// Try to unmarshal as an array of strings
	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*ks = ss
		return nil
	}

	return errors.New("keybinding must be a string or an array of strings")
}

// MarshalJSON allows KeybindingStrings to be marshalled to a single string if it contains only one element.
func (ks KeybindingStrings) MarshalJSON() ([]byte, error) {
	if len(ks) == 1 {
		return json.Marshal(ks[0])
	}
	return json.Marshal([]string(ks))
}

func newAction(fk OneKeymapConfig) *keymapv1.Action {
	ab := &keymapv1.Action{
		Name:    fk.ID,
		Comment: fk.Comment,
	}
	// Only create ActionConfig if there's actual data
	if fk.Description != "" || fk.Name != "" {
		ab.ActionConfig = &keymapv1.ActionConfig{
			Description: fk.Description,
			DisplayName: fk.Name,
		}
	}
	return ab
}

func mergeActionMetadata(ab *keymapv1.Action, fk OneKeymapConfig) {
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
}

// Load reads from the given reader, parses the user config file format,
// and converts it into the internal KeymapSetting proto message.
func Load(reader io.Reader) (*keymapv1.Keymap, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &keymapv1.Keymap{}, nil
	}

	var friendlyData OneKeymapSetting
	if err := json.Unmarshal(data, &friendlyData); err != nil {
		return nil, err
	}

	if friendlyData.Keymaps == nil {
		return nil, errInvalidConfig
	}

	setting := &keymapv1.Keymap{}
	// Group keybindings by Id ONLY. Preserve insertion order of first appearance.
	grouped := make(map[string]*keymapv1.Action)
	order := make([]string, 0)

	for _, fk := range friendlyData.Keymaps {
		key := fk.ID
		ab, ok := grouped[key]
		if !ok {
			ab = newAction(fk)
			grouped[key] = ab
			order = append(order, key)
		} else {
			mergeActionMetadata(ab, fk)
		}

		for _, keybindingStr := range fk.Keybinding {
			kb, err := ParseKeyBinding(keybindingStr, "+")
			if err != nil {
				return nil, fmt.Errorf("failed to parse keybinding '%s' for id '%s': %w", keybindingStr, fk.ID, err)
			}
			ab.Bindings = append(
				ab.Bindings,
				&keymapv1.Binding{KeyChords: kb.KeyChords, KeyChordsReadable: keybindingStr},
			)
		}
	}

	for _, k := range order {
		setting.Keybindings = append(setting.Keybindings, grouped[k])
	}

	return setting, nil
}

// Save takes a KeymapSetting proto message and writes it to the given writer
// in the user config file format.
func Save(writer io.Writer, setting *keymapv1.Keymap) error {
	friendlyData := OneKeymapSetting{}
	friendlyData.Version = configVersion
	// Group keybindings by a composite key of Id and Comment.
	type groupKey struct {
		ID          string
		Comment     string
		Description string
	}
	groupedKeybindings := make(map[groupKey]*OneKeymapConfig)

	for _, k := range setting.GetKeybindings() {
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
			keys, err := binding.Format(platform.PlatformMacOS, "+")
			if err != nil {
				return err
			}
			config.Keybinding = append(config.Keybinding, keys)
		}
	}

	for _, config := range groupedKeybindings {
		friendlyData.Keymaps = append(friendlyData.Keymaps, *config)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Use 2 spaces for indentation
	return encoder.Encode(friendlyData)
}
