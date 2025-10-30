package validateapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func TestUnsupportedActionRule_Validate_WithUnsupportedAction(t *testing.T) {
	// Create a test mapping config with limited editor support
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.supported": {
				ID: "actions.supported",
				VSCode: mappings.VscodeConfigs{{
					Command: "supported.command",
				}},
				Zed: mappings.ZedConfigs{{
					Action: "supported::action",
				}},
			},
			"actions.vscode.only": {
				ID: "actions.vscode.only",
				VSCode: mappings.VscodeConfigs{{
					Command: "vscode.only.command",
				}},
				// No Zed mapping
			},
		},
	}

	validator := validateapi.NewValidator(validateapi.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed))

	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.supported", "ctrl+s"),
			keymap.NewActioinBinding("actions.vscode.only", "ctrl+v"), // Unsupported in Zed
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "zed",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
	assert.NotNil(t, report.GetIssues()[0].GetUnsupportedAction())

	unsupported := report.GetIssues()[0].GetUnsupportedAction()
	assert.Equal(t, "actions.vscode.only", unsupported.GetAction())
	assert.Equal(t, "zed", unsupported.GetTargetEditor())
}

func TestUnsupportedActionRule_Validate_AllSupported(t *testing.T) {
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.universal": {
				ID: "actions.universal",
				VSCode: mappings.VscodeConfigs{{
					Command: "universal.command",
				}},
				Zed: mappings.ZedConfigs{{
					Action: "universal::action",
				}},
			},
		},
	}

	validator := validateapi.NewValidator(validateapi.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed))

	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.universal", "ctrl+u"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "zed",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetIssues())
}

func TestUnsupportedActionRule_Validate_DifferentEditors(t *testing.T) {
	mappingConfig := &mappings.MappingConfig{
		Mappings: map[string]mappings.ActionMappingConfig{
			"actions.test": {
				ID: "actions.test",
				VSCode: mappings.VscodeConfigs{{
					Command: "test.command",
				}},
				// No Zed mapping
			},
		},
	}

	// Test with VSCode target - should pass
	validatorVSCode := validateapi.NewValidator(
		validateapi.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeVSCode),
	)
	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.test", "ctrl+t"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validatorVSCode.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetIssues())

	// Test with Zed target - should fail
	validatorZed := validateapi.NewValidator(
		validateapi.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed),
	)
	opts.EditorType = "zed"

	report, err = validatorZed.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetIssues(), 1)
}
