package views

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

const (
	keyCtrlC = "ctrl+c"
	keyEsc   = "esc"
	keyQ     = "q"
)

type editorSelectorOption struct {
	displayName string
	editorType  string
	installed   bool
}

// buildEditorSelectOptions creates a sorted list of editor options with installed editors first.
func buildEditorSelectOptions(options []editorSelectorOption) []huh.Option[string] {
	installed := make([]huh.Option[string], 0)
	uninstalled := make([]huh.Option[string], 0)

	for _, opt := range options {
		label := opt.displayName
		if !opt.installed {
			label = fmt.Sprintf("%s (uninstalled)", opt.displayName)
		}
		huhOpt := huh.NewOption(label, opt.editorType)
		if opt.installed {
			installed = append(installed, huhOpt)
		} else {
			uninstalled = append(uninstalled, huhOpt)
		}
	}

	finalOpts := make([]huh.Option[string], 0, len(installed)+len(uninstalled))
	finalOpts = append(finalOpts, installed...)
	finalOpts = append(finalOpts, uninstalled...)

	return finalOpts
}
