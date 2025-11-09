package xcode

// xcodeKeybindingsPlist represents the structure of Xcode .idekeybindings plist file
type xcodeKeybindingsPlist struct {
	MenuKeyBindings menuKeyBindings `plist:"Menu Key Bindings"`
	TextKeyBindings textKeyBindings `plist:"Text Key Bindings"`
}

type menuKeyBindings struct {
	KeyBindings []xcodeKeybinding `plist:"Key Bindings"`
	Version     int               `plist:"Version"`
}

type textKeyBindings struct {
	KeyBindings map[string]*textActionValue `plist:"Key Bindings"`
	Version     int                         `plist:"Version"`
}

// textActionValue represents a text action value in Text Key Bindings.
// It supports both a single action (string) and multiple actions ([]string)
// via custom plist marshal/unmarshal.
type textActionValue struct {
	Items []string
}

// MarshalPlist implements howett.net/plist Marshaler.
func (v *textActionValue) MarshalPlist() (interface{}, error) {
	switch len(v.Items) {
	case 0:
		// Should not normally be emitted; return empty string to avoid nil
		return "", nil
	case 1:
		return v.Items[0], nil
	default:
		return v.Items, nil
	}
}

// UnmarshalPlist implements howett.net/plist Unmarshaler.
// It accepts either a single string or an array of strings.
func (v *textActionValue) UnmarshalPlist(unmarshal func(interface{}) error) error {
	// Try string first
	var s string
	if err := unmarshal(&s); err == nil {
		v.Items = []string{s}
		return nil
	}
	// Then try []string
	var arr []string
	if err := unmarshal(&arr); err == nil {
		v.Items = arr
		return nil
	}
	// Finally, support dict wrapper with Items key
	var wrapper struct {
		Items []string `plist:"Items"`
	}
	if err := unmarshal(&wrapper); err == nil && len(wrapper.Items) > 0 {
		v.Items = wrapper.Items
		return nil
	}
	// Fallback: empty
	v.Items = nil
	return nil
}

type xcodeKeybindingConfig = []xcodeKeybinding
type xcodeTextKeybinding = map[string]*textActionValue

// xcodeKeybinding represents a single keybinding in Xcode's .idekeybindings file.
type xcodeKeybinding struct {
	// Action is the Xcode action name (e.g., "moveWordLeft:", "selectWord:")
	Action string `plist:"Action"            json:"Action"`
	// Alternate indicates if this is an alternate key binding
	Alternate string `plist:"Alternate"         json:"Alternate"`
	// CommandGroupID is the command group ID
	CommandGroupID string `plist:"CommandGroupID"    json:"CommandGroupID"`
	// CommandID is the Xcode command ID for menu bindings
	CommandID string `plist:"CommandID"         json:"CommandID"`
	// Group is the menu group this action belongs to
	Group string `plist:"Group"             json:"Group"`
	// GroupID is the menu group ID
	GroupID string `plist:"GroupID"           json:"GroupID"`
	// GroupedAlternate indicates if this is a grouped alternate key binding
	GroupedAlternate string `plist:"GroupedAlternate"  json:"GroupedAlternate"`
	// KeyboardShortcut is the key binding string, e.g. "@k", "^g"
	KeyboardShortcut string `plist:"Keyboard Shortcut" json:"KeyboardShortcut"`
	// Navigation indicates if this is a navigation action
	Navigation string `plist:"Navigation"        json:"Navigation"`
	// ParentTitle is the parent menu title
	ParentTitle string `plist:"Parent Title"      json:"ParentTitle"`
	// Title is the display title of the action
	Title string `plist:"Title"             json:"Title"`
}
