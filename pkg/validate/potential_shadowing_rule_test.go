package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func TestPotentialShadowingRule_Validate_WithCriticalKeybinding(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	// Create keymaps with critical system shortcuts
	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.format.document", "cmd+q"), // Critical: quit on macOS
			keymap.NewActioinBinding("actions.edit.copy", "cmd+c"),       // Critical: copy on macOS
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetWarnings(), 2)

	// Check first warning (cmd+q)
	warning1 := report.GetWarnings()[0].GetPotentialShadowing()
	assert.NotNil(t, warning1)
	assert.Equal(t, "actions.format.document", warning1.GetAction())
	assert.Contains(t, warning1.GetMessage(), "quitting applications on macOS")

	// Check second warning (cmd+c)
	warning2 := report.GetWarnings()[1].GetPotentialShadowing()
	assert.NotNil(t, warning2)
	assert.Equal(t, "actions.edit.copy", warning2.GetAction())
	assert.Contains(t, warning2.GetMessage(), "copy on macOS")
}

func TestPotentialShadowingRule_Validate_NoCriticalKeybindings(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.custom.action", "ctrl+shift+f12"),
			keymap.NewActioinBinding("actions.another.action", "alt+shift+g"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Empty(t, report.GetWarnings())
}

func TestPotentialShadowingRule_Validate_WindowsCriticalShortcuts(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformWindows),
	)

	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.close.app", "alt+f4"),   // Critical: close app on Windows
			keymap.NewActioinBinding("actions.switch.app", "alt+tab"), // Critical: app switching
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetWarnings(), 2)

	// Check warnings contain appropriate messages
	for _, warning := range report.GetWarnings() {
		shadowing := warning.GetPotentialShadowing()
		assert.NotNil(t, shadowing)
		assert.Contains(t, shadowing.GetMessage(), "Windows")
	}
}

func TestPotentialShadowingRule_Validate_CaseInsensitive(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	// Test that case variations are still detected
	setting := &keymapv1.Keymap{
		Actions: []*keymapv1.Action{
			keymap.NewActioinBinding("actions.test", "CMD+Q"), // Uppercase should still match
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	require.NoError(t, err)
	assert.Len(t, report.GetWarnings(), 1)

	warning := report.GetWarnings()[0].GetPotentialShadowing()
	assert.NotNil(t, warning)
	assert.Contains(t, warning.GetMessage(), "quitting applications on macOS")
}
