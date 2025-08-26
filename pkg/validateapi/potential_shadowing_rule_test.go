package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestPotentialShadowingRule_Validate_WithCriticalKeybinding(t *testing.T) {
	validator := NewValidator(NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS))

	// Create keymaps with critical system shortcuts
	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			keymap.NewBinding("actions.format.document", "cmd+q"), // Critical: quit on macOS
			keymap.NewBinding("actions.edit.copy", "cmd+c"),       // Critical: copy on macOS
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 2)

	// Check first warning (cmd+q)
	warning1 := report.Warnings[0].GetPotentialShadowing()
	assert.NotNil(t, warning1)
	assert.Equal(t, "actions.format.document", warning1.Action)
	assert.Contains(t, warning1.Message, "quitting applications on macOS")

	// Check second warning (cmd+c)
	warning2 := report.Warnings[1].GetPotentialShadowing()
	assert.NotNil(t, warning2)
	assert.Equal(t, "actions.edit.copy", warning2.Action)
	assert.Contains(t, warning2.Message, "copy on macOS")
}

func TestPotentialShadowingRule_Validate_NoCriticalKeybindings(t *testing.T) {
	validator := NewValidator(NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS))

	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			keymap.NewBinding("actions.custom.action", "ctrl+shift+f12"),
			keymap.NewBinding("actions.another.action", "alt+shift+g"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 0)
}

func TestPotentialShadowingRule_Validate_WindowsCriticalShortcuts(t *testing.T) {
	validator := NewValidator(NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformWindows))

	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			keymap.NewBinding("actions.close.app", "alt+f4"),   // Critical: close app on Windows
			keymap.NewBinding("actions.switch.app", "alt+tab"), // Critical: app switching
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 2)

	// Check warnings contain appropriate messages
	for _, warning := range report.Warnings {
		shadowing := warning.GetPotentialShadowing()
		assert.NotNil(t, shadowing)
		assert.Contains(t, shadowing.Message, "Windows")
	}
}

func TestPotentialShadowingRule_Validate_CaseInsensitive(t *testing.T) {
	validator := NewValidator(NewPotentialShadowingRule(pluginapi.EditorTypeVSCode, platform.PlatformMacOS))

	// Test that case variations are still detected
	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.KeyBinding{
			keymap.NewBinding("actions.test", "CMD+Q"), // Uppercase should still match
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 1)

	warning := report.Warnings[0].GetPotentialShadowing()
	assert.NotNil(t, warning)
	assert.Contains(t, warning.Message, "quitting applications on macOS")
}
