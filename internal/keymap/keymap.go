package keymap

import (
	"encoding/json"
	"io"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// OneKeymapConfig is a struct that matches the user config file format.
type OneKeymapConfig struct {
	Id         string `json:"id"`
	Keybinding string `json:"keybinding"`
	Comment    string `json:"comment,omitempty"`
}

// OneKeymapSetting is the root struct for the user config file.
type OneKeymapSetting struct {
	// Version default to 1.0. When having breaking changes, increment the version.
	Version string            `json:"version"`
	Keymaps []OneKeymapConfig `json:"keymaps"`
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
		kb, err := ParseKeyBinding(fk.Keybinding, "+")
		if err != nil {
			// Potentially wrap this error for more context
			return nil, err
		}
		setting.Keybindings = append(setting.Keybindings, &keymapv1.KeyBinding{
			Id:        fk.Id,
			KeyChords: kb.KeyChords,
			Comment:   fk.Comment,
		})
	}

	return setting, nil
}

// Save takes a KeymapSetting proto message and writes it to the given writer
// in the user config file format.
func Save(writer io.Writer, setting *keymapv1.KeymapSetting) error {
	friendlyData := OneKeymapSetting{}
	friendlyData.Version = "1.0"
	for _, k := range setting.Keybindings {
		keybind := NewKeyBinding(k)
		keys, err := keybind.Format(platform.PlatformMacOS, oneKeymapDefaultKeyChordSeparator)
		if err != nil {
			return err
		}
		friendlyData.Keymaps = append(friendlyData.Keymaps, OneKeymapConfig{
			Id:         k.Id,
			Keybinding: keys,
			Comment:    k.Comment,
		})
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Use 2 spaces for indentation
	return encoder.Encode(friendlyData)
}
