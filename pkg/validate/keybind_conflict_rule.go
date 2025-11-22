package validate

import (
	"context"

	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	validateapi "github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

// KeybindConflictRule checks for keybinding conflicts where multiple actions
// are mapped to the same key combination.
type KeybindConflictRule struct{}

// NewKeybindConflictRule creates a new keybinding conflict validation rule.
func NewKeybindConflictRule() validateapi.ValidationRule {
	return &KeybindConflictRule{}
}

// Validate checks for keybinding conflicts in the keymap setting.
func (r *KeybindConflictRule) Validate(_ context.Context, validationContext *validateapi.ValidationContext) error {
	if len(validationContext.Setting.Actions) == 0 {
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
		map[string][]keymap.Action,
	) // key: formatted keybinding, value: list of actions having it

	for _, action := range validationContext.Setting.Actions {
		for _, b := range action.Bindings {
			// Format the key combination for comparison
			formatted := b.String(keybinding.FormatOption{
				Platform:  platform.PlatformMacOS,
				Separator: "+",
			})
			keybindingMap[formatted] = append(keybindingMap[formatted], action)
		}
	}

	// Check for conflicts (multiple actions for same keybinding)
	for keybindingStr, actions := range keybindingMap {
		if len(actions) > 1 {
			// Create action objects with editor commands
			var conflictActions []validateapi.ConflictAction
			for _, act := range actions {
				conflictAction := validateapi.ConflictAction{
					ActionID: act.Name,
				}

				// Try to get editor command from mapping config
				if mappingConfig != nil {
					if mapping := mappingConfig.Get(act.Name); mapping != nil {
						// Get editor command based on source editor from report
						editorCommand := getEditorCommand(mapping, validationContext.Report.SourceEditor)
						conflictAction.Context = editorCommand
					}
				}

				conflictActions = append(conflictActions, conflictAction)
			}

			// Create a keybinding conflict issue
			issue := validateapi.ValidationIssue{
				Type: validateapi.IssueTypeKeybindConflict,
				Details: validateapi.KeybindConflict{
					Keybinding: keybindingStr,
					Actions:    conflictActions,
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
