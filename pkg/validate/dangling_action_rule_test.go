package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	pkgkeymap "github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

// newAction creates a test action with bindings
func newAction(name string, bindings ...string) pkgkeymap.Action {
	action := pkgkeymap.Action{Name: name}
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

	setting := pkgkeymap.Keymap{
		Actions: []pkgkeymap.Action{
			newAction("valid.action", "a"),
			newAction("invalid.action", "b"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
	assert.NotNil(t, report.GetIssues()[0].GetDanglingAction())

	danglingAction := report.GetIssues()[0].GetDanglingAction()
	assert.Equal(t, "invalid.action", danglingAction.GetAction())
}
