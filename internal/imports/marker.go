package imports

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type Marker struct {
	imported map[string]bool
	skipped  map[string]skippedEntry
	order    []string

	// importedResults tracks detailed import results keyed by "mappedAction:editorAction"
	importedResults map[string]*pluginapi.KeybindingImportResult
	importedOrder   []string
}

type skippedEntry struct {
	err         error
	keybindings []keybinding.Keybinding
}

func NewMarker() *Marker {
	return &Marker{
		imported:        make(map[string]bool),
		skipped:         make(map[string]skippedEntry),
		importedResults: make(map[string]*pluginapi.KeybindingImportResult),
	}
}

// MarkImported records a successful import with full details for coverage reporting.
func (m *Marker) MarkImported(
	mappedAction string,
	editorSpecificAction string,
	originalKeybinding keybinding.Keybinding,
	importedKeybinding keybinding.Keybinding,
) {
	if mappedAction == "" || editorSpecificAction == "" {
		return
	}
	// mark as imported and clear any previous skipped entry
	m.imported[editorSpecificAction] = true
	delete(m.skipped, editorSpecificAction)

	key := mappedAction + "\x00" + editorSpecificAction
	result, exists := m.importedResults[key]
	if !exists {
		result = &pluginapi.KeybindingImportResult{
			MappedAction:         mappedAction,
			EditorSpecificAction: editorSpecificAction,
		}
		m.importedResults[key] = result
		m.importedOrder = append(m.importedOrder, key)
	}
	result.OriginalKeybindings = append(result.OriginalKeybindings, originalKeybinding)
	result.ImportedKeybindings = append(result.ImportedKeybindings, importedKeybinding)
}

// MarkSkipped records a skipped action with its keybinding for coverage reporting.
func (m *Marker) MarkSkipped(editorSpecificAction string, kb *keybinding.Keybinding, reasonErr error) {
	if editorSpecificAction == "" {
		return
	}
	if reasonErr == nil {
		reasonErr = pluginapi.ErrNotSupported
	}
	if m.imported[editorSpecificAction] {
		return
	}
	entry, exists := m.skipped[editorSpecificAction]
	if !exists {
		entry = skippedEntry{err: reasonErr}
		m.order = append(m.order, editorSpecificAction)
	}
	if kb != nil {
		entry.keybindings = append(entry.keybindings, *kb)
	}
	m.skipped[editorSpecificAction] = entry
}

func (m *Marker) Report() pluginapi.ImportSkipReport {
	result := make([]pluginapi.ImportSkipAction, 0, len(m.skipped))
	for _, action := range m.order {
		if m.imported[action] {
			continue
		}
		if entry, ok := m.skipped[action]; ok {
			result = append(result, pluginapi.ImportSkipAction{
				EditorSpecificAction: action,
				Keybindings:          entry.keybindings,
				Error:                entry.err,
			})
		}
	}
	return pluginapi.ImportSkipReport{SkipActions: result}
}

// ImportedReport returns the detailed import results for coverage reporting.
func (m *Marker) ImportedReport() pluginapi.ImportedReport {
	results := make([]pluginapi.KeybindingImportResult, 0, len(m.importedResults))
	for _, key := range m.importedOrder {
		if result, ok := m.importedResults[key]; ok {
			results = append(results, *result)
		}
	}
	return pluginapi.ImportedReport{Results: results}
}
