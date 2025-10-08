package mappings

import (
	"encoding/json"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// VscodeMappingConfig defines the structure for VSCode's mapping,
// including the command and its context.
type VscodeMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	Command string                 `yaml:"command"`
	When    string                 `yaml:"when"`
	Args    map[string]interface{} `yaml:"args,omitempty"`
}

// VscodeConfigs is a slice of VscodeMappingConfig that can be unmarshalled from either
// a single YAML object or a sequence of objects.
type VscodeConfigs []VscodeMappingConfig

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (v *VscodeConfigs) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		// It's a single object, unmarshal it and wrap in a slice.
		var single VscodeMappingConfig
		if err := node.Decode(&single); err != nil {
			return err
		}
		*v = []VscodeMappingConfig{single}
	case yaml.SequenceNode:
		// It's a sequence, unmarshal it directly into the slice.
		var slice []VscodeMappingConfig
		if err := node.Decode(&slice); err != nil {
			return err
		}
		*v = slice
	default:
		return fmt.Errorf(
			"cannot unmarshal!! (line %d, col %d): expected a mapping or sequence node for vscode config",
			node.Line,
			node.Column,
		)
	}
	return nil
}

func (v *VscodeConfigs) HasExplicitForImport() bool {
	return slices.ContainsFunc(*v, func(cfg VscodeMappingConfig) bool {
		return cfg.ForImport
	})
}

func checkVscodeDuplicateConfig(mappings map[string]ActionMappingConfig) error {
	seen := make(map[struct{ Command, When, Args string }]string)
	dups := make(map[string][]string) // key string -> list of universal action IDs
	for id, mapping := range mappings {
		for _, vscodeConfig := range mapping.VSCode {
			if vscodeConfig.Command == "" {
				continue
			}
			// Serialize args to string for comparison
			var argsStr string
			if vscodeConfig.Args != nil {
				argsBytes, _ := json.Marshal(vscodeConfig.Args)
				argsStr = string(argsBytes)
			}

			key := struct{ Command, When, Args string }{vscodeConfig.Command, vscodeConfig.When, argsStr}
			if originalID, exists := seen[key]; exists {
				dupKey := fmt.Sprintf(`{"command":%q,"when":%q,"args":%q}`, key.Command, key.When, key.Args)
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
	return &DuplicateActionMappingError{Editor: "vscode", Duplicates: dups}
}
