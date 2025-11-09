package pluginapi

import (
	"errors"

	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

var (
	// ErrNotSupported is returned when a plugin does not support a requested operation, like import or export.
	ErrNotSupported = errors.New("not supported")

	// ErrActionNotSupported is returned when a plugin does not support a requested action.
	ErrActionNotSupported = errors.New("action not supported")
)

type NotSupportedError struct {
	Note string
}

func (e *NotSupportedError) Error() string {
	return "not supported: " + e.Note
}

// EditorSupportOnlyOneKeybindingPerActionError is returned when an editor does not support
// assigning multiple keybindings to a single action.
type EditorSupportOnlyOneKeybindingPerActionError struct {
	SkipKeybinding *keymapv1.Keybinding
}

func (e *EditorSupportOnlyOneKeybindingPerActionError) Error() string {
	if e == nil || e.SkipKeybinding == nil {
		return "editor support only one keybinding per action"
	}
	return "editor support only one keybinding per action: " + e.SkipKeybinding.String()
}
