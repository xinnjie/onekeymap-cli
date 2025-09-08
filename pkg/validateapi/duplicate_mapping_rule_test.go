package validateapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestDuplicateMappingRule_Validate_WithDuplicates(t *testing.T) {
	validator := NewValidator(NewDuplicateMappingRule())

	// Create keymaps with duplicate mappings
	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"), // Duplicate
			keymap.NewActioinBinding("actions.edit.paste", "ctrl+v"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 1)
	assert.NotNil(t, report.Warnings[0].GetDuplicateMapping())

	duplicate := report.Warnings[0].GetDuplicateMapping()
	assert.Equal(t, "actions.edit.copy", duplicate.Action)
	assert.Contains(t, duplicate.Message, "multiple times")
}

func TestDuplicateMappingRule_Validate_NoDuplicates(t *testing.T) {
	validator := NewValidator(NewDuplicateMappingRule())

	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.paste", "ctrl+v"),
			keymap.NewActioinBinding("actions.edit.cut", "ctrl+x"),
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 0)
}

func TestDuplicateMappingRule_Validate_SameActionDifferentKeys(t *testing.T) {
	validator := NewValidator(NewDuplicateMappingRule())

	// Same action with different keys should not be flagged as duplicate
	setting := &keymapv1.KeymapSetting{
		Keybindings: []*keymapv1.ActionBinding{
			keymap.NewActioinBinding("actions.edit.copy", "ctrl+c"),
			keymap.NewActioinBinding("actions.edit.copy", "cmd+c"), // Different keys
		},
	}

	opts := importapi.ImportOptions{
		EditorType: "vscode",
	}

	report, err := validator.Validate(context.Background(), setting, opts)
	assert.NoError(t, err)
	assert.Len(t, report.Warnings, 0)
}
