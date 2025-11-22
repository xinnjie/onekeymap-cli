package validateapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

func newAction(name string, bindings ...string) keymap.Action {
	action := keymap.Action{Name: name}
	for _, b := range bindings {
		kb, _ := keybinding.NewKeybinding(b, keybinding.ParseOption{Separator: "+"})
		action.Bindings = append(action.Bindings, kb)
	}
	return action
}

func TestValidator_Validate_EmptyKeymaps(t *testing.T) {
	validator := validateapi.NewValidator()

	setting := keymap.Keymap{
		Actions: []keymap.Action{},
	}

	opts := importerapi.ImportOptions{
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

	validator := validateapi.NewValidator(
		validate.NewKeybindConflictRule(),
		validate.NewDanglingActionRule(mappingConfig),
	)

	// Create keymaps with both conflicts and dangling actions
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("valid.action", "ctrl+c"),
			newAction("invalid.action", "ctrl+c"), // This will cause dangling action
		},
	}

	opts := importerapi.ImportOptions{
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
