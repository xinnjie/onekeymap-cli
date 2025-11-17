package imports

import (
	pluginapi2 "github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type Marker struct {
	imported map[string]bool
	skipped  map[string]error
	order    []string
}

func NewMarker() *Marker {
	return &Marker{
		imported: make(map[string]bool),
		skipped:  make(map[string]error),
	}
}

func (m *Marker) MarkImported(editorSpecificAction string) {
	if editorSpecificAction == "" {
		return
	}
	m.imported[editorSpecificAction] = true
	delete(m.skipped, editorSpecificAction)
}

func (m *Marker) MarkSkippedForReason(editorSpecificAction string, reasonErr error) {
	if editorSpecificAction == "" {
		return
	}
	if reasonErr == nil {
		reasonErr = pluginapi2.ErrNotSupported
	}
	if m.imported[editorSpecificAction] {
		return
	}
	if _, exists := m.skipped[editorSpecificAction]; !exists {
		m.skipped[editorSpecificAction] = reasonErr
		m.order = append(m.order, editorSpecificAction)
	}
}

func (m *Marker) Report() pluginapi2.ImportSkipReport {
	result := make([]pluginapi2.ImportSkipAction, 0, len(m.skipped))
	for _, action := range m.order {
		if m.imported[action] {
			continue
		}
		if err, ok := m.skipped[action]; ok {
			result = append(result, pluginapi2.ImportSkipAction{
				EditorSpecificAction: action,
				Error:                err,
			})
		}
	}
	return pluginapi2.ImportSkipReport{SkipActions: result}
}
