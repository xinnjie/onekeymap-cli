package validateapi

import (
	"context"
	"errors"
	"math"

	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
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
	setting *keymapv1.Keymap,
	opts importerapi.ImportOptions,
) (*keymapv1.ValidationReport, error) {
	processed := len(setting.GetActions())
	if processed > math.MaxInt32 || processed < 0 {
		return nil, errors.New("keybindings count out of range")
	}
	report := &keymapv1.ValidationReport{
		SourceEditor: string(opts.EditorType),
		Summary: &keymapv1.Summary{
			MappingsProcessed: int32(processed),
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
	success := max(int(report.GetSummary().GetMappingsProcessed())-len(report.GetIssues()), 0)
	if success > math.MaxInt32 || success < 0 {
		return nil, errors.New("keybindings count out of range")
	}
	report.Summary.MappingsSucceeded = int32(success)

	return report, nil
}
