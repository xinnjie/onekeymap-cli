package onekeymap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

func TestDedupKeyBindings_DuplicatesByActionAndChords(t *testing.T) {
	kb1 := keymap.NewActioinBinding("actions.copy", "k")
	kb2 := keymap.NewActioinBinding("actions.copy", "k")
	in := []*keymapv1.ActionBinding{kb1, kb2}

	out := dedupKeyBindings(in)

	assert.Len(t, out, 1, "unexpected length")
	assert.Same(t, kb1, out[0], "expected first occurrence to be kept")
}

func TestDedupKeyBindings_KeepsOrderAndSkipsNil(t *testing.T) {
	kb1 := keymap.NewActioinBinding("actions.open", "o")
	var kbNil *keymapv1.ActionBinding
	kb3 := keymap.NewActioinBinding("actions.open", "o") // duplicate of kb1
	kb4 := keymap.NewActioinBinding("actions.save", "o") // same chords, different action

	in := []*keymapv1.ActionBinding{kb1, kbNil, kb3, kb4}
	out := dedupKeyBindings(in)

	assert.Len(t, out, 2, "unexpected length")
	assert.Same(t, kb1, out[0], "first element mismatch")
	assert.Same(t, kb4, out[1], "second element mismatch")
}

func TestDedupKeyBindings_DifferentActionsOrChordsNotDedup(t *testing.T) {
	kb1 := keymap.NewActioinBinding("actions.find", "f")
	kb2 := keymap.NewActioinBinding("actions.replace", "f") // different action
	kb3 := keymap.NewActioinBinding("actions.find", "g")    // different chords

	in := []*keymapv1.ActionBinding{kb1, kb2, kb3}
	out := dedupKeyBindings(in)

	assert.Len(t, out, 3, "unexpected length")
	assert.Same(t, kb1, out[0], "order changed")
	assert.Same(t, kb2, out[1], "order changed")
	assert.Same(t, kb3, out[2], "order changed")
}

func TestDedupKeyBindings_NilVsEmptyChordsConsideredEqual(t *testing.T) {
	kb1 := &keymapv1.ActionBinding{Id: "actions.nochord"}
	kb2 := &keymapv1.ActionBinding{Id: "actions.nochord", Bindings: []*keymapv1.Binding{}}

	in := []*keymapv1.ActionBinding{kb1, kb2}
	out := dedupKeyBindings(in)

	assert.Len(t, out, 1, "unexpected length")
	assert.Same(t, kb1, out[0], "expected first occurrence to be kept when nil vs empty chords")
}
