package export_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	exp "github.com/xinnjie/onekeymap-cli/internal/export"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

func newKeymapWithTwoBindings(action string, k1, k2 string) (*keymapv1.Keymap, *keymapv1.Action) {
	a := keymap.NewActioinBinding(action, k1)
	a.Bindings = append(a.Bindings, &keymapv1.KeybindingReadable{KeyChords: keymap.MustParseKeyBinding(k2).KeyChords})
	return &keymapv1.Keymap{Actions: []*keymapv1.Action{a}}, a
}

func TestMarker_ImplicitSkipForUnexportedBindings(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.dual", "cmd+a", "cmd+c")
	m := exp.NewMarker(km)

	// Export only first binding
	m.MarkExported(a.GetName(), a.GetBindings()[0].GetKeyChords())

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	// The skipped one should be the second binding with default not supported
	sk := rep.SkipActions[0]
	assert.Equal(t, a.GetName(), sk.Action)
	assert.ErrorIs(t, sk.Error, pluginapi.ErrActionNotSupported)
}

func TestMarker_ExplicitKeySkip(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.dual", "cmd+a", "cmd+b")
	m := exp.NewMarker(km)

	// Export first, explicitly skip second
	m.MarkExported(a.GetName(), a.GetBindings()[0].GetKeyChords())
	m.MarkSkippedForReason(
		a.GetName(),
		keymap.MustParseKeyBinding("cmd+b").KeyChords,
		&pluginapi.EditorSupportOnlyOneKeybindingPerActionError{
			SkipKeybinding: keymap.MustParseKeyBinding("cmd+b").KeyChords,
		},
	)

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, a.GetName(), sk.Action)
	var ose *pluginapi.EditorSupportOnlyOneKeybindingPerActionError
	require.ErrorAs(t, sk.Error, &ose)
	require.NotNil(t, ose)
	require.NotNil(t, ose.SkipKeybinding)
	assert.Equal(t, ose.SkipKeybinding, keymap.MustParseKeyBinding("cmd+b").KeyChords)
}

func TestMarker_ActionLevelSkipAppliesToAllUnexported(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.multi", "cmd+a", "cmd+b")
	m := exp.NewMarker(km)

	// Export only first, mark action-level skip
	note := "not available on this editor"
	m.MarkExported(a.GetName(), keymap.MustParseKeyBinding("cmd+a").KeyChords)
	m.MarkSkippedForReason(a.GetName(), nil, &pluginapi.NotSupportedError{Note: note})

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, a.GetName(), sk.Action)
	require.ErrorContains(t, sk.Error, note)
}

func TestMarker_PerKeyReasonOverridesActionLevel(t *testing.T) {
	km, a := newKeymapWithTwoBindings("actions.test.override", "cmd+x", "cmd+b")
	m := exp.NewMarker(km)

	// No exported; set action-level reason and a different per-key reason for second
	m.MarkSkippedForReason(a.GetName(), nil, &pluginapi.NotSupportedError{Note: "action"})
	m.MarkSkippedForReason(
		a.GetName(),
		keymap.MustParseKeyBinding("cmd+b").KeyChords,
		&pluginapi.EditorSupportOnlyOneKeybindingPerActionError{
			SkipKeybinding: keymap.MustParseKeyBinding("cmd+b").KeyChords,
		},
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
	assert.Equal(t, ose2.SkipKeybinding, keymap.MustParseKeyBinding("cmd+b").KeyChords)
}
