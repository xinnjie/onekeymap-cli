package mappings

import (
	"fmt"
	"io"

	actionmappings "github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings/action_mappings"
	"gopkg.in/yaml.v3"
)

// MappingConfig holds the final, merged mapping config, indexed by action ID.
// It serves as a `Anemic Domain Model` because I want the editor specific config query to be implemented in plugin
type MappingConfig struct {
	// Mappings is a map where the key is the one_keymap_id (e.g., "actions.editor.copy")
	// and the value is the detailed mapping information for that action.
	Mappings map[string]ActionMappingConfig
}

func NewMappingConfig() (*MappingConfig, error) {
	reader, err := actionmappings.ReadActionMapping()
	if err != nil {
		return nil, err
	}
	config, err := load(reader)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ActionMappingConfig holds the complete mapping information for a single action ID.
type ActionMappingConfig struct {
	ID          string                `yaml:"id"`
	Description string                `yaml:"description"`
	Name        string                `yaml:"name"`
	Category    string                `yaml:"category"`
	VSCode      VscodeConfigs         `yaml:"vscode"`
	Zed         ZedConfigs            `yaml:"zed"`
	IntelliJ    IntelliJMappingConfig `yaml:"intellij"`
	Vim         VimMappingConfig      `yaml:"vim"`
	Helix       HelixConfig           `yaml:"helix"`
}

// every editor action mapping config should hold EditorActionMapping to provide extra support
type EditorActionMapping struct {
	// when true, this config is used for import, otherwise it is only used for export
	ForImport bool `yaml:"forImport,omitempty"`

	// explicitly set to true if this config is not supported by the editor
	NotSupported bool `yaml:"notSupported,omitempty"`
	// reason why this config is not supported by the editor
	NotSupportedReason string `yaml:"notSupportedReason,omitempty"`
}

// configFormat is a struct that matches the structure of each YAML file.
type configFormat struct {
	Mappings []ActionMappingConfig `yaml:"mappings"`
}

func NewTestMappingConfig() (*MappingConfig, error) {
	reader, err := actionmappings.ReadTestActionMapping()
	if err != nil {
		return nil, err
	}
	config, err := load(reader)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// load parses a multi-document YAML stream from an io.Reader, and merges the documents
// into a single, validated MappingData structure.
func load(reader io.Reader) (*MappingConfig, error) {
	decoder := yaml.NewDecoder(reader)
	mergedMappings := make(map[string]ActionMappingConfig)

	for {
		var fileContent configFormat
		if err := decoder.Decode(&fileContent); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML stream: %w", err)
		}

		for _, mapping := range fileContent.Mappings {
			if _, exists := mergedMappings[mapping.ID]; exists {
				return nil, fmt.Errorf("duplicate action ID '%s' found in stream", mapping.ID)
			}
			mergedMappings[mapping.ID] = mapping
		}
	}

	if err := checkDuplicateEditorConfigs(mergedMappings); err != nil {
		return nil, err
	}

	return &MappingConfig{Mappings: mergedMappings}, nil
}

// FindByUniversalAction searches for a mapping by universal action ID.
func (mc *MappingConfig) FindByUniversalAction(action string) *ActionMappingConfig {
	if mapping, exists := mc.Mappings[action]; exists {
		return &mapping
	}
	return nil
}

// IsActionMapped checks if a universal action ID is defined in the mapping configuration.
func (mc *MappingConfig) IsActionMapped(action string) bool {
	_, exists := mc.Mappings[action]
	return exists
}

func checkDuplicateEditorConfigs(mappings map[string]ActionMappingConfig) error {
	if err := checkVscodeDuplicateConfig(mappings); err != nil {
		return err
	}
	if err := checkIntellijDuplicateConfig(mappings); err != nil {
		return err
	}
	if err := checkVimDuplicateConfig(mappings); err != nil {
		return err
	}
	if err := checkZedDuplicateConfig(mappings); err != nil {
		return err
	}
	return nil
}
