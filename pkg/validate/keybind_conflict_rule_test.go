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

func TestValidator_Validate_WithKeybindConflict(t *testing.T) {
	validator := validateapi.NewValidator(validate.NewKeybindConflictRule())

	// Create keymaps with conflicting keybindings
	setting := pkgkeymap.Keymap{
		Actions: []pkgkeymap.Action{
			newAction("action1", "ctrl+c"),
			newAction("action2", "ctrl+c"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
	assert.NotNil(t, report.GetIssues()[0].GetKeybindConflict())

	conflict := report.GetIssues()[0].GetKeybindConflict()
	assert.Len(t, conflict.GetActions(), 2)
}
