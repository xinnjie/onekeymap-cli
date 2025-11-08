package mappings

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// XcodeMappingConfig defines the structure for Xcode's mapping,
// including the action and optional command context.
type XcodeMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	// A Xcode action is either a Text Action or a Menu Action.
	// Only one of TextAction and MenuAction should be set.

	// The Xcode text action name(s) for Text Key Bindings. Accepts a string or array in YAML.
	TextAction XcodeTextAction `yaml:",inline"`

	// The Xcode menu action name for Menu Key Bindings (e.g., "moveWordLeft:", "selectWord:")
	MenuAction XcodeMenuAction `yaml:",inline"`
}

func checkTextBindings(items []string, id string, seen map[string]string, dups map[string][]string) {
	for _, action := range items {
		action = ensureTrailingColon(action)
		if action == "" {
			continue
		}
		if originalID, exists := seen[action]; exists {
			dupKey := fmt.Sprintf(`{"textAction":%q}`, action)
			if _, ok := dups[dupKey]; !ok {
				dups[dupKey] = []string{originalID}
			}
			dups[dupKey] = append(dups[dupKey], id)
		} else {
			seen[action] = id
		}
	}
}

func checkXcodeTextActionFormat(mappings map[string]ActionMappingConfig) error {
	var invalidIDs []string
	for id, mapping := range mappings {
		for _, xc := range mapping.Xcode {
			if len(xc.TextAction.TextAction.Items) == 0 {
				continue
			}
			for _, a := range xc.TextAction.TextAction.Items {
				if a == "" {
					continue
				}
				if !strings.HasSuffix(a, ":") {
					invalidIDs = append(invalidIDs, id)
					break
				}
			}
		}
	}
	if len(invalidIDs) > 0 {
		return fmt.Errorf("xcode textAction must end with ':' for ids: %v", invalidIDs)
	}
	return nil
}

// checkXcodeImportConstraints enforces that when a textAction defines multiple actions
// (i.e., an array with length > 1), the config must be export-only: disableImport: true.
func checkXcodeImportConstraints(mappings map[string]ActionMappingConfig) error {
	var invalidIDs []string
	for id, mapping := range mappings {
		for _, xc := range mapping.Xcode {
			if len(xc.TextAction.TextAction.Items) > 1 && !xc.DisableImport {
				invalidIDs = append(invalidIDs, id)
				break
			}
		}
	}
	if len(invalidIDs) > 0 {
		return fmt.Errorf("xcode textAction with multiple items requires disableImport: true for ids: %v", invalidIDs)
	}
	return nil
}

type XcodeTextAction struct {
	TextAction StringOrSlice `yaml:"textAction,omitempty"`
}

// StringOrSlice supports unmarshalling a YAML field that may be a single string
// or a sequence of strings.
type StringOrSlice struct {
	Items []string
}

func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	if node == nil || node.Kind == 0 {
		s.Items = nil
		return nil
	}
	switch node.Kind {
	case yaml.ScalarNode:
		var str string
		if err := node.Decode(&str); err != nil {
			return err
		}
		s.Items = []string{str}
		return nil
	case yaml.SequenceNode:
		var arr []string
		if err := node.Decode(&arr); err != nil {
			return err
		}
		s.Items = arr
		return nil
	case yaml.MappingNode:
		// Not expected for a textAction value
		return fmt.Errorf("cannot unmarshal mapping node for textAction (line %d, col %d)", node.Line, node.Column)
	default:
		return fmt.Errorf(
			"cannot unmarshal (line %d, col %d): unexpected node kind for textAction",
			node.Line,
			node.Column,
		)
	}
}

func ensureTrailingColon(s string) string {
	if s == "" {
		return s
	}
	if strings.HasSuffix(s, ":") {
		return s
	}
	return s + ":"
}

type XcodeMenuAction struct {
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
	seenMenuBindings := make(map[struct{ Action, CommandID string }]string)
	seenTextBindings := make(map[string]string)
	dups := make(map[string][]string) // key string -> list of universal action IDs

	for id, mapping := range mappings {
		for _, xcodeConfig := range mapping.Xcode {
			// Skip configs that are disabled for import (export-only)
			if xcodeConfig.DisableImport {
				continue
			}

			// Check Menu Key Bindings (Action + CommandID)
			if xcodeConfig.MenuAction.Action != "" {
				key := struct{ Action, CommandID string }{
					xcodeConfig.MenuAction.Action,
					xcodeConfig.MenuAction.CommandID,
				}
				if originalID, exists := seenMenuBindings[key]; exists {
					dupKey := fmt.Sprintf(`{"action":%q,"commandID":%q}`, key.Action, key.CommandID)
					if _, ok := dups[dupKey]; !ok {
						dups[dupKey] = []string{originalID}
					}
					dups[dupKey] = append(dups[dupKey], id)
				} else {
					seenMenuBindings[key] = id
				}
			}

			// Check Text Key Bindings (each TextAction item)
			checkTextBindings(xcodeConfig.TextAction.TextAction.Items, id, seenTextBindings, dups)
		}
	}

	if len(dups) == 0 {
		return nil
	}
	return &DuplicateActionMappingError{Editor: "xcode", Duplicates: dups}
}
