package imports_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/imports"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

func TestMarker_ReportEmptyWhenNoSkips(t *testing.T) {
	m := imports.NewMarker()
	rep := m.Report()
	require.Empty(t, rep.SkipActions)
}

func TestMarker_RecordSkipWithReason(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkippedForReason("editor.action.copy", errors.New("some reason"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, "editor.action.copy", sk.EditorSpecificAction)
	assert.ErrorContains(t, sk.Error, "some reason")
}

func TestMarker_DefaultErrorWhenReasonNil(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkippedForReason("editor.action.paste", nil)

	rep := m.Report()
	require.Len(t, rep.SkipActions, 1)
	sk := rep.SkipActions[0]
	assert.Equal(t, "editor.action.paste", sk.EditorSpecificAction)
	assert.ErrorIs(t, sk.Error, pluginapi.ErrNotSupported)
}

func TestMarker_ImportedOverridesSkip(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkippedForReason("editor.action.cut", errors.New("temporary"))
	m.MarkImported("editor.action.cut")

	rep := m.Report()
	require.Empty(t, rep.SkipActions)
}

func TestMarker_DeduplicateSkipsPreserveOrder(t *testing.T) {
	m := imports.NewMarker()
	m.MarkSkippedForReason("editor.action.copy", errors.New("first"))
	m.MarkSkippedForReason("editor.action.paste", errors.New("second"))
	m.MarkSkippedForReason("editor.action.copy", errors.New("override"))

	rep := m.Report()
	require.Len(t, rep.SkipActions, 2)
	assert.Equal(t, "editor.action.copy", rep.SkipActions[0].EditorSpecificAction)
	assert.Equal(t, "editor.action.paste", rep.SkipActions[1].EditorSpecificAction)
	assert.ErrorContains(t, rep.SkipActions[0].Error, "first")
}

func TestMarker_ImportPreventsLaterSkip(t *testing.T) {
	m := imports.NewMarker()
	m.MarkImported("editor.action.copy")
	m.MarkSkippedForReason("editor.action.copy", errors.New("should be ignored"))

	rep := m.Report()
	require.Empty(t, rep.SkipActions)
}
