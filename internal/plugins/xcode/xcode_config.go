package xcode

// xcodeKeybindingsPlist represents the structure of Xcode .idekeybindings plist file
type xcodeKeybindingsPlist struct {
	MenuKeyBindings menuKeyBindings `plist:"Menu Key Bindings"`
	TextKeyBindings textKeyBindings `plist:"Text Key Bindings"`
}

type menuKeyBindings struct {
	KeyBindings []xcodeKeybinding `plist:"Key Bindings"`
}

type textKeyBindings struct {
	KeyBindings map[string]xcodeTextAction `plist:"Key Bindings"`
}

// xcodeTextAction represents a text action value in Text Key Bindings.
// It can be either a single action (string) or multiple actions ([]string).
type xcodeTextAction interface{}

type xcodeKeybindingConfig = []xcodeKeybinding
type xcodeTextKeybinding = map[string]xcodeTextAction

// xcodeKeybinding represents a single keybinding in Xcode's .idekeybindings file.
type xcodeKeybinding struct {
	// Action is the Xcode action name (e.g., "moveWordLeft:", "selectWord:")
	Action string `plist:"Action"            json:"Action"`
	// CommandID is the Xcode command ID for menu bindings
	CommandID string `plist:"CommandID"         json:"CommandID"`
	// KeyboardShortcut is the key binding string, e.g. "@k", "^g"
	KeyboardShortcut string `plist:"Keyboard Shortcut" json:"KeyboardShortcut"`
	// Title is the display title of the action
	Title string `plist:"Title"             json:"Title"`
	// Alternate indicates if this is an alternate key binding
	Alternate string `plist:"Alternate"         json:"Alternate"`
	// Group is the menu group this action belongs to
	Group string `plist:"Group"             json:"Group"`
	// GroupID is the menu group ID
	GroupID string `plist:"GroupID"           json:"GroupID"`
	// GroupedAlternate indicates if this is a grouped alternate key binding
	GroupedAlternate string `plist:"GroupedAlternate"  json:"GroupedAlternate"`
	// Navigation indicates if this is a navigation action
	Navigation string `plist:"Navigation"        json:"Navigation"`
	// ParentTitle is the parent menu title
	ParentTitle string `plist:"Parent Title"      json:"ParentTitle"`
	// CommandGroupID is the command group ID
	CommandGroupID string `plist:"CommandGroupID"    json:"CommandGroupID"`
}
