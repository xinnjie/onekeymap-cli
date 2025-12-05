package export_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	exp "github.com/xinnjie/onekeymap-cli/internal/export"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

func newKeymapWithTwoBindings(action string, k1, k2 string) (keymap.Keymap, keymap.Action) {
	kb1, err := keybinding.NewKeybinding(k1, keybinding.ParseOption{Separator: "+"})
	if err != nil {
		panic(err)
	}
	kb2, err := keybinding.NewKeybinding(k2, keybinding.ParseOption{Separator: "+"})
	if err != nil {
		panic(err)
	}
	a := keymap.Action{Name: action, Bindings: []keybinding.Keybinding{kb1, kb2}}
	return keymap.Keymap{Actions: []keymap.Action{a}}, a
}

func TestMarker_ImplicitSkipForUnexportedBindings(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.dual", "cmd+a", "cmd+c")
	m := exp.NewMarker(&km)

	// Export only first binding
	m.MarkExported(a.Name, a.Bindings[0])

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	// The skipped one should be the second binding with default not supported
	sk := rep.SkipActions[0]
	assert.Equal(t, a.Name, sk.Action)
	assert.ErrorIs(t, sk.Error, pluginapi.ErrActionNotSupported)
}

func TestMarker_ExplicitKeySkip(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.dual", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Export first, explicitly skip second
	m.MarkExported(a.Name, a.Bindings[0])
	kb, err := keybinding.NewKeybinding("cmd+b", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	m.MarkSkippedForReason(
		a.Name,
		&kb,
		&pluginapi.EditorSupportOnlyOneKeybindingPerActionError{},
	)

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, a.Name, sk.Action)
	var ose *pluginapi.EditorSupportOnlyOneKeybindingPerActionError
	require.ErrorAs(t, sk.Error, &ose)
	require.NotNil(t, ose)
}

func TestMarker_ActionLevelSkipAppliesToAllUnexported(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.multi", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Export only first, mark action-level skip
	note := "not available on this editor"
	kb, err := keybinding.NewKeybinding("cmd+a", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	m.MarkExported(a.Name, kb)
	m.MarkSkippedForReason(a.Name, nil, &pluginapi.UnsupportedExportActionError{Note: note})

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, a.Name, sk.Action)
	require.ErrorContains(t, sk.Error, note)
}

func TestMarker_PerKeyReasonOverridesActionLevel(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.override", "cmd+x", "cmd+b")
	m := exp.NewMarker(&km)

	// No exported; set action-level reason and a different per-key reason for second
	m.MarkSkippedForReason(a.Name, nil, &pluginapi.UnsupportedExportActionError{Note: "action"})
	kb, err := keybinding.NewKeybinding("cmd+b", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	m.MarkSkippedForReason(
		a.Name,
		&kb,
		&pluginapi.EditorSupportOnlyOneKeybindingPerActionError{},
	)

	rep := m.Report()
	// Both keybindings should be reported as skipped
	require.Len(t, rep.SkipActions, 2)
	// Identify entries by error type and its embedded keybinding
	var first, second *pluginapi.ExportSkipAction
	for i := range rep.SkipActions {
		var ose *pluginapi.EditorSupportOnlyOneKeybindingPerActionError
		if errors.As(rep.SkipActions[i].Error, &ose) {
			second = &rep.SkipActions[i]
		} else {
			first = &rep.SkipActions[i]
		}
	}
	require.NotNil(t, first)
	require.NotNil(t, second)
	// First uses action-level reason
	require.ErrorContains(t, first.Error, "action")
	// Second uses per-key reason
	var ose2 *pluginapi.EditorSupportOnlyOneKeybindingPerActionError
	require.ErrorAs(t, second.Error, &ose2)
	assert.NotNil(t, ose2)
}

// Tests for ExportedReport

func TestMarker_ExportedReport_AllExported(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.dual", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Export both bindings
	m.MarkExported(a.Name, a.Bindings[0])
	m.MarkExported(a.Name, a.Bindings[1])

	rep := m.ExportedReport()
	require.Len(t, rep.Actions, 1)
	result := rep.Actions[0]
	assert.Equal(t, a.Name, result.Action)
	assert.Len(t, result.Requested, 2)
	assert.Len(t, result.Exported, 2)
	assert.Empty(t, result.Reason)
}

func TestMarker_ExportedReport_PartialExport(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.partial", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Export only first binding
	m.MarkExported(a.Name, a.Bindings[0])

	rep := m.ExportedReport()
	require.Len(t, rep.Actions, 1)
	result := rep.Actions[0]
	assert.Equal(t, a.Name, result.Action)
	assert.Len(t, result.Requested, 2)
	assert.Len(t, result.Exported, 1)
	// No explicit reason set, so Reason should be empty
	assert.Empty(t, result.Reason)
}

func TestMarker_ExportedReport_PartialExportWithReason(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.partial", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Export only first binding, mark second as skipped with reason
	m.MarkExported(a.Name, a.Bindings[0])
	kb, err := keybinding.NewKeybinding("cmd+b", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	m.MarkSkippedForReason(a.Name, &kb, &pluginapi.EditorSupportOnlyOneKeybindingPerActionError{})

	rep := m.ExportedReport()
	require.Len(t, rep.Actions, 1)
	result := rep.Actions[0]
	assert.Equal(t, a.Name, result.Action)
	assert.Len(t, result.Requested, 2)
	assert.Len(t, result.Exported, 1)
	assert.Contains(t, result.Reason, "only one keybinding")
}

func TestMarker_ExportedReport_NoneExported(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.none", "cmd+a", "cmd+b")
	m := exp.NewMarker(&km)

	// Mark action as not supported (no exports)
	m.MarkSkippedForReason(a.Name, nil, pluginapi.ErrActionNotSupported)

	rep := m.ExportedReport()
	require.Len(t, rep.Actions, 1)
	result := rep.Actions[0]
	assert.Equal(t, a.Name, result.Action)
	assert.Len(t, result.Requested, 2)
	assert.Empty(t, result.Exported)
	assert.Contains(t, result.Reason, "not supported")
}

func TestMarker_ExportedReport_MultipleActions(t *testing.T) {
	kb1, err := keybinding.NewKeybinding("cmd+a", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	kb2, err := keybinding.NewKeybinding("cmd+b", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)
	kb3, err := keybinding.NewKeybinding("cmd+c", keybinding.ParseOption{Separator: "+"})
	require.NoError(t, err)

	a1 := keymap.Action{Name: "actions.test.first", Bindings: []keybinding.Keybinding{kb1}}
	a2 := keymap.Action{Name: "actions.test.second", Bindings: []keybinding.Keybinding{kb2, kb3}}
	km := keymap.Keymap{Actions: []keymap.Action{a1, a2}}

	m := exp.NewMarker(&km)

	// Export all of first action, only first binding of second action
	m.MarkExported(a1.Name, a1.Bindings[0])
	m.MarkExported(a2.Name, a2.Bindings[0])

	rep := m.ExportedReport()
	require.Len(t, rep.Actions, 2)

	// Results are sorted by action name
	assert.Equal(t, "actions.test.first", rep.Actions[0].Action)
	assert.Len(t, rep.Actions[0].Requested, 1)
	assert.Len(t, rep.Actions[0].Exported, 1)

	assert.Equal(t, "actions.test.second", rep.Actions[1].Action)
	assert.Len(t, rep.Actions[1].Requested, 2)
	assert.Len(t, rep.Actions[1].Exported, 1)
}
