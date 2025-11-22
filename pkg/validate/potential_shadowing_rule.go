package validate

import (
	"context"
	"strings"

	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	validateapi "github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
)

// PotentialShadowingRule detects keybindings that might shadow critical system or editor shortcuts.
type PotentialShadowingRule struct {
	targetEditor pluginapi.EditorType
	platform     platform.Platform
}

// NewPotentialShadowingRule creates a new potential shadowing validation rule.
func NewPotentialShadowingRule(
	targetEditor pluginapi.EditorType,
	platform platform.Platform,
) validateapi.ValidationRule {
	return &PotentialShadowingRule{
		targetEditor: targetEditor,
		platform:     platform,
	}
}

// TODO(xinnjie): Read keybindings from system, e.g. read macos system keybindings from `~/Library/Preferences/com.apple.symbolichotkeys.plist`
// criticalKeybindingsByPlatform defines system-critical keybindings that should not be overridden, organized by platform.
//
//nolint:gochecknoglobals // shared reference data; treated as constant map and accessed read-only
var criticalKeybindingsByPlatform = map[platform.Platform]map[string]string{
	platform.PlatformMacOS: {
		// macOS specific shortcuts
		"cmd+q":       "quitting applications on macOS",
		"cmd+w":       "closing windows on macOS",
		"cmd+m":       "minimizing windows on macOS",
		"cmd+h":       "hiding applications on macOS",
		"cmd+tab":     "application switching on macOS",
		"cmd+space":   "Spotlight search on macOS",
		"cmd+shift+3": "screenshot on macOS",
		"cmd+shift+4": "screenshot selection on macOS",
		// Universal shortcuts on macOS
		"cmd+c": "copy on macOS",
		"cmd+v": "paste on macOS",
		"cmd+x": "cut on macOS",
		"cmd+z": "undo on macOS",
		"cmd+y": "redo on macOS",
		"cmd+s": "save on macOS",
		"cmd+a": "select all on macOS",
	},
	platform.PlatformWindows: {
		// Windows specific shortcuts
		"alt+f4":         "closing applications on Windows",
		"alt+tab":        "application switching on Windows",
		"ctrl+shift+esc": "task manager on Windows",
		"win+l":          "locking screen on Windows",
		// Universal shortcuts on Windows
		"ctrl+c": "copy on Windows",
		"ctrl+v": "paste on Windows",
		"ctrl+x": "cut on Windows",
		"ctrl+z": "undo on Windows",
		"ctrl+y": "redo on Windows",
		"ctrl+s": "save on Windows",
		"ctrl+a": "select all on Windows",
	},
	platform.PlatformLinux: {
		// Linux specific shortcuts
		"alt+f4":     "closing applications on Linux",
		"alt+tab":    "application switching on Linux",
		"super+l":    "locking screen on Linux",
		"ctrl+alt+t": "opening terminal on Linux",
		// Universal shortcuts on Linux
		"ctrl+c": "copy on Linux",
		"ctrl+v": "paste on Linux",
		"ctrl+x": "cut on Linux",
		"ctrl+z": "undo on Linux",
		"ctrl+y": "redo on Linux",
		"ctrl+s": "save on Linux",
		"ctrl+a": "select all on Linux",
	},
}

// Validate checks for keybindings that might shadow critical system shortcuts.
func (r *PotentialShadowingRule) Validate(_ context.Context, validationContext *validateapi.ValidationContext) error {
	setting := validationContext.Setting
	report := validationContext.Report

	// Get critical keybindings for the target platform
	criticalKeybindings, exists := criticalKeybindingsByPlatform[r.platform]
	if !exists {
		// If platform is not supported, skip validation
		return nil
	}

	for _, action := range setting.Actions {
		for _, b := range action.Bindings {
			// Format the key binding to get a consistent string representation
			formattedKeys := b.String(keybinding.FormatOption{
				Platform:  r.platform,
				Separator: "+",
			})
			// Normalize the key combination for comparison
			normalizedKeys := strings.ToLower(formattedKeys)
			// Check if this keybinding shadows a critical shortcut
			if description, isCritical := criticalKeybindings[normalizedKeys]; isCritical {
				// Add warning for potential shadowing
				warning := validateapi.ValidationIssue{
					Type: validateapi.IssueTypePotentialShadowing,
					Details: validateapi.PotentialShadowing{
						Keybinding:                  formattedKeys,
						Action:                      action.Name,
						CriticalShortcutDescription: "This key chord is the default for " + description + ".",
					},
				}
				report.Warnings = append(report.Warnings, warning)
			}
		}
	}

	return nil
}
