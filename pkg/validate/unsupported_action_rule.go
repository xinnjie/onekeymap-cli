package validate

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// UnsupportedActionRule detects actions that cannot be exported to a specific target editor.
type UnsupportedActionRule struct {
	mappingConfig *mappings.MappingConfig
	targetEditor  pluginapi.EditorType
}

// NewUnsupportedActionRule creates a new unsupported action validation rule.
func NewUnsupportedActionRule(
	mappingConfig *mappings.MappingConfig,
	targetEditor pluginapi.EditorType,
) validateapi.ValidationRule {
	return &UnsupportedActionRule{
		mappingConfig: mappingConfig,
		targetEditor:  targetEditor,
	}
}

// Validate checks for actions that cannot be exported to the target editor.
func (r *UnsupportedActionRule) Validate(_ context.Context, validationContext *validateapi.ValidationContext) error {
	setting := validationContext.Setting
	report := validationContext.Report

	for _, action := range setting.Actions {
		// Check if this action has a mapping for the target editor
		actionMapping, exists := r.mappingConfig.Mappings[action.Name]
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
			for _, b := range action.Bindings {
				// Format the key binding for the error message
				formattedKeys := b.String(keybinding.FormatOption{
					Platform:  platform.PlatformMacOS,
					Separator: " ",
				})
				// Add error for unsupported action
				issue := &keymapv1.ValidationIssue{
					Issue: &keymapv1.ValidationIssue_UnsupportedAction{
						UnsupportedAction: &keymapv1.UnsupportedAction{
							Action:       action.Name,
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
