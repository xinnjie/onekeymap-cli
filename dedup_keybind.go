package onekeymap

import (
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
	"google.golang.org/protobuf/proto"
)

// pairKey builds a stable identifier for a keymap by action and normalized key bindings.
// For multi-binding actions, it sorts the formatted bindings and joins them to create a
// deterministic signature per action.
func pairKey(km *keymapv1.Action) string {
	if km == nil || len(km.GetBindings()) == 0 {
		return km.GetName() + "\x00"
	}
	// Format each binding
	parts := make([]string, 0, len(km.GetBindings()))
	for _, b := range km.GetBindings() {
		if b == nil || b.GetKeyChords() == nil || len(b.GetKeyChords().GetChords()) == 0 {
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
	sig := km.GetName() + "\x00"
	for _, p := range parts {
		sig += p + "\x00"
	}
	return sig
}

// dedupKeyBindings removes duplicate keybindings based on (Action, KeyChords)
// using a deterministic signature. The first occurrence is kept and order is preserved.
func dedupKeyBindings(keybindings []*keymapv1.Action) []*keymapv1.Action {
	if len(keybindings) == 0 {
		return keybindings
	}
	// Merge by action ID, concatenating unique bindings while preserving first metadata and order
	idxByID := make(map[string]int, len(keybindings))
	out := make([]*keymapv1.Action, 0, len(keybindings))

	for _, kb := range keybindings {
		if kb == nil {
			continue
		}
		id := kb.GetName()
		if pos, ok := idxByID[id]; ok {
			// Merge bindings into existing
			existing := out[pos]
			for _, b := range kb.GetBindings() {
				if len(b.GetKeyChords().GetChords()) == 0 {
					continue
				}
				dup := false
				nb := keymap.NewKeyBinding(b)
				nbf := keymap.MustFormatKeyBinding(nb, platform.PlatformMacOS)
				for _, eb := range existing.GetBindings() {
					ebf := keymap.MustFormatKeyBinding(keymap.NewKeyBinding(eb), platform.PlatformMacOS)
					if ebf == nbf {
						dup = true
						break
					}
				}
				if !dup {
					existing.Bindings = append(existing.Bindings, b)
				}
			}
			// Optionally fill missing metadata from later entries
			if existing.GetName() == "" && kb.GetName() != "" {
				existing.Name = kb.GetName()
			}
			if kb.GetActionConfig() != nil {
				if existing.GetActionConfig() == nil {
					existing.ActionConfig = &keymapv1.ActionConfig{}
				}
				if existing.GetActionConfig().GetDescription() == "" && kb.GetActionConfig().GetDescription() != "" {
					existing.ActionConfig.Description = kb.GetActionConfig().GetDescription()
				}
				if existing.GetActionConfig().GetCategory() == "" && kb.GetActionConfig().GetCategory() != "" {
					existing.ActionConfig.Category = kb.GetActionConfig().GetCategory()
				}
			}
			continue
		}
		// First occurrence: create a fresh ActionBinding and deduplicate its own bindings
		fresh := &keymapv1.Action{
			Name: kb.GetName(),
		}
		// Only clone ActionConfig if it exists
		if kb.GetActionConfig() != nil {
			cloned := proto.Clone(kb.GetActionConfig())
			if ac, ok := cloned.(*keymapv1.ActionConfig); ok {
				fresh.ActionConfig = ac
			}
		}
		hadBindings := len(kb.GetBindings()) > 0
		for _, b := range kb.GetBindings() {
			if len(b.GetKeyChords().GetChords()) == 0 {
				continue
			}
			dup := false
			nbf := keymap.MustFormatKeyBinding(keymap.NewKeyBinding(b), platform.PlatformMacOS)
			for _, eb := range fresh.GetBindings() {
				ebf := keymap.MustFormatKeyBinding(keymap.NewKeyBinding(eb), platform.PlatformMacOS)
				if ebf == nbf {
					dup = true
					break
				}
			}
			if !dup {
				fresh.Bindings = append(fresh.Bindings, b)
			}
		}
		// If there were explicit bindings but all were invalid/empty -> drop this action entirely
		if len(fresh.GetBindings()) == 0 && hadBindings {
			continue
		}
		idxByID[id] = len(out)
		out = append(out, fresh)
	}
	return out
}
