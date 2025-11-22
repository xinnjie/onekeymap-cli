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

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Equal(t, "vscode", report.SourceEditor)
	assert.Equal(t, 0, report.Summary.MappingsProcessed)
	assert.Empty(t, report.Issues)
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

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)

	// Should have both keybind conflict and dangling action issues
	assert.Len(t, report.Issues, 2)

	// Check that we have both types of issues
	hasKeybindConflict := false
	hasDanglingAction := false

	for _, issue := range report.Issues {
		if issue.Type == validateapi.IssueTypeKeybindConflict {
			hasKeybindConflict = true
		}
		if issue.Type == validateapi.IssueTypeDanglingAction {
			hasDanglingAction = true
		}
	}

	assert.True(t, hasKeybindConflict, "Expected keybind conflict issue")
	assert.True(t, hasDanglingAction, "Expected dangling action issue")
}
