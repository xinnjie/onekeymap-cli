package keybindinglookup

import (
	"io"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
)

type KeybindingLookup interface {
	// Given editor-specific keybinding config and keybinding, return the keybinding config string.
	Lookup(editorSpecificConfig io.Reader, keybinding *keymap.KeyBinding) (keybindingConfig []string, err error)
}
