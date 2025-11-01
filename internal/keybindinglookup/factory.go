package keybindinglookup

import (
	"fmt"
	"strings"
)

// LookupFactory manages KeybindingLookup implementations for different editors
type LookupFactory struct {
	lookups map[string]func() KeybindingLookup
}

// NewLookupFactory creates a new factory instance
func NewLookupFactory() *LookupFactory {
	return &LookupFactory{
		lookups: make(map[string]func() KeybindingLookup),
	}
}

// Register registers a KeybindingLookup constructor for a specific editor
func (f *LookupFactory) Register(editorName string, constructor func() KeybindingLookup) {
	f.lookups[strings.ToLower(editorName)] = constructor
}

// CreateLookup creates a KeybindingLookup instance for the specified editor
func (f *LookupFactory) CreateLookup(editorName string) (KeybindingLookup, error) {
	constructor, exists := f.lookups[strings.ToLower(editorName)]
	if !exists {
		return nil, fmt.Errorf("unsupported editor: %s. Supported editors: %s",
			editorName, f.GetSupportedEditors())
	}
	return constructor(), nil
}

// GetSupportedEditors returns a comma-separated list of supported editor names
func (f *LookupFactory) GetSupportedEditors() string {
	var editors []string
	for editor := range f.lookups {
		editors = append(editors, editor)
	}
	return strings.Join(editors, ", ")
}

// IsSupported checks if an editor is supported
func (f *LookupFactory) IsSupported(editorName string) bool {
	_, exists := f.lookups[strings.ToLower(editorName)]
	return exists
}
