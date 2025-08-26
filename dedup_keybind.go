package onekeymap

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// pairKey builds a stable identifier for a keymap by action and normalized key binding.
func pairKey(km *keymapv1.KeyBinding) string {
	kb := keymap.NewKeyBinding(km)
	return kb.String()
}

// dedupKeyBindings removes duplicate keybindings based on (Action, KeyChords)
// using a deterministic signature. The first occurrence is kept and order is preserved.
func dedupKeyBindings(keybindings []*keymapv1.KeyBinding) []*keymapv1.KeyBinding {
	if len(keybindings) == 0 {
		return keybindings
	}
	seen := make(map[string]struct{}, len(keybindings))
	out := make([]*keymapv1.KeyBinding, 0, len(keybindings))
	for _, kb := range keybindings {
		if kb == nil {
			continue
		}
		// Use the unified String() method for consistent signature generation
		keyBinding := keymap.NewKeyBinding(kb)
		sig := keyBinding.String()
		if _, ok := seen[sig]; ok {
			continue
		}
		seen[sig] = struct{}{}
		out = append(out, kb)
	}
	return out
}
