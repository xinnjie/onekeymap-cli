# Export Coverage Design

This document describes how onekeymap-cli reports export coverage to help users understand what changes will be made and what limitations exist.

## Overview

When exporting a `keymap.Keymap` to an editor-specific configuration, users need to know:

1. **What changes will be made?** - The diff between before/after configurations
2. **How much of my keymap was successfully exported?** - Export coverage report

## Data Flow

```
User Input keymap.Keymap (requested export)
         │
         ▼
    ┌─────────┐
    │ Export  │◄── Original editor config (OriginalConfig)
    └────┬────┘
         │
         ▼
Exported editor config ──[Import]──► Effective keymap.Keymap (what actually works)
```

## Design Approach

### Two Separate Reports

| Report | Question Answered | Data Source |
|--------|-------------------|-------------|
| **Diff** | "What changes to the config file?" | Editor config before/after (existing) |
| **ExportCoverage** | "How much of my request was fulfilled?" | Input vs PluginExporter report |

### ExportCoverage Structure

```go
type ExportCoverage struct {
    // TotalActions is the number of actions requested for export.
    TotalActions int
    // FullyExported is the number of actions where all keybindings were exported.
    FullyExported int
    // PartiallyExported lists actions where some keybindings could not be exported.
    PartiallyExported []PartialExportedAction
}

type PartialExportedAction struct {
    Action    string                   // e.g. "actions.clipboard.copy"
    Requested []keybinding.Keybinding  // What user wanted
    Exported  []keybinding.Keybinding  // What was actually exported
    Reason    string                   // Why partial (e.g., editor limitation)
}
```

## Plugin Implementation

Each `PluginExporter` is responsible for reporting `ExportedReport`:

```go
type ExportedReport struct {
    Actions []ActionExportResult
}

type ActionExportResult struct {
    Action    string
    Requested []keybinding.Keybinding  // Keybindings requested to export
    Exported  []keybinding.Keybinding  // Keybindings actually exported
    Reason    string                   // Optional: why some were not exported
}
```

### When to Report Partial Export

A plugin should report partial export when:

1. **Editor limitation**: The editor doesn't support certain key combinations
2. **Conflict resolution**: Some keybindings were skipped due to conflicts
3. **Format restriction**: The editor config format can't represent certain bindings

## CLI Output Example

```
Export Summary:
  ✓ 45/50 actions fully exported
  △ 3 actions partially exported:
    - editor.copy: wanted 3 keybindings, got 2 (VSCode does not support F13+Shift)
    - editor.paste: wanted 2 keybindings, got 1 (Conflict with existing binding)
  ✗ 2 actions skipped:
    - custom.action1: no mapping for VSCode
    - custom.action2: unsupported modifier combination
```
