# Base keymap import behavior

## What is Base Keymap

Base Keymap refers to the default keybindings of different editors, such as the default shortcuts for VSCode, Zed, or IntelliJ.

This document describes how OneKeymap handles base keymaps through the import plugin system.

## Design Philosophy

OneKeymap configs are **self-contained and complete**. Each `onekeymap.json` file contains the full set of keymaps without external dependencies. There is no field about "base keymap" in the config schema.

Base keymaps (e.g., IntelliJ default shortcuts, VSCode default shortcuts) are treated as **import sources**, just like importing from an actual editor installation. Users can import a base keymap once to initialize their `onekeymap.json`, then customize it as needed.

## Schema

The OneKeymap JSON is simple and self-contained:

```json
{
  "version": "1.0",
  "keymaps": [
    { "id": "actions.copy", "keybinding": "cmd+c" },
    { "id": "actions.paste", "keybinding": "cmd+v" }
  ]
}
```

- `version`: config format version (currently "1.0")
- `keymaps`: complete list of all keymaps

## Built-in base keymaps

Built-in base keymaps are embedded JSON files under `onekeymap/onekeymap-cli/config/base/` at build time. These files serve as **import sources** for the base keymap import plugin.

Filenames follow the pattern:
- `{editor}-{platform}.json`, e.g., `intellij-mac.json`, `vscode-mac.json`, `intellij-windows.json`

You can list available base keymaps programmatically via `config/base.List()`.

## Import workflow

### Importing a base keymap

To initialize an `onekeymap.json` with a base keymap:

```bash
onekeymap import --from basekeymap --input intellij-mac
```

This works exactly like importing from VSCode or any other editor:
1. Reads the embedded base keymap JSON file
2. Converts it to the onekeymap format
3. Merges with existing `onekeymap.json` (if any)
4. Writes the complete, self-contained config

### Interactive mode

In interactive mode, the CLI will:
1. Prompt for the source (e.g., vscode, intellij, basekeymap)
2. If source is `basekeymap`, prompt to select which base keymap to import
3. Show preview of changes
4. Write the complete config

### Subsequent imports

After initializing with a base keymap, users can:
- Import from actual editors (e.g., `--from vscode`) to override specific shortcuts
- Manually edit `onekeymap.json` to customize shortcuts
- Import from another base keymap to completely replace the config

All imports use the same merge strategy: keymaps with the same `id` are replaced, new `id`s are appended.

## Load and Save behavior

### Load
`keymap.Load()` simply reads and parses the `onekeymap.json` file. No base resolution or merging is performed at load time, because the file is already complete.

The `LoadOptions.BaseReader` and `LoadOptions.BaseKeymap` options are removed.

### Save
`keymap.Save()` writes the complete keymap to JSON:

```go
keymap.Save(writer, setting, keymap.SaveOptions{
  Platform: platform.PlatformMacOS, // controls key format (modifier names, etc.)
})
```

Keymaps are written in stable order (sorted by `id`) for deterministic diffs.

## Base keymap import plugin

The base keymap import plugin (`basekeymap`) implements the standard `Plugin` interface:

- **EditorType**: Returns `"basekeymap"`
- **ConfigDetect**: Returns the list of available base keymaps (from `config/base.List()`)
- **Importer**: Reads the specified base keymap JSON and returns it as a `Keymap` proto
- **Exporter**: Not supported (returns `ErrNotSupported`)

### Plugin input

The `PluginImporter.Import()` receives the base keymap name (e.g., "intellij-mac") through the standard input mechanism:
- In CLI: `--input intellij-mac` or interactive selection
- The plugin reads the corresponding JSON from `config/base.Read(name)`
