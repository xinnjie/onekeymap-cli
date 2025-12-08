package imports_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keychord"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keycode"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

func TestMarker_ReportEmptyWhenNoSkips(t *testing.T) {
	m := imports.NewMarker()
	rep := m.Report()
	require.Empty(t, rep.SkipActions)
}

func TestMarker_RecordSkipWithReason(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkipped("editor.action.copy", nil, errors.New("some reason"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, "editor.action.copy", sk.EditorSpecificAction)
	assert.ErrorContains(t, sk.Error, "some reason")
}

func TestMarker_DefaultErrorWhenReasonNil(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkipped("editor.action.paste", nil, nil)

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, "editor.action.paste", sk.EditorSpecificAction)
	assert.ErrorIs(t, sk.Error, pluginapi.ErrNotSupported)
}

func TestMarker_ImportedWithDetailsOverridesSkip(t *testing.T) {
	m := imports.NewMarker()
	kb := makeTestKeybinding("ctrl+c")
	m.MarkSkipped("editor.action.cut", nil, errors.New("temporary"))
	m.MarkImported("actions.edit.cut", "editor.action.cut", kb, kb)

	rep := m.Report()
	require.Empty(t, rep.SkipActions)

	importedRep := m.ImportedReport()
	require.Len(t, importedRep.Results, 1)
	assert.Equal(t, "actions.edit.cut", importedRep.Results[0].MappedAction)
}

func TestMarker_DeduplicateSkipsPreserveOrder(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkipped("editor.action.copy", nil, errors.New("first"))
	m.MarkSkipped("editor.action.paste", nil, errors.New("second"))
	m.MarkSkipped("editor.action.copy", nil, errors.New("override"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 2)
	assert.Equal(t, "editor.action.copy", rep.SkipActions[0].EditorSpecificAction)
	assert.Equal(t, "editor.action.paste", rep.SkipActions[1].EditorSpecificAction)
	assert.ErrorContains(t, rep.SkipActions[0].Error, "first")
}

func TestMarker_ImportedWithDetailsPreventsLaterSkip(t *testing.T) {
	m := imports.NewMarker()
	kb := makeTestKeybinding("ctrl+c")
	m.MarkImported("actions.edit.copy", "editor.action.copy", kb, kb)
	m.MarkSkipped("editor.action.copy", nil, errors.New("should be ignored"))

	rep := m.Report()
	require.Empty(t, rep.SkipActions)
}

func TestMarker_ImportedReportEmpty(t *testing.T) {
	m := imports.NewMarker()
	rep := m.ImportedReport()
	require.Empty(t, rep.Results)
}

func TestMarker_ImportedReportTracksDetails(t *testing.T) {
	m := imports.NewMarker()
	kb1 := makeTestKeybinding("ctrl+c")
	kb2 := makeTestKeybinding("ctrl+v")

	m.MarkImported("actions.edit.copy", "editor.action.copy", kb1, kb1)
	m.MarkImported("actions.edit.paste", "editor.action.paste", kb2, kb2)

	rep := m.ImportedReport()
	require.Len(t, rep.Results, 2)

	assert.Equal(t, "actions.edit.copy", rep.Results[0].MappedAction)
	assert.Equal(t, "editor.action.copy", rep.Results[0].EditorSpecificAction)
	require.Len(t, rep.Results[0].OriginalKeybindings, 1)
	require.Len(t, rep.Results[0].ImportedKeybindings, 1)

	assert.Equal(t, "actions.edit.paste", rep.Results[1].MappedAction)
	assert.Equal(t, "editor.action.paste", rep.Results[1].EditorSpecificAction)
}

func TestMarker_ImportedReportAggregatesMultipleKeybindings(t *testing.T) {
	m := imports.NewMarker()
	kb1 := makeTestKeybinding("ctrl+c")
	kb2 := makeTestKeybinding("cmd+c")

	// Same action with multiple keybindings
	m.MarkImported("actions.edit.copy", "editor.action.copy", kb1, kb1)
	m.MarkImported("actions.edit.copy", "editor.action.copy", kb2, kb2)

	rep := m.ImportedReport()
	require.Len(t, rep.Results, 1)
	assert.Equal(t, "actions.edit.copy", rep.Results[0].MappedAction)
	require.Len(t, rep.Results[0].OriginalKeybindings, 2)
	require.Len(t, rep.Results[0].ImportedKeybindings, 2)
}

func TestMarker_SkippedWithKeybindingTracksKeybindings(t *testing.T) {
	m := imports.NewMarker()
	kb := makeTestKeybinding("ctrl+x")

	m.MarkSkipped("unknown.action", &kb, errors.New("not supported"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	assert.Equal(t, "unknown.action", rep.SkipActions[0].EditorSpecificAction)
	require.Len(t, rep.SkipActions[0].Keybindings, 1)
}

func TestMarker_SkippedWithKeybindingAggregatesKeybindings(t *testing.T) {
	m := imports.NewMarker()
	kb1 := makeTestKeybinding("ctrl+x")
	kb2 := makeTestKeybinding("cmd+x")

	m.MarkSkipped("unknown.action", &kb1, errors.New("not supported"))
	m.MarkSkipped("unknown.action", &kb2, errors.New("not supported"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	require.Len(t, rep.SkipActions[0].Keybindings, 2)
}

// makeTestKeybinding creates a simple keybinding for testing
func makeTestKeybinding(key string) keybinding.Keybinding {
	// Simple keybinding with a single keychord
	return keybinding.Keybinding{
		KeyChords: []keychord.KeyChord{
			{KeyCode: keycode.KeyCode(key)},
		},
	}
}
