package validateapi

import (
	"context"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
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
	seen := make(map[string]struct{})

	for _, ab := range setting.GetKeybindings() {
		if ab == nil {
			continue
		}
		for _, b := range ab.GetBindings() {
			if b == nil {
				continue
			}
			kb := keymap.NewKeyBinding(b)
			key := keymap.MustFormatKeyBinding(kb, platform.PlatformMacOS)
			composite := ab.GetName() + "\x00" + key
			if _, exists := seen[composite]; exists {
				warning := &keymapv1.ValidationIssue{
					Issue: &keymapv1.ValidationIssue_DuplicateMapping{
						DuplicateMapping: &keymapv1.DuplicateMapping{
							Action:     ab.GetName(),
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
