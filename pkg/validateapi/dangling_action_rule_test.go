package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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

	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("valid.action", "a"),
			keymap.NewActioinBinding("invalid.action", "b"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.NotNil(t, report.Issues[0].GetDanglingAction())

	danglingAction := report.Issues[0].GetDanglingAction()
	assert.Equal(t, "invalid.action", danglingAction.Action)
}
