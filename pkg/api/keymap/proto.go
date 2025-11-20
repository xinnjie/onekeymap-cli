package keymap

import (
	"fmt"

	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

// FromProto converts a keymapv1.Keymap to the API Keymap.
func FromProto(in *keymapv1.Keymap) (Keymap, error) {
	if in == nil {
		return Keymap{}, nil
	}
	out := Keymap{Actions: make([]Action, 0, len(in.GetActions()))}
	for _, a := range in.GetActions() {
		if a == nil {
			continue
		}
		var bindings []keybinding.Keybinding
		for _, b := range a.GetBindings() {
			if b == nil {
				continue
			}
			readable := b.GetKeyChordsReadable()
			if readable == "" {
				return Keymap{}, fmt.Errorf("missing key_chords_readable for action %q", a.GetName())
			}
			kb, err := keybinding.NewKeybinding(readable, keybinding.ParseOption{Separator: "+"})
			if err != nil {
				return Keymap{}, fmt.Errorf("parse keybinding for action %q: %w", a.GetName(), err)
			}
			bindings = append(bindings, kb)
		}
		out.Actions = append(out.Actions, Action{
			Name:     a.GetName(),
			Bindings: bindings,
		})
	}
	return out, nil
}

// ToProto converts the API Keymap to keymapv1.Keymap.
// Only the readable form is populated for each binding.
func ToProto(in Keymap, p platform.Platform) *keymapv1.Keymap {
	out := &keymapv1.Keymap{Actions: make([]*keymapv1.Action, 0, len(in.Actions))}
	for _, a := range in.Actions {
		pa := &keymapv1.Action{Name: a.Name}
		for _, b := range a.Bindings {
			readable := b.String(keybinding.FormatOption{Platform: p, Separator: "+"})
			pa.Bindings = append(pa.Bindings, &keymapv1.KeybindingReadable{KeyChordsReadable: readable})
		}
		out.Actions = append(out.Actions, pa)
	}
	return out
}
