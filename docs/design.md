# Design Document

## Overview

The `OneKeymap` app is a cross-platform desktop application that enables users to import, export, and synchronize keyboard shortcuts between different code editors. The application follows a plugin-based architecture where each editor is supported through isolated modules, ensuring maintainability and extensibility.

### Supported Editors

**VSCode Family:**

**IntelliJ Family:**

**Other Editors:**
- Zed
- Vim
- Helix
- Xcode

The core design principle is to maintain a unified, editor-agnostic internal keymap representation (`KeymapSetting` struct type). This struct type represents the "what" (action) and "how" (key chord), while the "when" (context) is handled exclusively by the editor-specific plugins, ensuring the core data model remains simple and universal.

## Architecture

The application follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface                            │
├───────────────────────────-─────────────────────────────────┤
│                  Core Service Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │   Import Svc    │  │   Export Svc    │                   │
│  └─────────────────┘  └─────────────────┘                   │
├─────────────────────────────────────────────────────────────┤
│                 Editor Implementations (Plugins)            │
│  (Handles all context-specific logic)                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │   VSCode        │  │   IntelliJ      │  │    Zed      │  │
│  │   Family        │  │   Family        │  │   Plugin    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  |
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │     Vim         │  │     Helix       │  │    Xcode    │  │
│  │    Plugin       │  │    Plugin       │  │   Plugin    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│      Domain Data Layer & Action Mappings (The "Truth")      │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │  KeymapSetting  │  │ action_mappings │                   │
│  └─────────────────┘  └─────────────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

## Action Mapping and Context Strategy

A key challenge is that editors like Zed and VSCode use a `context` (or `when` clause) to determine if a keybinding is active. Our strategy isolates this complexity within the plugins, keeping the central `KeymapSetting` clean.

### The `KeymapSetting` Struct Type: Pure and Simple

The `KeymapSetting` struct type remains the universal, editor-agnostic data structure. It **only** contains the logical action and the key combination, with no concept of context.

### The Action Mappings: A Modular Knowledge Base

The "brain" of the translation is a collection of YAML files located in a dedicated [`action_mappings`](onekeymap/onekeymap-cli/config/action_mappings) directory. This modular approach avoids a single monolithic file, enhancing maintainability and scalability. Each file represents a logical feature group , e.g. [editor.yaml](onekeymap/onekeymap-cli/config/action_mappings/editor.yaml).

