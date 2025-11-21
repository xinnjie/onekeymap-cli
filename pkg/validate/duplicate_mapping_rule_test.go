package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	pkgkeymap "github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

func TestDuplicateMappingRule_Validate_WithDuplicates(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	// Create keymaps with duplicate mappings
	setting := pkgkeymap.Keymap{
		Actions: []pkgkeymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.copy", "ctrl+c"), // Duplicate
			newAction("actions.edit.paste", "ctrl+v"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetWarnings(), 1)
	assert.NotNil(t, report.GetWarnings()[0].GetDuplicateMapping())

	duplicate := report.GetWarnings()[0].GetDuplicateMapping()
	assert.Equal(t, "actions.edit.copy", duplicate.GetAction())
	assert.Contains(t, duplicate.GetMessage(), "multiple times")
}

func TestDuplicateMappingRule_Validate_NoDuplicates(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	setting := pkgkeymap.Keymap{
		Actions: []pkgkeymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.paste", "ctrl+v"),
			newAction("actions.edit.cut", "ctrl+x"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetWarnings())
}

func TestDuplicateMappingRule_Validate_SameActionDifferentKeys(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewDuplicateMappingRule())

	// Same action with different keys should not be flagged as duplicate
	setting := pkgkeymap.Keymap{
		Actions: []pkgkeymap.Action{
			newAction("actions.edit.copy", "ctrl+c"),
			newAction("actions.edit.copy", "cmd+c"), // Different keys
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetWarnings())
}
