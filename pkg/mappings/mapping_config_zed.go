package mappings

import (
	"encoding/json"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type ZedMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	Action  string                 `yaml:"action"`
	Context string                 `yaml:"context"`
	Args    map[string]interface{} `yaml:"args,omitempty"`
}

type ZedConfigs []ZedMappingConfig

// UnmarshalYAML implements the yaml.Unmarshaler interface for ZedConfigs.
// It supports both a single mapping object and a sequence of objects.
func (z *ZedConfigs) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		var single ZedMappingConfig
		if err := node.Decode(&single); err != nil {
			return err
		}
		*z = []ZedMappingConfig{single}
	case yaml.SequenceNode:
		var slice []ZedMappingConfig
		if err := node.Decode(&slice); err != nil {
			return err
		}
		*z = slice
	default:
		return fmt.Errorf(
			"cannot unmarshal!! (line %d, col %d): expected a mapping or sequence node for zed config",
			node.Line,
			node.Column,
		)
	}
	return nil
}

func checkZedDuplicateConfig(mappings map[string]ActionMappingConfig) error {
	allowDuplicate := []string{}
	seen := make(map[struct{ Action, Context, Args string }]string)
	dups := make(map[string][]string) // key string -> list of universal action IDs
	for id, mapping := range mappings {
		if slices.Contains(allowDuplicate, id) {
			continue
		}
		for _, zconf := range mapping.Zed {
			if zconf.Action == "" {
				continue
			}
			// Serialize args to string for comparison
			var argsStr string
			if zconf.Args != nil {
				argsBytes, _ := json.Marshal(zconf.Args)
				argsStr = string(argsBytes)
			}

			key := struct{ Action, Context, Args string }{zconf.Action, zconf.Context, argsStr}
			if originalID, exists := seen[key]; exists {
				dupKey := fmt.Sprintf(`{"action":%q,"context":%q,"args":%q}`, key.Action, key.Context, key.Args)
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
	return &DuplicateActionMappingError{Editor: "zed", Duplicates: dups}
}
