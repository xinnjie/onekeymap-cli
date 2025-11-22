package pluginapi

import (
	"errors"

	keybinding "github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

var (
	// ErrNotSupported is returned when a plugin does not support a requested operation, like import or export.
	ErrNotSupported = errors.New("not supported")

	// ErrActionNotSupported is returned when a plugin does not support a requested action.
	ErrActionNotSupported = errors.New("action not supported")
)

// UnsupportedExportActionError is returned when a plugin does not support a requested action when exporting
type UnsupportedExportActionError struct {
	Note string
}

func (e *UnsupportedExportActionError) Error() string {
	return "not supported: " + e.Note
}

// EditorSupportOnlyOneKeybindingPerActionError is returned when an editor does not support
// assigning multiple keybindings to a single action.
type EditorSupportOnlyOneKeybindingPerActionError struct {
	SkipKeybinding *keybinding.Keybinding
}

func (e *EditorSupportOnlyOneKeybindingPerActionError) Error() string {
	if e == nil || e.SkipKeybinding == nil {
		return "editor support only one keybinding per action"
	}
	return "editor support only one keybinding per action: " + e.SkipKeybinding.String(
		keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"},
	)
}
