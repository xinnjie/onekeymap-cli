package validate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

func TestPotentialShadowingRule_Validate_WithCriticalKeybinding(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	// Create keymaps with critical system shortcuts
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.format.document", "cmd+q"), // Critical: quit on macOS
			newAction("actions.edit.copy", "cmd+c"),       // Critical: copy on macOS
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Warnings, 2)

	// Check first warning (cmd+q)
	assert.Equal(t, validateapi.IssueTypePotentialShadowing, report.Warnings[0].Type)
	warning1, ok := report.Warnings[0].Details.(validateapi.PotentialShadowing)
	require.True(t, ok)
	assert.Equal(t, "actions.format.document", warning1.Action)
	assert.Contains(t, warning1.CriticalShortcutDescription, "quitting applications on macOS")

	// Check second warning (cmd+c)
	assert.Equal(t, validateapi.IssueTypePotentialShadowing, report.Warnings[1].Type)
	warning2, ok := report.Warnings[1].Details.(validateapi.PotentialShadowing)
	require.True(t, ok)
	assert.Equal(t, "actions.edit.copy", warning2.Action)
	assert.Contains(t, warning2.CriticalShortcutDescription, "copy on macOS")
}

func TestPotentialShadowingRule_Validate_NoCriticalKeybindings(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.custom.action", "ctrl+shift+f12"),
			newAction("actions.another.action", "alt+shift+g"),
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Empty(t, report.Warnings)
}

func TestPotentialShadowingRule_Validate_WindowsCriticalShortcuts(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformWindows),
	)

	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.close.app", "alt+f4"),   // Critical: close app on Windows
			newAction("actions.switch.app", "alt+tab"), // Critical: app switching
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Warnings, 2)

	// Check warnings contain appropriate messages
	for _, warning := range report.Warnings {
		assert.Equal(t, validateapi.IssueTypePotentialShadowing, warning.Type)
		shadowing, ok := warning.Details.(validateapi.PotentialShadowing)
		require.True(t, ok)
		assert.Contains(t, shadowing.CriticalShortcutDescription, "Windows")
	}
}

func TestPotentialShadowingRule_Validate_CaseInsensitive(t *testing.T) {
	validator := validateapi.NewValidator(
		validate.NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS),
	)

	// Test that case variations are still detected
	setting := keymap.Keymap{
		Actions: []keymap.Action{
			newAction("actions.test", "CMD+Q"), // Uppercase should still match
		},
	}

	opts := importerapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts.EditorType)
	require.NoError(t, err)
	assert.Len(t, report.Warnings, 1)

	assert.Equal(t, validateapi.IssueTypePotentialShadowing, report.Warnings[0].Type)
	warning, ok := report.Warnings[0].Details.(validateapi.PotentialShadowing)
	require.True(t, ok)
	assert.Contains(t, warning.CriticalShortcutDescription, "quitting applications on macOS")
}
