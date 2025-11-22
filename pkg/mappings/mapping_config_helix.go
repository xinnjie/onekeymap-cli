package mappings

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type HelixMappingConfig struct {
	EditorActionMapping `yaml:",inline"`

	Command string `yaml:"command"`
	Mode    string `yaml:"mode"`
}

type HelixConfig []HelixMappingConfig

// UnmarshalYAML implements the yaml.Unmarshaler interface for HelixConfigs.
// It supports both a single mapping object and a sequence of objects.
func (h *HelixConfig) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		var single HelixMappingConfig
		if err := node.Decode(&single); err != nil {
			return err
		}
		*h = []HelixMappingConfig{single}
	case yaml.SequenceNode:
		var slice []HelixMappingConfig
		if err := node.Decode(&slice); err != nil {
			return err
		}
		*h = slice
	default:
		return fmt.Errorf(
			"cannot unmarshal! (line %d, col %d): expected a mapping or sequence node for helix config",
			node.Line,
			node.Column,
		)
	}
	return nil
}
