## Conflict Resolution Strategy

To manage keybinding issues during the application's lifecycle, the application will generate a comprehensive, machine-readable report. This strategy is expanded beyond simple import/export failures to include a robust validation and conflict detection mechanism that identifies a wide range of potential problems in a user's keymap configuration.

This provides clear, actionable feedback to the user and establishes a foundation for future interactive resolution features.

The detected issues are categorized into two main types:
1.  **Internal Consistency Conflicts**: Problems that can be identified by statically analyzing the `KeymapSetting` (`onekeymap.json`) against the `action_mappings` knowledge base. These checks occur during the `import` process or when using a dedicated `validate` command.
2.  **Export-Context Conflicts**: Issues that only become apparent when exporting the universal `KeymapSetting` to a specific editor's format, as they depend on the target editor's capabilities and conventions.

### Validation Architecture

To ensure a modular, maintainable, and extensible design, the conflict detection logic is built around a **Strategy Pattern**. Each validation check is encapsulated as an independent rule that conforms to a `ValidationRule` interface.

A central `Validator` service orchestrates the process by executing a pipeline of these rules. Each rule receives a `ValidationContext`, adds any detected issues to the report within that context, and passes the context to the next rule.

This approach offers several advantages:
-   **Separation of Concerns**: Each rule is self-contained and focuses on a single type of conflict.
-   **Extensibility**: New validation rules can be added without modifying existing logic.
-   **Testability**: Each rule can be unit-tested in isolation.

The core components of this architecture are defined as follows:

```go
// pkg/api/validateapi/validtation_rule.go

type ValidationRule interface {
	Validate(ctx context.Context, validationContext *ValidationContext) error
}
```

### Types of Detectable Conflicts and Warnings

#### Internal Consistency Conflicts

1.  **Keybinding Conflict (Error)**
    *   **Description**: The most critical conflict, where a single key chord (e.g., `ctrl+c`) is mapped to multiple, distinct `one_keymap_id`s (e.g., `actions.editor.copy` and `actions.editor.paste`). This creates ambiguity, as the system cannot determine which action to execute.
    *   **Example**: `ctrl+c` is bound to both `copy` and `paste`.
    *   **Detection Stage**: During `import` or `validate`.

2.  **Dangling Action ID (Error)**
    *   **Description**: A keymap entry in the user's configuration refers to an `action` ID that does not exist in the entire `action_mappings` collection. This is typically caused by a typo or the removal/renaming of an action.
    *   **Example**: The user's config specifies `actions.editor.copi` instead of `actions.editor.copy`.
    *   **Detection Stage**: During `import` or `validate`.

3.  **Duplicate Mapping (Warning)**
    *   **Description**: The exact same combination of `action` and `keys` appears more than once in the configuration. While not a functional error, it indicates redundancy and potential manual editing mistakes.
    *   **Example**: The line `{ "action": "actions.editor.copy", "keys": "ctrl+c" }` appears twice in the file.
    *   **Detection Stage**: During `import` or `validate`.

#### Export-Context Conflicts

1.  **Unsupported Action for Target (Error)**
    *   **Description**: A valid action in the user's configuration cannot be exported to a specific editor because there is no corresponding mapping for that editor in the `action_mappings`.
    *   **Example**: Exporting `actions.git.interactiveRebase` to Zed, but the mapping only defines a `vscode` command.
    *   **Detection Stage**: During `export`.

2.  **Potential Shadowing of Native Keybinding (Warning)**
    *   **Description**: A user's keymap overrides a critical, conventional keybinding for the target editor or its operating system (e.g., `Cmd+Q` for quitting on macOS, `Ctrl+Alt+Delete` on Windows). This is flagged as a warning to prevent unintentional behavior that might disrupt the user's workflow.
    *   **Example**: Mapping `Cmd+Q` to "Format Document" when exporting to any app on macOS.
    *   **Detection Stage**: During `export`.

### Machine-Readable Conflict Report

The report is a machine-readable structure detailing all issues found. Its schema is formally defined by the `ValidationReport` message in `proto/keymap/v1/report.proto`. This structured format is designed to be parsed by other tools or by a future interactive conflict resolver.

**Example `ValidationReport` (JSON representation):**
```json
{
  "source_editor": "vscode",
  "summary": {
    "mappings_processed": 155,
    "mappings_succeeded": 150,
    "conflicts_found": 3,
    "warnings_issued": 2,
    "keybind_conflicts": 1,
    "dangling_actions": 1,
    "unsupported_on_export": 1
  },
  "issues": [
    {
      "type": "keybind_conflict",
      "keybinding": "ctrl+shift+x",
      "actions": [
        "actions.extensions.show",
        "actions.debug.run"
      ]
    },
    {
      "type": "dangling_action",
      "action_id": "actions.editor.sellectAll",
      "keybinding": "ctrl+a",
      "suggestion": "Did you mean 'actions.editor.selectAll'?"
    },
    {
      "type": "unsupported_action_for_target",
      "action_id": "actions.git.interactiveRebase",
      "keybinding": "ctrl+alt+r",
      "target_editor": "zed"
    }
  ],
  "warnings": [
    {
      "type": "duplicate_mapping",
      "action_id": "actions.editor.copy",
      "keybinding": "ctrl+c",
      "message": "This keymap is defined multiple times in the source configuration."
    },
    {
      "type": "potential_shadowing",
      "keybinding": "cmd+q",
      "action_id": "actions.editor.formatDocument",
      "target_editor": "vscode",
      "message": "This key chord is the default for quitting applications on macOS."
    }
  ]
}
```
