package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestValidator_Validate_WithKeybindConflict(t *testing.T) {
	validator := NewValidator(NewKeybindConflictRule())

	// Create keymaps with conflicting keybindings
	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			keymap.NewBinding("action1", "ctrl+c"),
			keymap.NewBinding("action2", "ctrl+c"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.NotNil(t, report.Issues[0].GetKeybindConflict())

	conflict := report.Issues[0].GetKeybindConflict()
	assert.Len(t, conflict.Actions, 2)
}
