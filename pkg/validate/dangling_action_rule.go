package validate

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// DanglingActionRule checks for actions that don't exist in the action mappings.
type DanglingActionRule struct {
	mappingConfig *mappings.MappingConfig
}

// NewDanglingActionRule creates a new dangling action validation rule.
func NewDanglingActionRule(mappingConfig *mappings.MappingConfig) validateapi.ValidationRule {
	return &DanglingActionRule{
		mappingConfig: mappingConfig,
	}
}

// Validate checks for dangling actions in the keymap setting.
func (r *DanglingActionRule) Validate(_ context.Context, validationContext *validateapi.ValidationContext) error {
	if len(validationContext.Setting.Actions) == 0 {
		return nil
	}

	for _, action := range validationContext.Setting.Actions {
		// Check if the action exists in the mapping configuration
		if _, exists := r.mappingConfig.Mappings[action.Name]; !exists {
			// Create a dangling action issue
			danglingAction := &keymapv1.DanglingAction{
				Action:     action.Name,
				Keybinding: action.Name, // Use action as placeholder for keybinding
				Suggestion: "Check if the action ID is correct or if it needs to be added to action mappings",
			}

			issue := &keymapv1.ValidationIssue{
				Issue: &keymapv1.ValidationIssue_DanglingAction{
					DanglingAction: danglingAction,
				},
			}

			validationContext.Report.Issues = append(validationContext.Report.Issues, issue)
		}
	}

	return nil
}
