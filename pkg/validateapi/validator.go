package validateapi

import (
	"context"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// Validator is a chain of responsibility container for validation rules.
type Validator struct {
	rules []ValidationRule
}

// NewValidator creates a new validator with no rules.
func NewValidator(rules ...ValidationRule) *Validator {
	return &Validator{
		rules: rules,
	}
}

// Validate executes all validation rules in the chain.
func (v *Validator) Validate(
	ctx context.Context,
	setting *keymapv1.KeymapSetting,
	opts importapi.ImportOptions,
) (*keymapv1.ValidationReport, error) {
	report := &keymapv1.ValidationReport{
		SourceEditor: string(opts.EditorType),
		Summary: &keymapv1.Summary{
			MappingsProcessed: int32(len(setting.GetKeybindings())),
			MappingsSucceeded: 0, // Will be updated by rules
		},
		Issues:   make([]*keymapv1.ValidationIssue, 0),
		Warnings: make([]*keymapv1.ValidationIssue, 0),
	}

	validationContext := &ValidationContext{
		Setting: setting,
		Report:  report,
		Options: opts,
	}

	// Execute all rules in the chain
	for _, rule := range v.rules {
		if err := rule.Validate(ctx, validationContext); err != nil {
			return nil, err
		}
	}

	// Update succeeded count (total - issues)
	issueCount := len(report.GetIssues())
	report.Summary.MappingsSucceeded = report.GetSummary().GetMappingsProcessed() - int32(issueCount)

	return report, nil
}
