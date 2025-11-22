package validateapi

import (
	"context"
	"errors"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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
	setting keymap.Keymap,
	editorType pluginapi.EditorType,
) (*ValidationReport, error) {
	processed := len(setting.Actions)
	if processed < 0 {
		return nil, errors.New("keybindings count out of range")
	}
	report := &ValidationReport{
		SourceEditor: string(editorType),
		Summary: Summary{
			MappingsProcessed: processed,
			MappingsSucceeded: 0, // Will be updated by rules
		},
		Issues:   make([]ValidationIssue, 0),
		Warnings: make([]ValidationIssue, 0),
	}

	validationContext := &ValidationContext{
		Setting:    setting,
		Report:     report,
		EditorType: editorType,
	}

	// Execute all rules in the chain
	for _, rule := range v.rules {
		if err := rule.Validate(ctx, validationContext); err != nil {
			return nil, err
		}
	}

	// Update succeeded count (total - issues)
	success := max(report.Summary.MappingsProcessed-len(report.Issues), 0)
	if success < 0 {
		return nil, errors.New("keybindings count out of range")
	}
	report.Summary.MappingsSucceeded = success

	return report, nil
}
