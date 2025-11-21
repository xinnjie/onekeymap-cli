package validate

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	validateapi "github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// DuplicateMappingRule detects duplicate mappings (same action and keys combination).
type DuplicateMappingRule struct{}

// NewDuplicateMappingRule creates a new duplicate mapping validation rule.
func NewDuplicateMappingRule() validateapi.ValidationRule {
	return &DuplicateMappingRule{}
}

// Validate checks for duplicate mappings and adds warnings to the report.
func (r *DuplicateMappingRule) Validate(_ context.Context, validationContext *validateapi.ValidationContext) error {
	setting := validationContext.Setting
	report := validationContext.Report

	// Track seen combinations of action + formatted keys
	seen := make(map[string]struct{})

	for _, action := range setting.Actions {
		for _, b := range action.Bindings {
			key := b.String(keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			})
			composite := action.Name + "\x00" + key
			if _, exists := seen[composite]; exists {
				warning := &keymapv1.ValidationIssue{
					Issue: &keymapv1.ValidationIssue_DuplicateMapping{
						DuplicateMapping: &keymapv1.DuplicateMapping{
							Action:     action.Name,
							Keybinding: key,
							Message:    "This keymap is defined multiple times in the source configuration.",
						},
					},
				}
				report.Warnings = append(report.Warnings, warning)
			} else {
				seen[composite] = struct{}{}
			}
		}
	}

	return nil
}
