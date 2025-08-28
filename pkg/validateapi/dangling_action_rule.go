package validateapi

import (
	"context"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// DanglingActionRule checks for actions that don't exist in the action mappings.
type DanglingActionRule struct {
	mappingConfig *mappings.MappingConfig
}

// NewDanglingActionRule creates a new dangling action validation rule.
func NewDanglingActionRule(mappingConfig *mappings.MappingConfig) ValidationRule {
	return &DanglingActionRule{
		mappingConfig: mappingConfig,
	}
}

// Validate checks for dangling actions in the keymap setting.
func (r *DanglingActionRule) Validate(ctx context.Context, validationContext *ValidationContext) error {
	if validationContext.Setting == nil || len(validationContext.Setting.Keybindings) == 0 {
		return nil
	}

	for _, kb := range validationContext.Setting.Keybindings {
		if kb == nil {
			continue
		}

		// Check if the action exists in the mapping configuration
		if _, exists := r.mappingConfig.Mappings[kb.Id]; !exists {
			// Create a dangling action issue
			danglingAction := &keymapv1.DanglingAction{
				Action:     kb.Id,
				Keybinding: kb.Id, // Use action as placeholder for keybinding
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
