package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	mappings2 "github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

func TestUnsupportedActionRule_Validate_WithUnsupportedAction(t *testing.T) {
	// Create a test mapping config with limited editor support
	mappingConfig := &mappings2.MappingConfig{
		Mappings: map[string]mappings2.ActionMappingConfig{
			"actions.supported": {
				ID: "actions.supported",
				VSCode: mappings2.VscodeConfigs{{
					Command: "supported.command",
				}},
				Zed: mappings2.ZedConfigs{{
					Action: "supported::action",
				}},
			},
			"actions.vscode.only": {
				ID: "actions.vscode.only",
				VSCode: mappings2.VscodeConfigs{{
					Command: "vscode.only.command",
				}},
				// No Zed mapping
			},
		},
	}

	validator := validateapi.NewValidator(validate.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed))

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.supported", "ctrl+s"),
			newAction("actions.vscode.only", "ctrl+v"), // Unsupported in Zed
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "zed",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Issues, 1)
	assert.Equal(t, validateapi.IssueTypeUnsupportedAction, report.Issues[0].Type)

	unsupported, ok := report.Issues[0].Details.(validateapi.UnsupportedAction)
	require.True(t, ok)
	assert.Equal(t, "actions.vscode.only", unsupported.Action)
	assert.Equal(t, "zed", unsupported.TargetEditor)
}

func TestUnsupportedActionRule_Validate_AllSupported(t *testing.T) {
	mappingConfig := &mappings2.MappingConfig{
		Mappings: map[string]mappings2.ActionMappingConfig{
			"actions.universal": {
				ID: "actions.universal",
				VSCode: mappings2.VscodeConfigs{{
					Command: "universal.command",
				}},
				Zed: mappings2.ZedConfigs{{
					Action: "universal::action",
				}},
			},
		},
	}

	validator := validateapi.NewValidator(validate.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed))

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.universal", "ctrl+u"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "zed",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Empty(t, report.Issues)
}

func TestUnsupportedActionRule_Validate_DifferentEditors(t *testing.T) {
	mappingConfig := &mappings2.MappingConfig{
		Mappings: map[string]mappings2.ActionMappingConfig{
			"actions.test": {
				ID: "actions.test",
				VSCode: mappings2.VscodeConfigs{{
					Command: "test.command",
				}},
				// No Zed mapping
			},
		},
	}

	// Test with VSCode target - should pass
	validatorVSCode := validateapi.NewValidator(
		validate.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeVSCode),
	)
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.test", "ctrl+t"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validatorVSCode.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Empty(t, report.Issues)

	// Test with Zed target - should fail
	validatorZed := validateapi.NewValidator(
		validate.NewUnsupportedActionRule(mappingConfig, pluginapi.EditorTypeZed),
	)
	opts.EditorType = "zed"

	report, err = validatorZed.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Issues, 1)
}
