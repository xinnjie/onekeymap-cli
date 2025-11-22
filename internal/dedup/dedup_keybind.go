package dedup

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
)

func mergeIntoExistingActionStruct(existing *keymap.Action, kb keymap.Action) {
	// Merge bindings into existing
	for _, b := range kb.Bindings {
		if len(b.KeyChords) == 0 {
			continue
		}
		dup := false
		nbf := b.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
		for _, eb := range existing.Bindings {
			ebf := eb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
			if ebf == nbf {
				dup = true
				break
			}
		}
		if !dup {
			existing.Bindings = append(existing.Bindings, b)
		}
	}
}

// Actions removes duplicate keybindings based on (Action, KeyChords)
// using a deterministic signature. The first occurrence is kept and order is preserved.
func Actions(actions []keymap.Action) []keymap.Action {
	if len(actions) == 0 {
		return actions
	}
	// Merge by action ID, concatenating unique bindings while preserving first metadata and order
	idxByID := make(map[string]int, len(actions))
	out := make([]keymap.Action, 0, len(actions))

	for _, kb := range actions {
		id := kb.Name
		if pos, ok := idxByID[id]; ok {
			mergeIntoExistingActionStruct(&out[pos], kb)
			continue
		}
		// First occurrence: create a fresh ActionBinding and deduplicate its own bindings
		fresh := keymap.Action{
			Name: kb.Name,
		}

		hadBindings := len(kb.Bindings) > 0
		for _, b := range kb.Bindings {
			if len(b.KeyChords) == 0 {
				continue
			}
			dup := false
			nbf := b.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
			for _, eb := range fresh.Bindings {
				ebf := eb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
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
		if len(fresh.Bindings) == 0 && hadBindings {
			continue
		}
		idxByID[id] = len(out)
		out = append(out, fresh)
	}
	return out
}
