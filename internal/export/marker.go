package export

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type Marker struct {
	keymap *keymapv1.Keymap
	// per-action exported key set (canonical keybinding string -> true)
	exported map[string]map[string]bool
	// per-action per-keybinding skip reason
	skippedKeys map[string]map[string]error
	// action-level skip reason (applies to all unexported keybindings)
	skippedAction map[string]error
}

func NewMarker(keymap *keymapv1.Keymap) *Marker {
	return &Marker{
		keymap:        keymap,
		exported:      make(map[string]map[string]bool),
		skippedKeys:   make(map[string]map[string]error),
		skippedAction: make(map[string]error),
	}
}

// MarkExported marks a specific keybinding of an action as exported.
// Exporter should always call this method for each exported keybinding.
func (m *Marker) MarkExported(action string, keybinding *keymapv1.Keybinding) {
	if m == nil || keybinding == nil {
		return
	}
	if _, ok := m.exported[action]; !ok {
		m.exported[action] = make(map[string]bool)
	}
	m.exported[action][canonicalKeybindingID(keybinding)] = true
}

// MarkSkippedForReason marks an action or a specific keybinding as skipped for a reason.
// If keybinding is nil, the reason is applied at action level for all unexported keybindings.
// If not called, any keybinding not marked as exported will be filled with
// pluginapi.ErrActionNotSupported in SkipReport.
func (m *Marker) MarkSkippedForReason(action string, keybinding *keymapv1.Keybinding, reasonErr error) {
	if m == nil {
		return
	}
	if reasonErr == nil {
		reasonErr = pluginapi.ErrActionNotSupported
	}
	if keybinding == nil {
		if _, exists := m.skippedAction[action]; !exists {
			m.skippedAction[action] = reasonErr
		}
		return
	}
	if _, ok := m.skippedKeys[action]; !ok {
		m.skippedKeys[action] = make(map[string]error)
	}
	k := canonicalKeybindingID(keybinding)
	if _, exists := m.skippedKeys[action][k]; !exists {
		m.skippedKeys[action][k] = reasonErr
	}
}

func (m *Marker) Report() pluginapi.SkipReport {
	if m == nil || m.keymap == nil {
		return pluginapi.SkipReport{}
	}
	actions := m.keymap.GetActions()
	// ensure stable order by sorting action IDs for determinism in tests
	ids := make([]string, 0, len(actions))
	for _, a := range actions {
		if a == nil {
			continue
		}
		ids = append(ids, a.GetName())
	}
	slices.Sort(ids)
	var result []pluginapi.SkipAction
	for _, id := range ids {
		// Find the action in the original slice to access its bindings
		var act *keymapv1.Action
		for _, a := range actions {
			if a != nil && a.GetName() == id {
				act = a
				break
			}
		}
		if act == nil {
			continue
		}
		// iterate each binding
		for _, br := range act.GetBindings() {
			if br == nil || br.GetKeyChords() == nil {
				continue
			}
			kb := br.GetKeyChords()
			key := canonicalKeybindingID(kb)
			// exported? skip
			if expForAct, ok := m.exported[id]; ok {
				if expForAct[key] {
					continue
				}
			}
			// explicit per-key skip reason?
			if perKey, ok := m.skippedKeys[id]; ok {
				if err, ok2 := perKey[key]; ok2 {
					result = append(result, pluginapi.SkipAction{Action: id, Error: err})
					continue
				}
			}
			// action-level skip reason?
			if err, ok := m.skippedAction[id]; ok {
				result = append(result, pluginapi.SkipAction{Action: id, Error: err})
				continue
			}
			// default
			result = append(result, pluginapi.SkipAction{Action: id, Error: pluginapi.ErrActionNotSupported})
		}
	}
	return pluginapi.SkipReport{SkipActions: result}
}

// canonicalKeybindingID builds a stable string identifier for a keybinding.
// It sorts modifiers in each chord to ensure consistent identity.
func canonicalKeybindingID(kb *keymapv1.Keybinding) string {
	if kb == nil {
		return ""
	}
	parts := make([]string, 0, len(kb.GetChords()))
	for _, ch := range kb.GetChords() {
		if ch == nil {
			continue
		}
		mods := append([]keymapv1.KeyModifier(nil), ch.GetModifiers()...)
		sort.Slice(mods, func(i, j int) bool { return mods[i] < mods[j] })
		parts = append(parts, fmt.Sprintf("%d:%v", ch.GetKeyCode(), mods))
	}
	return strings.Join(parts, " ")
}
