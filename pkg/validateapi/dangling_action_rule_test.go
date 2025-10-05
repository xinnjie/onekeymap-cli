package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func TestValidator_Validate_WithDanglingAction(t *testing.T) {
	// Create a test mapping config
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"valid.action": {
				ID: "valid.action",
			},
		},
	}

	validator := NewValidator(NewDanglingActionRule(mappingConfig))

	setting := &keymapv1.Keymap{
		Keybindings: []*keymapv1.Action{
			keymap.NewActioinBinding("valid.action", "a"),
			keymap.NewActioinBinding("invalid.action", "b"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
	assert.NotNil(t, report.GetIssues()[0].GetDanglingAction())

	danglingAction := report.GetIssues()[0].GetDanglingAction()
	assert.Equal(t, "invalid.action", danglingAction.GetAction())
}
