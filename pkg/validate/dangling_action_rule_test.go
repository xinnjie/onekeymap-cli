package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

// newAction creates a test action with bindings
func newAction(name string, bindings ...string) keymap.Action {
	action := keymap.Action{Name: name}
	for _, b := range bindings {
		kb, _ := keybinding.NewKeybinding(b, keybinding.ParseOption{Separator: "+"})
		action.Bindings = append(action.Bindings, kb)
	}
	return action
}

func TestValidator_Validate_WithDanglingAction(t *testing.T) {
	// Create a test mapping config
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"valid.action": {
				ID: "valid.action",
			},
		},
	}

	validator := validateapi.NewValidator(validate.NewDanglingActionRule(mappingConfig))

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("valid.action", "a"),
			newAction("invalid.action", "b"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.Equal(t, validateapi.IssueTypeDanglingAction, report.Issues[0].Type)

	danglingAction, ok := report.Issues[0].Details.(validateapi.DanglingAction)
	require.True(t, ok)
	assert.Equal(t, "invalid.action", danglingAction.Action)
}
