package validateapi

import (
	"context"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// DuplicateMappingRule detects duplicate mappings (same action and keys combination).
type DuplicateMappingRule struct{}

// NewDuplicateMappingRule creates a new duplicate mapping validation rule.
func NewDuplicateMappingRule() ValidationRule {
	return &DuplicateMappingRule{}
}

// Validate checks for duplicate mappings and adds warnings to the report.
func (r *DuplicateMappingRule) Validate(ctx context.Context, validationContext *ValidationContext) error {
	setting := validationContext.Setting
	report := validationContext.Report

	// Track seen combinations of action + formatted keys
	seen := make(map[string]*keymapv1.KeyBinding)

	for _, binding := range setting.GetKeybindings() {
		if binding == nil {
			continue
		}

		// Create KeyBinding with action to use String() method
		kb := keymap.NewKeyBinding(binding)

		// Use the unified String() method for consistent key generation
		key := kb.String()

		if _, exists := seen[key]; exists {
			// Found a duplicate - add warning
			warning := &keymapv1.ValidationIssue{
				Issue: &keymapv1.ValidationIssue_DuplicateMapping{
					DuplicateMapping: &keymapv1.DuplicateMapping{
						Action:     binding.Action,
						Keybinding: kb.String(),
						Message:    "This keymap is defined multiple times in the source configuration.",
					},
				},
			}
			report.Warnings = append(report.Warnings, warning)
		} else {
			// First time seeing this combination
			seen[key] = binding
		}
	}

	return nil
}
