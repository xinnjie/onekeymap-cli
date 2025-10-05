package validateapi

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// UnsupportedActionRule detects actions that cannot be exported to a specific target editor.
type UnsupportedActionRule struct {
	mappingConfig *mappings.MappingConfig
	targetEditor  pluginapi.EditorType
}

// NewUnsupportedActionRule creates a new unsupported action validation rule.
func NewUnsupportedActionRule(mappingConfig *mappings.MappingConfig, targetEditor pluginapi.EditorType) ValidationRule {
	return &UnsupportedActionRule{
		mappingConfig: mappingConfig,
		targetEditor:  targetEditor,
	}
}

// Validate checks for actions that cannot be exported to the target editor.
func (r *UnsupportedActionRule) Validate(ctx context.Context, validationContext *ValidationContext) error {
	setting := validationContext.Setting
	report := validationContext.Report

	for _, ab := range setting.GetKeybindings() {
		if ab == nil {
			continue
		}

		// Check if this action has a mapping for the target editor
		actionMapping, exists := r.mappingConfig.Mappings[ab.GetName()]
		if !exists {
			// This would be caught by DanglingActionRule, skip here
			continue
		}

		// Check if the target editor is supported for this action
		var hasTargetMapping bool
		switch r.targetEditor {
		case pluginapi.EditorTypeVSCode:
			hasTargetMapping = actionMapping.VSCode != nil
		case pluginapi.EditorTypeZed:
			hasTargetMapping = actionMapping.Zed != nil
		case pluginapi.EditorTypeHelix:
			hasTargetMapping = actionMapping.Helix != nil
		case pluginapi.EditorTypeIntelliJ:
			hasTargetMapping = actionMapping.IntelliJ.Action != ""
		default:
			// Unknown editor type, skip
			continue
		}

		if !hasTargetMapping {
			for _, b := range ab.GetBindings() {
				if b == nil {
					continue
				}
				// Format the key binding for the error message
				kb := keymap.NewKeyBinding(b)
				formattedKeys, err := kb.Format(platform.PlatformMacOS, " ")
				if err != nil {
					formattedKeys = "unknown"
				}
				// Add error for unsupported action
				issue := &keymapv1.ValidationIssue{
					Issue: &keymapv1.ValidationIssue_UnsupportedAction{
						UnsupportedAction: &keymapv1.UnsupportedAction{
							Action:       ab.GetName(),
							Keybinding:   formattedKeys,
							TargetEditor: string(r.targetEditor),
						},
					},
				}
				report.Issues = append(report.Issues, issue)
			}
		}
	}

	return nil
}
