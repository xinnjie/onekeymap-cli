package export

import (
	"slices"

	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type Marker struct {
	keymap *keymap.Keymap
	// per-action exported key set (canonical keybinding string -> true)
	exported map[string]map[string]bool
	// per-action per-keybinding skip reason
	skippedKeys map[string]map[string]error
	// action-level skip reason (applies to all unexported keybindings)
	skippedAction map[string]error
}

func NewMarker(keymap *keymap.Keymap) *Marker {
	return &Marker{
		keymap:        keymap,
		exported:      make(map[string]map[string]bool),
		skippedKeys:   make(map[string]map[string]error),
		skippedAction: make(map[string]error),
	}
}

// MarkExported marks a specific keybinding of an action as exported.
// Exporter should always call this method for each exported keybinding.
func (m *Marker) MarkExported(action string, kb keybinding.Keybinding) {
	if m == nil {
		return
	}
	if _, ok := m.exported[action]; !ok {
		m.exported[action] = make(map[string]bool)
	}
	m.exported[action][kb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: " "})] = true
}

// MarkSkippedForReason marks an action or a specific keybinding as skipped for a reason.
// If keybinding is nil, the reason is applied at action level for all unexported keybindings.
// If not called, any keybinding not marked as exported will be filled with
// pluginapi.ErrActionNotSupported in SkipReport.
func (m *Marker) MarkSkippedForReason(action string, kb *keybinding.Keybinding, reasonErr error) {
	if m == nil {
		return
	}
	if reasonErr == nil {
		reasonErr = pluginapi.ErrActionNotSupported
	}
	if kb == nil {
		if _, exists := m.skippedAction[action]; !exists {
			m.skippedAction[action] = reasonErr
		}
		return
	}
	if _, ok := m.skippedKeys[action]; !ok {
		m.skippedKeys[action] = make(map[string]error)
	}
	k := kb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: " "})
	if _, exists := m.skippedKeys[action][k]; !exists {
		m.skippedKeys[action][k] = reasonErr
	}
}

func (m *Marker) Report() pluginapi.ExportSkipReport {
	actions := m.keymap.Actions
	// ensure stable order by sorting action IDs for determinism in tests
	ids := make([]string, 0, len(actions))
	for _, a := range actions {
		ids = append(ids, a.Name)
	}
	slices.Sort(ids)
	var result []pluginapi.ExportSkipAction
	for _, id := range ids {
		// Find the action in the original slice to access its bindings
		var act *keymap.Action
		for i := range actions {
			if actions[i].Name == id {
				act = &actions[i]
				break
			}
		}
		if act == nil {
			continue
		}
		// iterate each binding
		for _, kb := range act.Bindings {
			if len(kb.KeyChords) == 0 {
				continue
			}
			key := kb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: " "})
			// exported? skip
			if expForAct, ok := m.exported[id]; ok {
				if expForAct[key] {
					continue
				}
			}
			// explicit per-key skip reason?
			if perKey, ok := m.skippedKeys[id]; ok {
				if err, ok2 := perKey[key]; ok2 {
					result = append(result, pluginapi.ExportSkipAction{Action: id, Error: err})
					continue
				}
			}
			// action-level skip reason?
			if err, ok := m.skippedAction[id]; ok {
				result = append(result, pluginapi.ExportSkipAction{Action: id, Error: err})
				continue
			}
			// default
			result = append(result, pluginapi.ExportSkipAction{Action: id, Error: pluginapi.ErrActionNotSupported})
		}
	}
	return pluginapi.ExportSkipReport{SkipActions: result}
}
