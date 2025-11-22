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

func TestValidator_Validate_WithKeybindConflict(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewKeybindConflictRule())

	// Create keymaps with conflicting keybindings
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("action1", "ctrl+c"),
			newAction("action2", "ctrl+c"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.Equal(t, validateapi.IssueTypeKeybindConflict, report.Issues[0].Type)

	conflict, ok := report.Issues[0].Details.(validateapi.KeybindConflict)
	require.True(t, ok)
	assert.Len(t, conflict.Actions, 2)
}
