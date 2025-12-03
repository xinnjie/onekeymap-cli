# Non-destructive Export

## Overview

Non-destructive export is a core feature of OneKeymap that allows exporting keymaps to editors while **preserving user's existing custom keybindings**. Instead of completely overwriting the target editor's keymap configuration, OneKeymap intelligently merges managed keybindings with existing user customizations.

## Design Philosophy

Traditional keymap export tools often replace the entire configuration file, causing users to lose their custom keybindings. OneKeymap takes a different approach:

1. **Preserve user customizations** - Existing keybindings that are not managed by OneKeymap are kept intact
2. **Managed keybindings take priority** - When conflicts occur, OneKeymap-managed keybindings override user keybindings
3. **Non-key sections are preserved** - Editor-specific settings (themes, editor options, etc.) remain unchanged

## How It Works

When exporting to an editor, OneKeymap:

1. Reads the existing configuration file (if provided via `ExistingConfig`)
2. Identifies which keybindings are "managed" (defined in the OneKeymap keymap setting)
3. Identifies which keybindings are "unmanaged" (user's custom keybindings)
4. Merges them together, with managed keybindings taking priority on conflicts
5. Writes the merged configuration

### Conflict Resolution

When a managed keybinding conflicts with an existing user keybinding (same key combination), the managed keybinding **always wins**. This ensures consistent behavior across editors.

## Supported Editors

Non-destructive export is supported across all major editor plugins:

### VS Code

- **Format**: JSON array in `keybindings.json`
- **Behavior**: Preserves user keybindings, handles trailing commas and comments (JSONC format)
- **Conflict key**: `key` + `command` combination

```json
// Existing user config
[
  { "key": "cmd+x", "command": "custom.user.command" }
]

// After export with actions.edit.copy -> cmd+c
[
  { "key": "cmd+c", "command": "editor.action.clipboardCopyAction", "when": "editorTextFocus" },
  { "key": "cmd+x", "command": "custom.user.command" }
]
```

### IntelliJ IDEA

- **Format**: XML keymap file
- **Behavior**: Preserves user-defined actions, managed actions override conflicting ones
- **Conflict key**: Action ID

```xml
<!-- Existing user config -->
<keymap name="Onekeymap" parent="$default">
  <action id="CustomUserAction">
    <keyboard-shortcut first-keystroke="meta X" />
  </action>
</keymap>

<!-- After export: CustomUserAction preserved, $Copy added -->
<keymap name="Onekeymap" parent="$default">
  <action id="$Copy">
    <keyboard-shortcut first-keystroke="meta C" />
  </action>
  <action id="CustomUserAction">
    <keyboard-shortcut first-keystroke="meta X" />
  </action>
</keymap>
```

### Zed

- **Format**: JSON array with context-based bindings
- **Behavior**: Preserves user contexts and bindings, merges within same context
- **Conflict key**: Context + key combination

```json
// Existing user config
[
  { "context": "Editor", "bindings": { "cmd-x": "custom::UserAction" } },
  { "context": "Workspace", "bindings": { "cmd-shift-p": "custom::WorkspaceAction" } }
]

// After export with actions.edit.copy -> cmd+c
[
  { "context": "Editor", "bindings": { "cmd-c": "editor::Copy", "cmd-x": "custom::UserAction" } },
  { "context": "Workspace", "bindings": { "cmd-shift-p": "custom::WorkspaceAction" } }
]
```

### Helix

- **Format**: TOML configuration
- **Behavior**: Preserves all non-key sections (theme, editor settings), merges key bindings by mode
- **Conflict key**: Mode + key combination

```toml
# Existing user config
theme = "onedark"

[editor]
line-number = "relative"

[keys.normal]
"C-x" = "custom_user_command"

# After export with actions.edit.copy -> ctrl+c
# theme and editor sections preserved, keys merged
theme = "onedark"

[editor]
line-number = "relative"

[keys.normal]
"C-x" = "custom_user_command"

[keys.insert]
"C-c" = "yank"
```

### Xcode

- **Format**: Property List (plist) with Menu Key Bindings and Text Key Bindings
- **Behavior**: Preserves unmanaged menu and text bindings, merges by action/shortcut
- **Conflict key**: Action identifier or keyboard shortcut

```xml
<!-- Existing user config -->
<dict>
  <key>Menu Key Bindings</key>
  <dict>
    <key>Key Bindings</key>
    <array>
      <dict>
        <key>Action</key><string>customUserAction:</string>
        <key>Keyboard Shortcut</key><string>@k</string>
      </dict>
    </array>
  </dict>
</dict>

<!-- After export: customUserAction preserved, managed action added -->
```

## Edge Cases

### JSONC Support (VS Code, Zed)

Both VS Code and Zed support JSONC format (JSON with Comments). OneKeymap correctly parses configs with:
- Single-line comments (`//`)
- Trailing commas

The output is always valid JSON (comments are not preserved in output).

### Order Preservation

Some editors (VS Code, Xcode) preserve the order of keybindings based on the existing config. Managed keybindings are inserted at appropriate positions to maintain consistency.
