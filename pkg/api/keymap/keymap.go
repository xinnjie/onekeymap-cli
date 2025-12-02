package keymap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

type Keymap struct {
	Actions []Action
}

// HasAction returns true if the keymap contains an action with the given name.
func (k *Keymap) HasAction(name string) bool {
	for _, a := range k.Actions {
		if a.Name == name {
			return true
		}
	}
	return false
}

type Action struct {
	Name     string
	Bindings []keybinding.Keybinding
}

const (
	configVersion = "1.0"
)

var (
	errInvalidConfig = errors.New("invalid config format: 'keymaps' field is missing")
)

// LoadOptions provides advanced options for loading a OneKeymap config.
type LoadOptions struct {
	// Reserved for future use
}

// Load reads from reader and builds a keymap.
func Load(reader io.Reader, _ LoadOptions) (Keymap, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return Keymap{}, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return Keymap{}, nil
	}

	friendlyData, err := parseOneKeymapSetting(data)
	if err != nil {
		return Keymap{}, err
	}

	return buildKeymapFromFriendly(friendlyData)
}

// SaveOptions provides advanced options for saving a OneKeymap config.
type SaveOptions struct {
	Platform platform.Platform
}

// Save writes the keymap to the writer in JSON format.
func Save(writer io.Writer, km Keymap, opt SaveOptions) error {
	friendlyData := oneKeymapSetting{}
	friendlyData.Keymaps = make([]oneKeymapConfig, 0)
	friendlyData.Version = configVersion

	// Group keybindings by action name
	grouped := make(map[string]*oneKeymapConfig)
	var order []string

	for _, action := range km.Actions {
		if _, exists := grouped[action.Name]; !exists {
			grouped[action.Name] = &oneKeymapConfig{
				ID:         action.Name,
				Keybinding: make(keybindingStrings, 0),
			}
			order = append(order, action.Name)
		}

		config := grouped[action.Name]
		p := opt.Platform
		if p == "" {
			p = platform.PlatformMacOS
		}

		for _, binding := range action.Bindings {
			keys := binding.String(keybinding.FormatOption{
				Platform:  p,
				Separator: "+",
			})
			config.Keybinding = append(config.Keybinding, keys)
		}
	}

	for _, name := range order {
		friendlyData.Keymaps = append(friendlyData.Keymaps, *grouped[name])
	}

	sort.Slice(friendlyData.Keymaps, func(i, j int) bool {
		return friendlyData.Keymaps[i].ID < friendlyData.Keymaps[j].ID
	})

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Use 2 spaces for indentation
	return encoder.Encode(friendlyData)
}

// oneKeymapSetting is the root struct for the user config file.
type oneKeymapSetting struct {
	Version string            `json:"version"`
	Keymaps []oneKeymapConfig `json:"keymaps"`
}

// oneKeymapConfig is a struct that matches the user config file format.
type oneKeymapConfig struct {
	ID         string            `json:"id,omitempty"`
	Keybinding keybindingStrings `json:"keybinding,omitempty"`
	Comment    string            `json:"comment,omitempty"`
}

// keybindingStrings is a custom type to handle single or multiple keybindings.
type keybindingStrings []string

// UnmarshalJSON allows keybindingStrings to be unmarshalled from either a single string or an array of strings.
func (ks *keybindingStrings) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*ks = []string{s}
		return nil
	}

	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*ks = ss
		return nil
	}

	return errors.New("keybinding must be a string or an array of strings")
}

// MarshalJSON allows keybindingStrings to be marshalled to a single string if it contains only one element.
func (ks keybindingStrings) MarshalJSON() ([]byte, error) {
	if len(ks) == 1 {
		return json.Marshal(ks[0])
	}
	return json.Marshal([]string(ks))
}

// parseOneKeymapSetting parses the raw JSON into `oneKeymapSetting`.
func parseOneKeymapSetting(data []byte) (oneKeymapSetting, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return oneKeymapSetting{}, err
	}

	var friendlyData oneKeymapSetting
	if err := json.Unmarshal(data, &friendlyData); err != nil {
		return oneKeymapSetting{}, err
	}

	var unknownFieldsPresent bool
	for field := range raw {
		switch field {
		case "keymaps", "version":
			// allowed fields
		default:
			unknownFieldsPresent = true
		}
	}

	if len(friendlyData.Keymaps) == 0 && unknownFieldsPresent {
		return oneKeymapSetting{}, errInvalidConfig
	}

	return friendlyData, nil
}

// buildKeymapFromFriendly converts the friendly format to the API Keymap.
func buildKeymapFromFriendly(friendlyData oneKeymapSetting) (Keymap, error) {
	km := Keymap{}
	grouped := make(map[string]*Action)
	var order []string

	for _, fk := range friendlyData.Keymaps {
		action, exists := grouped[fk.ID]
		if !exists {
			action = &Action{
				Name:     fk.ID,
				Bindings: make([]keybinding.Keybinding, 0),
			}
			grouped[fk.ID] = action
			order = append(order, fk.ID)
		}

		for _, keybindingStr := range fk.Keybinding {
			kb, err := keybinding.NewKeybinding(keybindingStr, keybinding.ParseOption{
				Separator: "+",
			})
			if err != nil {
				return Keymap{}, fmt.Errorf(
					"failed to parse keybinding '%s' for id '%s': %w",
					keybindingStr,
					fk.ID,
					err,
				)
			}
			action.Bindings = append(action.Bindings, kb)
		}
	}

	for _, name := range order {
		km.Actions = append(km.Actions, *grouped[name])
	}

	return km, nil
}
