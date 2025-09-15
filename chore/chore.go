package chore

import _ "embed"

var (
	//go:embed vscode-intellij.json
	VscodeIntellijKeymapJSON []byte

	//go:embed vscode-mac-default.json
	VscodeMacDefaultKeymapJSON []byte
)
