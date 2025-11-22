package validateapi

// ValidationReport is the overall report of a validation run.
type ValidationReport struct {
	// The source editor of the validation run.
	SourceEditor string
	// The summary of the validation run.
	Summary Summary
	// The issues detected during the validation run.
	Issues []ValidationIssue
	// The warnings issued during the validation run.
	Warnings []ValidationIssue
}

// Summary is the summary of a validation run.
type Summary struct {
	// The total number of mappings processed.
	MappingsProcessed int
	// The number of mappings that succeeded.
	MappingsSucceeded int
}

// ValidationIssue is a single issue detected during validation.
type ValidationIssue struct {
	// Type indicates the kind of issue
	Type IssueType
	// Details contains the issue-specific data
	Details IssueDetails
}

// IssueType represents the type of validation issue.
type IssueType string

const (
	IssueTypeKeybindConflict    IssueType = "keybind_conflict"
	IssueTypeDanglingAction     IssueType = "dangling_action"
	IssueTypeUnsupportedAction  IssueType = "unsupported_action"
	IssueTypeDuplicateMapping   IssueType = "duplicate_mapping"
	IssueTypePotentialShadowing IssueType = "potential_shadowing"
)

// IssueDetails holds the details for different issue types.
type IssueDetails interface {
	issueDetails()
}

// KeybindConflict is a keybinding conflict detected during validation.
type KeybindConflict struct {
	// The keybinding that is in conflict.
	Keybinding string
	// The actions that are in conflict.
	Actions []ConflictAction
}

func (KeybindConflict) issueDetails() {}

// ConflictAction represents an action involved in a keybinding conflict.
type ConflictAction struct {
	// The action ID.
	ActionID string
	// The context or condition for the action (optional).
	Context string
}

// DanglingAction is a dangling action detected during validation.
type DanglingAction struct {
	// The action that is dangling.
	Action string
	// The target editor where the action is not found.
	TargetEditor string
	// A suggestion for fixing the issue.
	Suggestion string
}

func (DanglingAction) issueDetails() {}

// UnsupportedAction is an unsupported action detected during validation.
type UnsupportedAction struct {
	// The action that is unsupported.
	Action string
	// The keybinding for the unsupported action.
	Keybinding string
	// The target editor that does not support the action.
	TargetEditor string
}

func (UnsupportedAction) issueDetails() {}

// DuplicateMapping is a duplicate mapping detected during validation.
type DuplicateMapping struct {
	// The action that has duplicate mappings.
	Action string
	// The keybinding that is duplicated.
	Keybinding string
}

func (DuplicateMapping) issueDetails() {}

// PotentialShadowing is a potential shadowing issue detected during validation.
type PotentialShadowing struct {
	// The keybinding that might shadow a critical shortcut.
	Keybinding string
	// The action being mapped.
	Action string
	// Description of the critical shortcut being shadowed.
	CriticalShortcutDescription string
}

func (PotentialShadowing) issueDetails() {}