A key feature of this design is its ability to handle both simple and stateful actions. For a given universal `id`, the mapping for an editor can be either a single command object or a **list of command objects**. This list-based approach is how we model [stateful toggles](#stateful-toggles).

Directory Structure:
```
action_mappings/
├── editor.yaml
├── navigation.yaml
├── view-management.yaml
└── ...
```

**Example `editor.yaml`:**
```yaml
# action_mappings/editor.yaml
mappings:
  - id: "actions.editor.copy"
    description: "Copies the current selection."
    zed:
      action: "editor::Copy"
      context: "Editor && vim_mode != 'insert'"
    vscode:
      command: "editor.action.clipboardCopyAction"
      when: "editorTextFocus && editorHasSelection"
  # ... other editor-related actions
```

**Action Mapping Loading, Merging, and Validation:**
The core services will initialize the mapping data using the following process:
1.  **Read Directory**: Scan the `action_mappings` directory for all `*.yaml` files.
2.  **Parse and Merge**: Iterate through each file, parse its content, and merge the `mappings` lists into a single, unified collection in memory.
3.  **Conflict Detection**: During the merge process, maintain a set of all encountered `id`s. If an `id` is found that already exists in the set, the application must raise a critical error and halt. This enforces the global uniqueness of action `id`s and prevents ambiguous mappings.

#### Stateful Toggles
While the core `KeymapSetting` remains simple, the mapping layer is designed to handle more complex scenarios, such as stateful "toggle" actions where the same key performs different commands based on application state (e.g., using a key to both open and close a panel). This is achieved by allowing a single, universal `id` in the `action_mappings` to map to a **list** of context-dependent commands for a specific editor.

This design allows us to model a user's logical intent (e.g., "toggle search panel") as a single, unified action while still precisely recreating the nuanced, context-aware behavior in the target editor. The specifics of this implementation are detailed in the "Action Mappings" and "Plugins" sections below.

**Example:**
```json
# Stateful "toggle" mapping
- id: "actions.view.toggleSearch"
  description: "Toggles the visibility of the Search view."
  # For VSCode, this logic requires two distinct, context-dependent commands.
  vscode:
    - command: "workbench.view.search"
      when: "!searchViewletVisible"
      description: "Show Search View"
    - command: "workbench.action.toggleSidebarVisibility" # Or a more specific hide command
      when: "searchViewletVisible"
      description: "Hide Search View"
  # For Zed, the same logic is handled by a single native toggle action.
  zed:
    action: "workspace::ToggleLeftDock"
```
### 3. Plugins

Plugins are responsible for all logic related to context. They use the `action_mappings` knowledge base to translate between the editor-specific format and the universal `KeymapSetting`.

For simple mappings, the process is a direct lookup. For more complex scenarios like stateful toggles, the plugins follow the logic outlined in the `Stateful Toggles` section above.

####  3.1 Import Workflow
**General Import Workflow (Example: from VSCode for a simple action):**
1.  **Parse**: The plugin reads a keybinding, e.g., `ctrl+c` for the command `editor.action.clipboardCopyAction`.
2.  **Query**: It searches `action_mappings` to find the corresponding universal `id` (`actions.copy`).
3.  **Generate**: It creates a `Keymap` message with the universal action and the parsed key binding. The editor-specific `when` clause is discarded, its meaning now encoded in the universal `id`.

**Stateful Toggle Import Workflow (Example: from VSCode for a toggle action):**
1.  When the importer encounters a keybinding from VSCode, such as `(key: "alt+3", command: "workbench.view.search", when: "!searchViewletVisible")`.
2.  It iterates through `action_mappings.yaml`. When it inspects `actions.view.toggleSearch`, it scans the list under the `vscode` key.
3.  It finds a matching `command` and `when` entry in the list.
4.  It then maps this keybinding to the universal `actions.view.toggleSearch` action.
5.  When the importer encounters the second binding for `alt+3` (for hiding), it repeats the process, again mapping to `actions.view.toggleSearch`.
6.  The core service deduplicates the results when generating the final `KeymapSetting`, ensuring the combination `(action: "actions.view.toggleSearch", keys: "alt+3")` appears only once.

####  3.2 Export Workflow
**General Export Workflow (Example: to Zed for a simple action):**
1.  **Read**: The plugin reads a `Keymap` message, e.g., for the action `actions.copy`.
2.  **Query**: It finds the entry for this `id` in `action_mappings`.
3.  **Find Target**: It retrieves the `zed` field, which contains the target action (`editor::Copy`) and any necessary context.
4.  **Generate**: It constructs the final JSON object for Zed's `keymap.json`.

**Stateful Toggle Export Workflow (Example: to VSCode for a toggle action):**
1.  When exporting to VSCode, the exporter reads `{ action: "actions.view.toggleSearch", keys: "alt+3" }`.
2.  It looks up `actions.view.toggleSearch` in `action_mappings.yaml`.
3.  It discovers that the `vscode` field is a list.
4.  Therefore, it generates a keybinding for each item in the list, all using `alt+3` as the shortcut, perfectly reconstructing the original context-based toggle behavior in `keybindings.json`.
5.  When exporting to Zed, it finds that the `zed` field is a single object, so it generates only one keybinding, as before.

This approach ensures that `KeymapSetting` is a truly universal format, while preserving the fidelity of context-aware keymaps during translation.

## Conflict Resolution Strategy

To manage keybinding conflicts during the import process without introducing high complexity in the initial phase, the application will generate a machine-readable conflict report. This approach provides clear feedback to the user and establishes a foundation for future interactive resolution features.

Read [Conflict Resolution](.kiro/specs/one-keymap/conflict_resolution.md) for more details.

## Components and Interfaces

To promote modularity, testability, and adherence to Go's best practices, the core services are defined as interfaces. This decouples the application's components from their concrete implementations. The interfaces are designed to be flexible, using `io.Reader`/`io.Writer` for data streams and options structs for configuration.

### Interfaces

#### Core Service Interfaces

##### Importer

The [`Importer`](../pkg/api/importerapi/importer.go) interface defines the contract for converting editor-specific keymaps into the universal format.

```go
// Importer defines the interface for the import service, which handles the
// conversion of editor-specific keymaps into the universal format.
type Importer interface {
	// Import converts keymaps from a source stream. It returns the converted
	// settings and a report detailing any conflicts or unmapped actions.
	Import(ctx context.Context, opts ImportOptions) (*ImportResult, error)
}
```

##### Exporter

The [`Exporter`](../pkg/api/exporterapi/exporter.go) interface defines the contract for converting universal keymap into an editor-specific format.

```go
// Exporter defines the contract for converting a universal KeymapSetting
// into an editor-specific format.
type Exporter interface {
	// Export converts a KeymapSetting and writes it to a destination stream.
	// It returns a report detailing any issues encountered during the conversion.
	Export(
		ctx context.Context,
		destination io.Writer,
		setting keymap.Keymap,
		opts ExportOptions,
	) (*ExportReport, error)
}
```

#### Plugin Interfaces

The [plugin](../pkg/api/pluginapi/plugin.go) system provides a standard interface for all editor-specific implementations:

```go
// Plugin is the core interface that all editor plugins must implement.
// It defines the contract for importing and exporting keymaps.
type Plugin interface {
    // EditorType returns the unique identifier for the plugin (e.g., "vscode", "zed").
    EditorType() EditorType

    // ConfigDetect returns the default path to the editor's configuration file based on the platform.
    // Return multiple paths if the editor has multiple configuration files.
    ConfigDetect(opts ConfigDetectOptions) (paths []string, installed bool, err error)

    // Importer returns an instance of PluginImporter for the plugin.
    // Return ErrNotSupported if the plugin does not support importing.
    Importer() (PluginImporter, error)

    // Exporter returns an instance of PluginExporter for the plugin.
    // Return ErrNotSupported if the plugin does not support exporting.
    Exporter() (PluginExporter, error)
}
```

## Data Models

### KeymapSetting Structure

The central data structure is defined in `pkg/api/keymap/keymap.go`:

- **KeymapSetting**: Root container for all keymaps.
- **Keymap**: Individual mapping between a logical action and a key chord.
- **KeyChord**: Represents a key combination with modifiers.
- **Key**: Individual key with code (e.g., "a", "enter", "space").
- **KeyModifier**: Enum for modifier keys (Shift, Ctrl, Alt, Meta).

### KeymapSetting Config File Format

To ensure a good user experience for manual editing and sharing, the `KeymapSetting` is serialized to a human-readable config file (`onekeymap.json` or `.yaml`). This format intentionally differs from the direct JSON representation of the internal structure to prioritize readability and ease of use.

A dedicated adapter layer is responsible for translating between this user-friendly format and the internal `KeymapSetting` structure.

**User-Friendly Config Format (`onekeymap.json`):**
```json
{
  "keymaps": [
    {
      "action": "actions.edit.copy",
      "keys": "ctrl+c"
    },
    {
      "action": "actions.edit.paste",
      "keys": "ctrl+v"
    },
    {
      "action": "actions.view.showCommandPalette",
      "keys": "shift shift"
    },
    {
      "action": "actions.view.showSearch",
      "keys": "ctrl+shift+f"
    }
  ]
}
```
