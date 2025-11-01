package keybindinglookup_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/internal/keybindinglookup"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
)

// Mock KeybindingLookup implementation for testing
type mockKeybindingLookup struct {
	name string
}

func (m *mockKeybindingLookup) Lookup(_ io.Reader, _ *keymap.KeyBinding) ([]string, error) {
	return []string{"mock-result-" + m.name}, nil
}

func newMockLookup(name string) func() keybindinglookup.KeybindingLookup {
	return func() keybindinglookup.KeybindingLookup {
		return &mockKeybindingLookup{name: name}
	}
}

func TestLookupFactory(t *testing.T) {
	factory := keybindinglookup.NewLookupFactory()

	// Test registration
	factory.Register("mock1", newMockLookup("mock1"))
	factory.Register("Mock2", newMockLookup("mock2")) // Test case insensitivity

	// Test CreateLookup with registered editors
	lookup1, err := factory.CreateLookup("mock1")
	require.NoError(t, err)
	assert.NotNil(t, lookup1)

	lookup2, err := factory.CreateLookup("MOCK2") // Test case insensitivity
	require.NoError(t, err)
	assert.NotNil(t, lookup2)

	// Test CreateLookup with unregistered editor
	_, err = factory.CreateLookup("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported editor: nonexistent")
	assert.Contains(t, err.Error(), "Supported editors:")

	// Test IsSupported
	assert.True(t, factory.IsSupported("mock1"))
	assert.True(t, factory.IsSupported("MOCK1")) // Case insensitive
	assert.True(t, factory.IsSupported("mock2"))
	assert.False(t, factory.IsSupported("nonexistent"))

	// Test GetSupportedEditors
	supported := factory.GetSupportedEditors()
	assert.Contains(t, supported, "mock1")
	assert.Contains(t, supported, "mock2")
}

func TestLookupFactory_EmptyFactory(t *testing.T) {
	factory := keybindinglookup.NewLookupFactory()

	// Test with empty factory
	_, err := factory.CreateLookup("any")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported editor: any")

	// Test GetSupportedEditors with empty factory
	supported := factory.GetSupportedEditors()
	assert.Empty(t, supported)

	// Test IsSupported with empty factory
	assert.False(t, factory.IsSupported("any"))
}

func TestLookupFactory_Integration(t *testing.T) {
	factory := keybindinglookup.NewLookupFactory()
	factory.Register("test-editor", newMockLookup("test"))

	// Create lookup and test it works
	lookup, err := factory.CreateLookup("test-editor")
	require.NoError(t, err)

	// Mock a keybinding (we don't need a real one for this test)
	keybinding, err := keymap.ParseKeyBinding("cmd+k", "+")
	require.NoError(t, err)

	// Test the lookup functionality (using nil reader since mock doesn't use it)
	results, err := lookup.Lookup(nil, keybinding)
	require.NoError(t, err)
	assert.Equal(t, []string{"mock-result-test"}, results)
}
