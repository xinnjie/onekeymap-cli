package mappings

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	actionmappings "github.com/xinnjie/onekeymap-cli/config/action_mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
	// Featured means that this action is implemented within one or few editors and is not portable
	Featured bool `yaml:"featured"`
	// FeaturedReason: Why this action is not common across editors, or recommand users to use another portable actioin of similar utility.
	FeaturedReason string                `yaml:"featuredReason"`
	Category       string                `yaml:"category"`
	VSCode         VscodeConfigs         `yaml:"vscode"`
	Windsurf       VscodeConfigs         `yaml:"windsurf"`
	Cursor         VscodeConfigs         `yaml:"cursor"`
	Zed            ZedConfigs            `yaml:"zed"`
	IntelliJ       IntelliJMappingConfig `yaml:"intellij"`
	Vim            VimMappingConfig      `yaml:"vim"`
	Helix          HelixConfig           `yaml:"helix"`
	Xcode          XcodeConfigs          `yaml:"xcode"`
	// Children is a list of child action IDs for UI hierarchical grouping only.
	// This field has no effect on export/import logic.
	Children []string `yaml:"children,omitempty"`
	// Fallbacks is a list of action IDs used for export fallback.
	// When this action is not supported by a target editor, the system will try each fallback in order.
	Fallbacks []string `yaml:"fallbacks,omitempty"`
}

const explicitlyNotSupported = "__explicitly_not_supported__"

// IsSupported checks if the action is supported by the given editor type.
// Returns (supported, note).
func (am *ActionMappingConfig) IsSupported(editorType pluginapi.EditorType) (bool, string) {
	switch editorType {
	case pluginapi.EditorTypeVSCode:
		return am.isSupportedVSCode()
	case pluginapi.EditorTypeWindsurf, pluginapi.EditorTypeWindsurfNext, pluginapi.EditorTypeCursor:
		return am.isSupportedVSCodeVariant(editorType)
	case pluginapi.EditorTypeIntelliJ:
		return am.isSupportedIntelliJ()
	case pluginapi.EditorTypeZed:
		return am.isSupportedZed()
	case pluginapi.EditorTypeVim:
		return am.isSupportedVim()
	case pluginapi.EditorTypeHelix:
		return am.isSupportedHelix()
	case pluginapi.EditorTypeXcode:
		return am.isSupportedXcode()
	default:
		return false, ""
	}
}

func (am *ActionMappingConfig) isSupportedVSCode() (bool, string) {
	if len(am.VSCode) == 0 {
		return false, ""
	}
	var notes []string
	for _, vc := range am.VSCode {
		if vc.NotSupported {
			if vc.Note == "" {
				return false, explicitlyNotSupported
			}
			return false, vc.Note
		}
		if vc.Note != "" {
			notes = append(notes, vc.Note)
		}
	}
	hasMapping := slices.ContainsFunc(am.VSCode, func(vc VscodeMappingConfig) bool {
		return vc.Command != ""
	})
	return hasMapping, strings.Join(notes, ", ")
}

func (am *ActionMappingConfig) isSupportedIntelliJ() (bool, string) {
	if am.IntelliJ == (IntelliJMappingConfig{}) {
		return false, ""
	}
	if am.IntelliJ.NotSupported {
		if am.IntelliJ.Note == "" {
			return false, explicitlyNotSupported
		}
		return false, am.IntelliJ.Note
	}
	return am.IntelliJ.Action != "", am.IntelliJ.Note
}

func (am *ActionMappingConfig) isSupportedZed() (bool, string) {
	if len(am.Zed) == 0 {
		return false, ""
	}
	var notes []string
	for _, zc := range am.Zed {
		if zc.NotSupported {
			if zc.Note == "" {
				return false, explicitlyNotSupported
			}
			return false, zc.Note
		}
		if zc.Note != "" {
			notes = append(notes, zc.Note)
		}
	}
	hasMapping := slices.ContainsFunc(am.Zed, func(zc ZedMappingConfig) bool {
		return zc.Action != ""
	})
	return hasMapping, strings.Join(notes, ", ")
}

func (am *ActionMappingConfig) isSupportedVim() (bool, string) {
	if am.Vim == (VimMappingConfig{}) {
		return false, ""
	}
	if am.Vim.NotSupported {
		if am.Vim.Note == "" {
			return false, explicitlyNotSupported
		}
		return false, am.Vim.Note
	}
	return am.Vim.Command != "", am.Vim.Note
}

func (am *ActionMappingConfig) isSupportedHelix() (bool, string) {
	if len(am.Helix) == 0 {
		return false, ""
	}
	var notes []string
	for _, hc := range am.Helix {
		if hc.NotSupported {
			if hc.Note == "" {
				return false, explicitlyNotSupported
			}
			return false, hc.Note
		}
		if hc.Note != "" {
			notes = append(notes, hc.Note)
		}
	}
	hasMapping := slices.ContainsFunc(am.Helix, func(hc HelixMappingConfig) bool {
		return hc.Command != ""
	})
	return hasMapping, strings.Join(notes, ", ")
}

func (am *ActionMappingConfig) isSupportedXcode() (bool, string) {
	if len(am.Xcode) == 0 {
		return false, ""
	}
	var notes []string
	for _, xc := range am.Xcode {
		if xc.NotSupported {
			if xc.Note == "" {
				return false, explicitlyNotSupported
			}
			return false, xc.Note
		}
		if xc.Note != "" {
			notes = append(notes, xc.Note)
		}
	}
	hasMapping := slices.ContainsFunc(am.Xcode, func(xc XcodeMappingConfig) bool {
		if xc.MenuAction.Action != "" {
			return true
		}
		return len(xc.TextAction.TextAction.Items) > 0
	})
	return hasMapping, strings.Join(notes, ", ")
}

// EditorActionMapping provides extra flags for editor-specific configurations.
type EditorActionMapping struct {
	// when true, this config is only used for export, otherwise it is used for both import and export
	DisableImport bool `yaml:"disableImport,omitempty"`

	// explicitly set to true if this config is not supported by the editor
	NotSupported bool `yaml:"notSupported,omitempty"`
	// reason why this config is not supported by the editor
	Note string `yaml:"note,omitempty"`
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

	if err := checkEditorConfigs(mergedMappings); err != nil {
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

// GetExportAction returns the action config to use for export.
// If the action is supported by the editor, it returns the action itself.
// If not supported, it falls back to the first supported fallback action.
// Returns nil if no suitable action is found.
// The second return value is true if a fallback was used.
func (mc *MappingConfig) GetExportAction(
	actionID string,
	editorType pluginapi.EditorType,
) (*ActionMappingConfig, bool) {
	mapping := mc.Get(actionID)
	if mapping == nil {
		return nil, false
	}

	// Check if the action itself is supported
	if supported, _ := mapping.IsSupported(editorType); supported {
		return mapping, false
	}

	// Fallback to fallbacks if present
	for _, fallbackID := range mapping.Fallbacks {
		fallbackMapping := mc.Get(fallbackID)
		if fallbackMapping == nil {
			continue
		}
		if supported, _ := fallbackMapping.IsSupported(editorType); supported {
			return fallbackMapping, true
		}
	}

	return nil, false
}

func checkEditorConfigs(mappings map[string]ActionMappingConfig) error {
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
	if err := checkXcodeDuplicateConfig(mappings); err != nil {
		return err
	}
	if err := checkXcodeTextActionFormat(mappings); err != nil {
		return err
	}
	if err := checkXcodeImportConstraints(mappings); err != nil {
		return err
	}
	return nil
}
