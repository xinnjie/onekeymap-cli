package validateapi

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// KeybindConflictRule checks for keybinding conflicts where multiple actions
// are mapped to the same key combination.
type KeybindConflictRule struct{}

// NewKeybindConflictRule creates a new keybinding conflict validation rule.
func NewKeybindConflictRule() ValidationRule {
	return &KeybindConflictRule{}
}

// Validate checks for keybinding conflicts in the keymap setting.
func (r *KeybindConflictRule) Validate(ctx context.Context, validationContext *ValidationContext) error {
	if validationContext.Setting == nil || len(validationContext.Setting.GetKeybindings()) == 0 {
		return nil
	}

	// Load mapping config to get editor commands
	mappingConfig, err := mappings.NewMappingConfig()
	if err != nil {
		// If we can't load mappings, we can still detect conflicts without editor commands
		mappingConfig = nil
	}

	// Group keybindings by their formatted key combination
	keybindingMap := make(
		map[string][]*keymapv1.Action,
	) // key: formatted keybinding, value: list of action bindings having it

	for _, ab := range validationContext.Setting.GetKeybindings() {
		if ab == nil {
			continue
		}
		for _, b := range ab.GetBindings() {
			if b == nil {
				continue
			}
			// Format the key combination for comparison
			formatted, err := keymap.NewKeyBinding(b).Format(platform.PlatformMacOS, "+")
			if err != nil {
				// Skip invalid keybindings but don't fail validation
				continue
			}
			keybindingMap[formatted] = append(keybindingMap[formatted], ab)
		}
	}

	// Check for conflicts (multiple actions for same keybinding)
	for keybinding, keybindings := range keybindingMap {
		if len(keybindings) > 1 {
			// Create action objects with editor commands
			var actions []*keymapv1.KeybindConflict_Action
			for _, kb := range keybindings {
				action := &keymapv1.KeybindConflict_Action{
					Action: kb.GetName(),
				}

				// Try to get editor command from mapping config
				if mappingConfig != nil {
					if mapping := mappingConfig.FindByUniversalAction(kb.GetName()); mapping != nil {
						// Get editor command based on source editor from report
						editorCommand := getEditorCommand(mapping, validationContext.Report.GetSourceEditor())
						action.EditorCommand = editorCommand
					}
				}

				actions = append(actions, action)
			}

			// Create a keybinding conflict issue
			conflict := &keymapv1.KeybindConflict{
				Keybinding: keybinding,
				Actions:    actions,
			}

			issue := &keymapv1.ValidationIssue{
				Issue: &keymapv1.ValidationIssue_KeybindConflict{
					KeybindConflict: conflict,
				},
			}

			validationContext.Report.Issues = append(validationContext.Report.Issues, issue)
		}
	}

	return nil
}

// getEditorCommand extracts the editor command from mapping config based on source editor.
func getEditorCommand(mapping *mappings.ActionMappingConfig, sourceEditor string) string {
	switch sourceEditor {
	case "vscode":
		if len(mapping.VSCode) > 0 {
			return mapping.VSCode[0].Command
		}
	case "zed":
		if len(mapping.Zed) > 0 {
			return mapping.Zed[0].Action
		}
	case "intellij":
		return mapping.IntelliJ.Action
	case "helix":
		if len(mapping.Helix) > 0 {
			return mapping.Helix[0].Command
		}
	case "vim":
		return mapping.Vim.Command
	}
	return ""
}
