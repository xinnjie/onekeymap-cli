package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

func TestDuplicateMappingRule_Validate_WithDuplicates(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	// Create keymaps with duplicate mappings
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.copy", "ctrl+c"), // Duplicate
			newAction("actions.edit.paste", "ctrl+v"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Warnings, 1)
	assert.Equal(t, validateapi.IssueTypeDuplicateMapping, report.Warnings[0].Type)

	duplicate, ok := report.Warnings[0].Details.(validateapi.DuplicateMapping)
	require.True(t, ok)
	assert.Equal(t, "actions.edit.copy", duplicate.Action)
}

func TestDuplicateMappingRule_Validate_NoDuplicates(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.paste", "ctrl+v"),
			newAction("actions.edit.cut", "ctrl+x"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Empty(t, report.Warnings)
}

func TestDuplicateMappingRule_Validate_SameActionDifferentKeys(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	// Same action with different keys should not be flagged as duplicate
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.copy", "cmd+c"), // Different keys
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Empty(t, report.Warnings)
}
