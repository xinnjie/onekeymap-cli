package keybindinglookup

import (
	"io"

	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
)

type KeybindingLookup interface {
	// Given editor-specific keybinding config and keybinding, return the keybinding config string.
	Lookup(editorSpecificConfig io.Reader, keybinding keybinding.Keybinding) (keybindingConfig []string, err error)
}
