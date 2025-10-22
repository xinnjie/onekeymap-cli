package mappings

import (
	"errors"
	"fmt"
	"io"
	"slices"

	actionmappings "github.com/xinnjie/onekeymap-cli/config/action_mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	"gopkg.in/yaml.v3"
)

// MappingConfig holds the final, merged mapping config, indexed by action ID.
// It serves as a `Anemic Domain Model` because I want the editor specific config query to be implemented in plugin.
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

// IsSupported checks if the action is supported by the given editor type.
// Returns (supported, notSupportedReason).
func (am *ActionMappingConfig) IsSupported(editorType pluginapi.EditorType) (bool, string) {
	switch editorType {
	case pluginapi.EditorTypeVSCode:
		for _, vc := range am.VSCode {
			if vc.NotSupported {
				return false, vc.NotSupportedReason
			}
		}
		hasMapping := slices.ContainsFunc(am.VSCode, func(vc VscodeMappingConfig) bool {
			return vc.Command != ""
		})
		return hasMapping, ""
	case pluginapi.EditorTypeIntelliJ:
		if am.IntelliJ.NotSupported {
			return false, am.IntelliJ.NotSupportedReason
		}
		return am.IntelliJ.Action != "", ""
	case pluginapi.EditorTypeZed:
		for _, zc := range am.Zed {
			if zc.NotSupported {
				return false, zc.NotSupportedReason
			}
		}
		hasMapping := slices.ContainsFunc(am.Zed, func(zc ZedMappingConfig) bool {
			return zc.Action != ""
		})
		return hasMapping, ""
	case pluginapi.EditorTypeVim:
		if am.Vim.NotSupported {
			return false, am.Vim.NotSupportedReason
		}
		return am.Vim.Command != "", ""
	case pluginapi.EditorTypeHelix:
		for _, hc := range am.Helix {
			if hc.NotSupported {
				return false, hc.NotSupportedReason
			}
		}
		hasMapping := slices.ContainsFunc(am.Helix, func(hc HelixMappingConfig) bool {
			return hc.Command != ""
		})
		return hasMapping, ""
	default:
		return false, ""
	}
}

// EditorActionMapping provides extra flags for editor-specific configurations.
type EditorActionMapping struct {
	// when true, this config is only used for export, otherwise it is used for both import and export
	DisableImport bool `yaml:"disableImport,omitempty"`

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
			if errors.Is(err, io.EOF) {
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

// Get searches for a mapping by universal action ID.
func (mc *MappingConfig) Get(actionID string) *ActionMappingConfig {
	if mapping, exists := mc.Mappings[actionID]; exists {
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
