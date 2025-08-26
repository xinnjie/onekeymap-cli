package chore

import _ "embed"

var (
	//go:embed vscode-intellij.json
	VscodeIntellijKeymapJson []byte

	//go:embed vscode-mac-default.json
	VscodeMacDefaultKeymapJson []byte
)
