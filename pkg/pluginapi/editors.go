package pluginapi

import (
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type EditorType string

func NewEditorTypeFromAPI(e keymapv1.EditorType) EditorType {
	switch e {
	case keymapv1.EditorType_VSCODE:
		return EditorTypeVSCode
	case keymapv1.EditorType_WINDSURF:
		return EditorTypeWindsurf
	case keymapv1.EditorType_WINDSURF_NEXT:
		return EditorTypeWindsurfNext
	case keymapv1.EditorType_CURSOR:
		return EditorTypeCursor
	case keymapv1.EditorType_INTELLIJ:
		return EditorTypeIntelliJ
	case keymapv1.EditorType_INTELLIJ_COMMUNITY:
		return EditorTypeIntelliJCommunity
	case keymapv1.EditorType_PYCHARM:
		return EditorTypePyCharm
	case keymapv1.EditorType_WEBSTORM:
		return EditorTypeWebStorm
	case keymapv1.EditorType_CLION:
		return EditorTypeClion
	case keymapv1.EditorType_PHPSTORM:
		return EditorTypePhpStorm
	case keymapv1.EditorType_RUBYMINE:
		return EditorTypeRubyMine
	case keymapv1.EditorType_GOLAND:
		return EditorTypeGoLand
	case keymapv1.EditorType_ZED:
		return EditorTypeZed
	case keymapv1.EditorType_VIM:
		return EditorTypeVim
	case keymapv1.EditorType_HELIX:
		return EditorTypeHelix
	default:
		return EditorTypeUnknown
	}
}

const (
	EditorTypeUnknown EditorType = "unknown"

	// VSCode series

	// EditorTypeVSCode represents Visual Studio Code editor.
	EditorTypeVSCode       EditorType = "vscode"
	EditorTypeWindsurf     EditorType = "vscode.windsurf"
	EditorTypeWindsurfNext EditorType = "vscode.windsurf-next"
	EditorTypeCursor       EditorType = "vscode.cursor"

	// IntelliJ idea series

	// EditorTypeIntelliJ represents IntelliJ IDEA Ultimate editor.
	EditorTypeIntelliJ          EditorType = "intellij"
	EditorTypeIntelliJCommunity EditorType = "intellij.intellij-community"
	EditorTypePyCharm           EditorType = "intellij.pycharm"
	EditorTypeWebStorm          EditorType = "intellij.webstorm"
	EditorTypeClion             EditorType = "intellij.clion"
	EditorTypePhpStorm          EditorType = "intellij.phpstorm"
	EditorTypeRubyMine          EditorType = "intellij.rubymine"
	EditorTypeGoLand            EditorType = "intellij.goland"

	EditorTypeZed   EditorType = "zed"
	EditorTypeVim   EditorType = "vim"
	EditorTypeHelix EditorType = "helix"
)

// AppName returns the human-readable display name for the editor type.
func (e EditorType) AppName() string {
	switch e {
	case EditorTypeVSCode:
		return "VSCode"
	case EditorTypeWindsurf:
		return "Windsurf"
	case EditorTypeWindsurfNext:
		return "Windsurf Next"
	case EditorTypeCursor:
		return "Cursor"
	case EditorTypeIntelliJ:
		return "IntelliJ IDEA Ultimate"
	case EditorTypeIntelliJCommunity:
		return "IntelliJ IDEA Community"
	case EditorTypePyCharm:
		return "PyCharm"
	case EditorTypeWebStorm:
		return "WebStorm"
	case EditorTypeClion:
		return "CLion"
	case EditorTypePhpStorm:
		return "PhpStorm"
	case EditorTypeRubyMine:
		return "RubyMine"
	case EditorTypeGoLand:
		return "GoLand"
	case EditorTypeZed:
		return "Zed"
	case EditorTypeVim:
		return "Vim"
	case EditorTypeHelix:
		return "Helix"
	default:
		return "Unknown"
	}
}
