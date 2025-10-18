# Contributing to OneKeymap CLI

First off, thank you for considering contributing! It's people like you that make OneKeymap CLI such a great tool.

This document provides guidelines for contributing to the project.

## How Can I Contribute?

There are many ways to contribute, from writing tutorials or blog posts, improving the documentation, submitting bug reports and feature requests, or writing code which can be incorporated into OneKeymap CLI itself.

A great place to start is by enhancing the action mappings.

## Enhancing the Action Mapping Configuration

The "Action Mapping" is the core translation layer of OneKeymap CLI. It's a knowledge base that connects editor-specific commands (like VSCode's `editor.action.clipboardCopyAction` or Zed's `editor::Copy`) to a universal, editor-agnostic action ID (like `actions.editor.copy`).

By improving these mappings, you help OneKeymap CLI support more commands across more editors, making it more powerful for everyone.

### File Structure

All action mappings are defined in YAML files located in the `config/action_mappings/` directory. The mappings are split into logical groups to keep them organized and maintainable.

```
config/action_mappings/
├── editor.yaml
├── navigation.yaml
├── view-management.yaml
└── ...
```

When the application starts, it reads all `*.yaml` files in this directory, merges them into a single collection, and validates that every `id` is unique.

### How to Add or Modify a Mapping

Each mapping entry in the YAML files follows a specific structure. Here’s a breakdown of how to define simple and complex mappings.

#### 1. Simple (One-to-One) Mappings

A simple mapping connects one universal action to one specific command in each editor.

**Example:** Mapping the universal "copy" action.

```yaml
# config/action_mappings/editor.yaml
mappings:
  - id: "actions.editor.copy"
    description: "Copies the current selection to the clipboard."
    vscode:
      command: "editor.action.clipboardCopyAction"
      when: "editorTextFocus" # Optional: context for when the command is active
    zed:
      action: "editor::Copy"
      context: "Editor && vim_mode != 'insert'" # Optional: Zed's context
    intellij:
      action: "$Copy"
```

**Key Fields:**
-   **`id`**: The unique, universal identifier for the action. This is the source of truth. Use the format `actions.<category>.<verb>`.
-   **`description`**: A human-readable explanation of what the action does.
-   **`[editor_name]`**: A key for each supported editor (e.g., `vscode`, `zed`, `intellij`).
    -   **`command` / `action`**: The editor-specific command ID. The key name (`command` or `action`) depends on the editor's terminology.
    -   **`when` / `context`**: (Optional) The context in which the keybinding is active. This is crucial for avoiding conflicts and ensuring shortcuts work as expected. This logic is handled entirely by the editor plugins during import/export.

#### 2. Complex (Stateful Toggle) Mappings

Some actions are "stateful toggles," meaning the same key does different things depending on the application's state (e.g., a key that opens a panel if it's closed, and closes it if it's open).

To handle this, you can define the editor-specific mapping as a **list of command objects** instead of a single object. Each object in the list represents a different state.

**Example:** Mapping a universal "toggle search" action.

```yaml
# config/action_mappings/view-management.yaml
mappings:
  - id: "actions.view.toggleSearch"
    description: "Toggles the visibility of the Search view."
    # For VSCode, this requires two context-dependent commands.
    vscode:
      - command: "workbench.view.search"
        when: "workbench.view.search.active && neverMatch =~ /doesNotMatch/"
        description: "Show Search View"
      - command: "workbench.action.toggleSidebarVisibility"
        when: "searchViewletVisible"
        description: "Hide Search View"
    # For Zed, a single native action handles the toggle logic.
    zed:
      action: "project_search::ToggleFocus"
      context: "Workspace"
```

During import, the plugin will recognize either of the VSCode commands as matching the universal `actions.view.toggleSearch` ID. During export, it will generate both keybindings for the same key, correctly recreating the toggle behavior.

### Contribution Workflow

1.  **Find an Unmapped Command**: Identify a command in an editor that you'd like to use in OneKeymap CLI.
2.  **Define a Universal ID**: Create a new, descriptive, and unique `id` for this action.
3.  **Locate or Create a YAML File**: Find the appropriate `*.yaml` file in `config/action_mappings/` (e.g., `editor.yaml` for text editing commands) or create a new one if a suitable category doesn't exist.
4.  **Add the Mapping**: Add your new mapping to the file, providing the editor-specific command and context.
5.  **Add Mappings for Other Editors (Optional but Recommended)**: To make the action truly universal, add the corresponding commands for other supported editors.
6.  **Submit a Pull Request**: Open a PR with your changes. We'll review it and merge it in.

Thank you for helping make OneKeymap better!
