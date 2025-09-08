package onekeymap

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

// pairKey builds a stable identifier for a keymap by action and normalized key bindings.
// For multi-binding actions, it sorts the formatted bindings and joins them to create a
// deterministic signature per action.
func pairKey(km *keymapv1.ActionBinding) string {
	if km == nil || len(km.GetBindings()) == 0 {
		return km.GetId() + "\x00"
	}
	// Format each binding
	parts := make([]string, 0, len(km.GetBindings()))
	for _, b := range km.GetBindings() {
		if b == nil {
			continue
		}
		parts = append(parts, keymap.MustFormatKeyBinding(keymap.NewKeyBinding(b), platform.PlatformMacOS))
	}
	// Simple insertion sort to avoid importing sort in this small file
	for i := 1; i < len(parts); i++ {
		j := i
		for j > 0 && parts[j] < parts[j-1] {
			parts[j], parts[j-1] = parts[j-1], parts[j]
			j--
		}
	}
	// Join with NUL to avoid ambiguity
	sig := km.GetId() + "\x00"
	for _, p := range parts {
		sig += p + "\x00"
	}
	return sig
}

// dedupKeyBindings removes duplicate keybindings based on (Action, KeyChords)
// using a deterministic signature. The first occurrence is kept and order is preserved.
func dedupKeyBindings(keybindings []*keymapv1.ActionBinding) []*keymapv1.ActionBinding {
	if len(keybindings) == 0 {
		return keybindings
	}
	seen := make(map[string]struct{}, len(keybindings))
	out := make([]*keymapv1.ActionBinding, 0, len(keybindings))
	for _, kb := range keybindings {
		if kb == nil {
			continue
		}
		// Use the unified String() method for consistent signature generation
		sig := pairKey(kb)
		if _, ok := seen[sig]; ok {
			continue
		}
		seen[sig] = struct{}{}
		out = append(out, kb)
	}
	return out
}
