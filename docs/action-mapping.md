# OneKeymap Action Mappings

This directory contains the action mapping definitions that serve as the "brain" of the OneKeymap system. These YAML files define how actions from different editors are translated to and from the universal OneKeymap format.

## Overview

The Action Mappings System is the core component that enables OneKeymap to translate keyboard shortcuts between different editors. Each mapping file defines a set of actions with their corresponding implementations across supported editors.

## File Structure

```
action_mappings/
├── editor.yaml        # Core text editing operations
├── navigation.yaml    # Cursor movement and navigation
└── workspace.yaml     # Workspace-level operations
```

## Mapping File Format

Each YAML file follows this structure:

```yaml
mappings:
  - id: "actions.category.actionName"
    name: "A short, human-readable name for the action"
    description: "A longer, more detailed description of what the action does"
    category: "editor.editing.selection"
    featured: false
    featuredReason: "This action is widely supported and is considered portable."
    vscode:
      command: "vscode.command.name"
      when: "contextCondition"
      note: "optional special note of this command"
    zed:
      action: "zed::Action"
      context: "ContextName"
    intellij:
      action: "ActionName"
      context: "ContextName"
    vim:
      command: "vim_command"
      mode: "normal|insert|visual"
    # To nest actions, add a 'children' key with child action IDs.
    children:
      - "actions.category.actionName.child"
```

### Field Definitions

- **`id`**: Unique universal action identifier following the pattern `actions.category.actionName`.
- **`name`**: A short, human-readable name for the action.
- **`description`**: A clear, human-readable explanation of what the action does.
- **`category`**: The category of the action, used for grouping and organization.
- **`featured`**: A boolean indicating if the action is not widely portable across editors. Set to `true` for editor-specific or non-standard actions.
- **`featuredReason`**: An explanation for why an action is `featured`, or a recommendation to use a more portable alternative.
- **`children`** (optional): A list of child action IDs (string array). In the UI, child actions will be collapsed under their parent action.
### Editor-specific sections (`vscode`, `zed`, `intellij`, `vim`, `helix`, `xcode`):
  - These sections contain the specific implementation details for each editor. For editors that support multiple configurations for a single action (like VSCode), this is a list of mappings.
  - **`disableImport`** (optional): If `true`, this mapping will only be used for exporting keymaps and will be ignored during import.
  - **`notSupported`** (optional): If `true`, this action is explicitly marked as not supported for the editor.
  - **`note`** (optional): A string explaining why the action is not supported.

#### Export Fallback Mechanism (for `children`):

When exporting keybindings for a specific editor, if a parent action is marked as `notSupported` or has no definition for that editor, the system will attempt to find a suitable replacement within its `children` list. It will iterate through the `children` in the defined order and select the *first* child action that *is* supported by the target editor. This ensures a graceful fallback for editor-specific limitations.

**Key Characteristics of the Fallback Mechanism:**

1.  **Export-Only Behavior**:
    *   **Export**: The fallback logic is applied *only* during the export process. If a parent action (e.g., `actions.go.callHierarchy`) is bound to `Cmd+Shift+H` but is not supported in the target editor (e.g., VSCode), the system will look for the first supported child action (e.g., `actions.go.callHierarchy.peek`) in the `children` list and export `Cmd+Shift+H` bound to VSCode's `peek` command.
    *   **Import**: When importing keybindings, this fallback logic is *not* applied. A keybinding imported for a specific child action (e.g., `actions.go.callHierarchy.peek`) will always be mapped back to that exact child action, not its parent.

2.  **Order as Priority**:
    *   The `children` list functions as an ordered **priority list**. The exporter will traverse this list from top to bottom, selecting the *first* child action that is supported by the target editor. This allows maintainers to explicitly control the fallback preference.

3.  **Export Process Transparency **:
    *   It is recommended that the export process provides clear logging or warnings when a fallback occurs. For example: "Action `actions.go.callHierarchy` is not directly supported in VSCode; falling back to `actions.go.callHierarchy.peek`." This ensures users understand why a specific keybinding might be mapped to a different, though related, action than originally intended.

4.  **Conflict Resolution: Precise Match Wins**:
    *   If a user has defined keybindings for both a parent action and one of its child actions, the system will prioritize the most specific mapping. Keybindings explicitly set for a child action will always take precedence over any keybinding inherited from its parent or determined via the fallback mechanism.

## Supported Editors

### VSCode
- **Command**: The exact command ID from VSCode's command palette
- **When**: Optional context clause that determines when the command is available
- Example: `editor.action.clipboardCopyAction` with `when: "editorTextFocus"`

### Zed
- **Action**: The Zed action identifier (usually in format `module::Action`)
- **Context**: The context where this action is available
- Example: `editor::Copy` with `context: "Editor"`

### IntelliJ IDEA
- **Action**: The IntelliJ action ID
- **Context**: Optional context specification
- Example: `Copy` with `context: "EditorFocus"`

### Vim
- **Command**: The vim command or key sequence
- **Mode**: The vim mode where this command applies (normal, insert, visual, etc.)
- Example: `y` with `mode: "visual"`

## Adding New Mappings

To add a new action mapping:

1. **Choose the appropriate file** based on the action category
2. **Create a unique ID** following the naming convention: `actions.category.actionName`
3. **Add comprehensive mappings** for all supported editors
4. **Test the mapping** by running the CLI after rebuild

### Common Validation Errors
1. **Duplicate ID**: Each action ID must be unique across all files
2. **Invalid YAML**: Syntax errors in YAML format
3. **Missing required fields**: ID and description are mandatory

## Best Practices

### Naming Conventions
- Use descriptive, hierarchical IDs: `actions.category.specificAction`
- Keep descriptions concise but clear
- Use consistent terminology across similar actions

### Editor-Specific Guidelines

#### VSCode
- Use exact command IDs from VSCode's built-in commands
- Include appropriate `when` clauses to avoid conflicts
- Reference: [VSCode Built-in Commands](https://code.visualstudio.com/api/references/commands)

#### Zed
- Follow Zed's action naming convention: `module::Action`
- Use appropriate contexts to match Zed's behavior
- Reference: [Zed Keybinding Documentation](https://zed.dev/docs/key-bindings)

#### IntelliJ
- Use standard IntelliJ action IDs
- Reference: Find action IDs in IntelliJ via Help → Find Action → search for desired action
