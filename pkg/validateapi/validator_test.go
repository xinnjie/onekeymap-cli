package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestValidator_Validate_EmptyKeymaps(t *testing.T) {
	validator := NewValidator()

	setting := &keymapv1.Keymap{
		Keybindings: []*keymapv1.ActionBinding{},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Equal(t, "vscode", report.GetSourceEditor())
	assert.Equal(t, int32(0), report.GetSummary().GetMappingsProcessed())
	assert.Empty(t, report.GetIssues())
}

func TestValidator_Validate_ChainOfRules(t *testing.T) {
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"valid.action": {
				ID: "valid.action",
			},
		},
	}

	validator := NewValidator(NewKeybindConflictRule(), NewDanglingActionRule(mappingConfig))

	// Create keymaps with both conflicts and dangling actions
	setting := &keymapv1.Keymap{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("valid.action", "ctrl+c"),
			keymap.NewActioinBinding("invalid.action", "ctrl+c"), // This will cause dangling action
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)

	// Should have both keybind conflict and dangling action issues
	assert.Len(t, report.GetIssues(), 2)

	// Check that we have both types of issues
	hasKeybindConflict := false
	hasDanglingAction := false

	for _, issue := range report.GetIssues() {
		if issue.GetKeybindConflict() != nil {
			hasKeybindConflict = true
		}
		if issue.GetDanglingAction() != nil {
			hasDanglingAction = true
		}
	}

	assert.True(t, hasKeybindConflict, "Expected keybind conflict issue")
	assert.True(t, hasDanglingAction, "Expected dangling action issue")
}
