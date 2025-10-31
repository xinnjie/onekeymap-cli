package mappings

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// XcodeMappingConfig defines the structure for Xcode's mapping,
// including the action and optional command context.
type XcodeMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	// The Xcode action name (e.g., "moveWordLeft:", "selectWord:")
	Action string `yaml:"action"`
	// The command group ID
	CommandGroupID string `yaml:"commandGroupID,omitempty"`
	// The Xcode command ID for menu bindings (optional)
	CommandID string `yaml:"commandID,omitempty"`
	// Whether this is an alternate key binding
	Alternate string `yaml:"alternate,omitempty"`
	// The menu group this action belongs to
	Group string `yaml:"group,omitempty"`
	// The menu group ID
	GroupID string `yaml:"groupID,omitempty"`
	// Whether this is a grouped alternate key binding
	GroupedAlternate string `yaml:"groupedAlternate,omitempty"`
	// Whether this is a navigation action
	Navigation string `yaml:"navigation,omitempty"`
	// The parent title for nested menu items
	ParentTitle string `yaml:"parentTitle,omitempty"`
	// The title of the action
	Title string `yaml:"title,omitempty"`
}

// XcodeConfigs is a slice of XcodeMappingConfig that can be unmarshalled from either
// a single YAML object or a sequence of objects.
type XcodeConfigs []XcodeMappingConfig

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (x *XcodeConfigs) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		// It's a single object, unmarshal it and wrap in a slice.
		var single XcodeMappingConfig
		if err := node.Decode(&single); err != nil {
			return err
		}
		*x = []XcodeMappingConfig{single}
	case yaml.SequenceNode:
		// It's a sequence, unmarshal it directly into the slice.
		var slice []XcodeMappingConfig
		if err := node.Decode(&slice); err != nil {
			return err
		}
		*x = slice
	default:
		return fmt.Errorf(
			"cannot unmarshal!! (line %d, col %d): expected a mapping or sequence node for xcode config",
			node.Line,
			node.Column,
		)
	}
	return nil
}

func checkXcodeDuplicateConfig(mappings map[string]ActionMappingConfig) error {
	seen := make(map[struct{ Action, CommandID string }]string)
	dups := make(map[string][]string) // key string -> list of universal action IDs
	for id, mapping := range mappings {
		for _, xcodeConfig := range mapping.Xcode {
			if xcodeConfig.Action == "" {
				continue
			}
			// Skip configs that are disabled for import (export-only)
			if xcodeConfig.DisableImport {
				continue
			}

			key := struct{ Action, CommandID string }{xcodeConfig.Action, xcodeConfig.CommandID}
			if originalID, exists := seen[key]; exists {
				dupKey := fmt.Sprintf(`{"action":%q,"commandID":%q}`, key.Action, key.CommandID)
				if _, ok := dups[dupKey]; !ok {
					dups[dupKey] = []string{originalID}
				}
				dups[dupKey] = append(dups[dupKey], id)
				continue
			}
			seen[key] = id
		}
	}
	if len(dups) == 0 {
		return nil
	}
	return &DuplicateActionMappingError{Editor: "xcode", Duplicates: dups}
}
