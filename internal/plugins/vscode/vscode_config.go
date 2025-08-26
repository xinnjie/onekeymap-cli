package vscode

import (
	"encoding/json"
	"log/slog"
)

type vscodeKeybindingConfig = []vscodeKeybinding

// vscodeKeybinding represents a single keybinding in VSCode's keybindings.json.
type vscodeKeybinding struct {
	Key     string     `json:"key"`
	Command string     `json:"command"`
	When    string     `json:"when,omitempty"`
	Args    vscodeArgs `json:"args,omitempty"`
}

type vscodeArgs map[string]any

func (a vscodeArgs) LogValue() slog.Value {
	b, err := json.Marshal(a)
	if err != nil {
		return slog.StringValue(err.Error())
	}
	return slog.StringValue(string(b))
}
