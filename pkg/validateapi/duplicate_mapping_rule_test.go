package validateapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func TestDuplicateMappingRule_Validate_WithDuplicates(t *testing.T) {
	validator := validateapi.NewValidator(validateapi.NewDuplicateMappingRule())

	// Create keymaps with duplicate mappings
	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"), // Duplicate
			keymap.NewActioinBinding("actions.edit.paste", "ctrl+v"),
		},
	}

	opts := importapi.ImportOptions{
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
	validator := validateapi.NewValidator(validateapi.NewDuplicateMappingRule())

	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.paste", "ctrl+v"),
			keymap.NewActioinBinding("actions.edit.cut", "ctrl+x"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetWarnings())
}

func TestDuplicateMappingRule_Validate_SameActionDifferentKeys(t *testing.T) {
	validator := validateapi.NewValidator(validateapi.NewDuplicateMappingRule())

	// Same action with different keys should not be flagged as duplicate
	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.copy", "cmd+c"), // Different keys
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetWarnings())
}
