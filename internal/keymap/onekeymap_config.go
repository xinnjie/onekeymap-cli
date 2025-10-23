package keymap

import (
	"encoding/json"
	"errors"
)

// OneKeymapSetting is the root struct for the user config file.
type OneKeymapSetting struct {
	// Version default to 1.0. When having breaking changes, increment the version.
	Version string            `json:"version"`
	Keymaps []OneKeymapConfig `json:"keymaps"`
}

// OneKeymapConfig is a struct that matches the user config file format.
type OneKeymapConfig struct {
	ID          string            `json:"id,omitempty"`
	Keybinding  KeybindingStrings `json:"keybinding,omitempty"`
	Comment     string            `json:"comment,omitempty"`
	Description string            `json:"description,omitempty"`
	Name        string            `json:"name,omitempty"`
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
