package vscode

import (
	"encoding/json"
	"log/slog"
)

type vscodeKeybindingConfig = []vscodeKeybinding

// vscodeKeybinding represents a single keybinding in VSCode's keybindings.json.
type vscodeKeybinding struct {
	// Key is the key binding string, e.g. "ctrl+c"
	Key string `json:"key"`
	// Command is the command to execute, e.g. "editor.action.clipboardCopyAction"
	Command string `json:"command"`
	// When is the condition under which the keybinding is active, e.g. "editorTextFocus"
	When string `json:"when,omitempty"`
	// Args is the arguments to pass to the command
	Args vscodeArgs `json:"args,omitempty"`
}

type vscodeArgs map[string]any

func (a vscodeArgs) LogValue() slog.Value {
	b, err := json.Marshal(a)
	if err != nil {
		return slog.StringValue(err.Error())
	}
	return slog.StringValue(string(b))
}
