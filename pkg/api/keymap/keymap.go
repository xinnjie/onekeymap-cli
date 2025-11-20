package keymap

import "github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"

type Keymap struct {
	Actions []Action
}

type Action struct {
	Name     string
	Bindings []keybinding.Keybinding
}
