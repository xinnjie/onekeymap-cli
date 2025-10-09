package pluginapi

type EditorType string

const (
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
