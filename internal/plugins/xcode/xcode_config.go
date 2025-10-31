package xcode

// xcodeKeybindingsPlist represents the structure of Xcode .idekeybindings plist file
type xcodeKeybindingsPlist struct {
	MenuKeyBindings menuKeyBindings `plist:"Menu Key Bindings"`
}

type menuKeyBindings struct {
	KeyBindings []xcodeKeybinding `plist:"Key Bindings"`
}

type xcodeKeybindingConfig = []xcodeKeybinding

// xcodeKeybinding represents a single keybinding in Xcode's .idekeybindings file.
type xcodeKeybinding struct {
	// Action is the Xcode action name (e.g., "moveWordLeft:", "selectWord:")
	Action string `plist:"Action"`
	// CommandID is the Xcode command ID for menu bindings
	CommandID string `plist:"CommandID"`
	// KeyboardShortcut is the key binding string, e.g. "@k", "^g"
	KeyboardShortcut string `plist:"Keyboard Shortcut"`
	// Title is the display title of the action
	Title string `plist:"Title"`
	// Alternate indicates if this is an alternate key binding
	Alternate string `plist:"Alternate"`
	// Group is the menu group this action belongs to
	Group string `plist:"Group"`
	// GroupID is the menu group ID
	GroupID string `plist:"GroupID"`
	// GroupedAlternate indicates if this is a grouped alternate key binding
	GroupedAlternate string `plist:"GroupedAlternate"`
	// Navigation indicates if this is a navigation action
	Navigation string `plist:"Navigation"`
	// ParentTitle is the parent menu title
	ParentTitle string `plist:"Parent Title"`
	// CommandGroupID is the command group ID
	CommandGroupID string `plist:"CommandGroupID"`
}
