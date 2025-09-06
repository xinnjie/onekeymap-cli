package keymap

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// OneKeymapConfig is a struct that matches the user config file format.
type OneKeymapConfig struct {
	Id          string            `json:"id"`
	Keybinding  KeybindingStrings `json:"keybinding"`
	Comment     string            `json:"comment,omitempty"`
	Description string            `json:"description,omitempty"`
	Name        string            `json:"name,omitempty"`
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

	return fmt.Errorf("keybinding must be a string or an array of strings")
}

// MarshalJSON allows KeybindingStrings to be marshalled to a single string if it contains only one element.
func (ks KeybindingStrings) MarshalJSON() ([]byte, error) {
	if len(ks) == 1 {
		return json.Marshal(ks[0])
	}
	return json.Marshal([]string(ks))
}

// Load reads from the given reader, parses the user config file format,
// and converts it into the internal KeymapSetting proto message.
func Load(reader io.Reader) (*keymapv1.KeymapSetting, error) {
	decoder := json.NewDecoder(reader)
	var friendlyData OneKeymapSetting
	if err := decoder.Decode(&friendlyData); err != nil {
		return nil, err
	}

	setting := &keymapv1.KeymapSetting{}
	for _, fk := range friendlyData.Keymaps {
		for _, keybindingStr := range fk.Keybinding {
			kb, err := ParseKeyBinding(keybindingStr, "+")
			if err != nil {
				// Potentially wrap this error for more context
				return nil, fmt.Errorf("failed to parse keybinding '%s' for id '%s': %w", keybindingStr, fk.Id, err)
			}
			setting.Keybindings = append(setting.Keybindings, &keymapv1.KeyBinding{
				Id:        fk.Id,
				KeyChords: kb.KeyChords,
				Comment:   fk.Comment,
			})
		}
	}

	return setting, nil
}

// Save takes a KeymapSetting proto message and writes it to the given writer
// in the user config file format.
func Save(writer io.Writer, setting *keymapv1.KeymapSetting) error {
	friendlyData := OneKeymapSetting{}
	friendlyData.Version = "1.0"
	// Group keybindings by a composite key of Id and Comment.
	type groupKey struct {
		ID          string
		Comment     string
		Description string
	}
	groupedKeybindings := make(map[groupKey]*OneKeymapConfig)

	for _, k := range setting.Keybindings {
		key := groupKey{ID: k.Id, Comment: k.Comment, Description: k.Description}
		config, ok := groupedKeybindings[key]
		if !ok {
			config = &OneKeymapConfig{
				Id:          k.Id,
				Comment:     k.Comment,
				Description: k.Description,
				Name:        k.Name,
			}
			groupedKeybindings[key] = config
		}

		binding := NewKeyBinding(k)
		keys, err := binding.Format(platform.PlatformMacOS, "+")
		if err != nil {
			return err
		}
		config.Keybinding = append(config.Keybinding, keys)
	}

	for _, config := range groupedKeybindings {
		friendlyData.Keymaps = append(friendlyData.Keymaps, *config)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Use 2 spaces for indentation
	return encoder.Encode(friendlyData)
}
