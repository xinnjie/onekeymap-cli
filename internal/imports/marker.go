package imports

import (
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
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
		reasonErr = pluginapi.ErrNotSupported
	}
	if m.imported[editorSpecificAction] {
		return
	}
	if _, exists := m.skipped[editorSpecificAction]; !exists {
		m.skipped[editorSpecificAction] = reasonErr
		m.order = append(m.order, editorSpecificAction)
	}
}

func (m *Marker) Report() pluginapi.ImportSkipReport {
	result := make([]pluginapi.ImportSkipAction, 0, len(m.skipped))
	for _, action := range m.order {
		if m.imported[action] {
			continue
		}
		if err, ok := m.skipped[action]; ok {
			result = append(result, pluginapi.ImportSkipAction{
				EditorSpecificAction: action,
				Error:                err,
			})
		}
	}
	return pluginapi.ImportSkipReport{SkipActions: result}
}
