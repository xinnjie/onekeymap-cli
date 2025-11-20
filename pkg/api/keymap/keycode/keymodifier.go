package keycode

type KeyModifier string

const (
	KeyModifierShift KeyModifier = "shift"
	KeyModifierCtrl  KeyModifier = "ctrl"
	KeyModifierAlt   KeyModifier = "alt"
	// KeyModifierMeta is Command(⌘) key on macOS, Windows(⊞) key on Windows, super key on Linux
	KeyModifierMeta KeyModifier = "meta"
)

type MetaPlatformSpecificKeyModifier string

const (
	MetaPlatformSpecificKeyModifierMac     MetaPlatformSpecificKeyModifier = "cmd"
	MetaPlatformSpecificKeyModifierLinux   MetaPlatformSpecificKeyModifier = "meta"
	MetaPlatformSpecificKeyModifierWindows MetaPlatformSpecificKeyModifier = "windows"
)
