---
description: OneKeymap VSCode plugin import behavior is kind of complex, so we document it here.
---

# OneKeymap VSCode plugin Import Behavior

This document describes how VSCode keybindings are imported into OneKeymap and how ambiguities are resolved deterministically.

## Matching Priority

The importer resolves a VSCode keybinding to a OneKeymap action using the following priority:

1. Exact match: command + args match, and when matches exactly (non-empty equality)
2. Wildcard-when match: command + args match, and when in mapping is empty (treated as wildcard)
3. Fallback A: command + args match while when differs (when is ignored)
4. Fallback B: command only (only considered when args is nil in the incoming keybinding)

Within each priority level, candidates are filtered and selected deterministically (see below).

## Import Control

Each VSCode mapping entry supports an optional boolean field `disableImport`.

### Excluding specific commands

By default, all VSCode mapping entries are used for both import (resolving VSCode keybindings to OneKeymap actions) and export (generating VSCode `keybindings.json`).

If an entry is marked with `disableImport: true`, it is **excluded** from the import process. It will still be used for export.

This is useful when an action maps to multiple VSCode commands, but you want to enforce a canonical command for import, or avoid ambiguity.

### Example

```yaml
mappings:
  - id: "actions.view.toggleExplorer"
    description: "Toggle Explorer view"
    vscode:
      - command: "workbench.view.explorer"
        when: "viewContainer.workbench.view.explorer.enabled"
        # Default is disableImport: false, so this is used for import.
      - command: "workbench.action.toggleSidebarVisibility"
        when: "explorerViewletFocus"
        disableImport: true # This command is ignored during import.
```

## Rationale

- Determinism: We collect all candidates first and then decide, removing any non-determinism from Go map iteration order.
- Explicit preference: `disableImport` lets authors exclude mappings that are only intended for export or are ambiguous.
- Backwards compatible: Default behavior (import everything) works for most cases.

## Notes and Recommendations

- Keep `when` empty only for truly generic/wildcard contexts. More specific `when` conditions are preferred at higher priority.
- If you have multiple mappings for the same action, decide if all should be importable. If one is "legacy" or "secondary", use `disableImport: true`.
- When multiple candidates remain, the importer picks the smallest action ID lexicographically to ensure stable outcomes.
- Args equality is computed via canonical JSON encoding to avoid map iteration order and numeric type (e.g., int vs float) discrepancies.
- The importer skips unknown or unparsable keybindings and may return an empty `KeymapSetting` without error.
