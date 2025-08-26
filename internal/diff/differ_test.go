package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonDiffer_NoChange_ReturnsEmpty(t *testing.T) {
	d := NewJsonDiffer()

	before := map[string]any{"a": 1, "b": map[string]any{"c": 2}}
	after := map[string]any{"a": 1, "b": map[string]any{"c": 2}}

	ds, err := d.Diff(before, after)
	require.NoError(t, err)
	assert.Equal(t, "", ds)
}

func TestJsonDiffer_MapChanged_ReturnsDiff(t *testing.T) {
	d := NewJsonDiffer()

	before := map[string]any{"a": 1}
	after := map[string]any{"a": 2}

	got, err := d.Diff(before, after)
	require.NoError(t, err)
	// verify content contains removal and addition of key a
	assert.Contains(t, got, "\n-  \"a\": 1", "should show removal of old value")
	assert.Contains(t, got, "\n+  \"a\": 2", "should show addition of new value")
}

func TestJsonDiffer_ArrayChanged_ReturnsDiff(t *testing.T) {
	d := NewJsonDiffer()

	var before []any
	after := []any{"x"}

	got, err := d.Diff(before, after)
	require.NoError(t, err)
	// ShowArrayIndex=true should include the index and an added element line
	assert.Contains(t, got, "\n+  0: \"x\"", "should show added element at index 0")
}

func TestJsonDiffer_TypeMismatch_ReturnsError(t *testing.T) {
	d := NewJsonDiffer()

	before := map[string]any{"a": 1}
	after := []any{"x"}

	_, err := d.Diff(before, after)
	assert.Error(t, err)
}
