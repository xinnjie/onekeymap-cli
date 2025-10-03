package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestValidator_Validate_WithKeybindConflict(t *testing.T) {
	validator := NewValidator(NewKeybindConflictRule())

	// Create keymaps with conflicting keybindings
	setting := &keymapv1.Keymap{
		Keybindings: []*keymapv1.Action{
			keymap.NewActioinBinding("action1", "ctrl+c"),
			keymap.NewActioinBinding("action2", "ctrl+c"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
	assert.NotNil(t, report.GetIssues()[0].GetKeybindConflict())

	conflict := report.GetIssues()[0].GetKeybindConflict()
	assert.Len(t, conflict.GetActions(), 2)
}
