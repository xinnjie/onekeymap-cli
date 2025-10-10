<div align="center">
  <a href="https://github.com/xinnjie/onekeymap-cli">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="assets/logo-onekeymap.svg" />
      <img src="assets/logo-onekeymap.svg" alt="OneKeymap CLI logo" height="128" />
    </picture>
  </a>
  <h1>OneKeymap CLI</h1>
</div>

[![Go Version](https://img.shields.io/github/go-mod/go-version/xinnjie/onekeymap-cli)](https://go.dev/)[![License](https://img.shields.io/github/license/xinnjie/onekeymap-cli)](LICENSE.md)[![Release](https://img.shields.io/github/v/release/xinnjie/onekeymap-cli)](https://github.com/xinnjie/onekeymap-cli/releases)

---
**Sync your keyboard shortcuts across all your code editors.**

OneKeymap CLI is a powerful command-line tool that lets you import, export, and synchronize keyboard shortcuts between VSCode, IntelliJ IDEA, Zed, Helix, and more. Stop reconfiguring keybindings every time you switch editors—maintain one universal keymap and deploy it everywhere.

[![asciicast](https://asciinema.org/a/746874.svg)](https://asciinema.org/a/746874)


---

## 🚀 Quick Start

> ❗️Currently the OneKeymap CLI only supports macOS. Windows and Linux support is coming soon.

### Installation

**macOS (Homebrew):**
```bash
brew install xinnjie/homebrew-onekeymap/onekeymap-cli
```

**Go Install:**
```bash
go install github.com/xinnjie/onekeymap-cli/cmd/onekeymap-cli@latest
```

**From Release:**
Download the latest binary from [GitHub Releases](https://github.com/xinnjie/onekeymap-cli/releases).

### Basic Usage

**Import from editors interactively:**

```bash
onekeymap-cli import
```

**Export to editors interactively:**

```bash
onekeymap-cli export
```

---

## Useful Commands

The following commands help you get the most from the OneKeymap CLI:

- **`onekeymap-cli help`** Quick summary of all subcommands and available flags.
- **`onekeymap-cli import`** Convert editor-specific shortcuts into the universal `onekeymap.json` format.
- **`onekeymap-cli export`** Generate editor keymap files from your universal keymap.
- **`onekeymap-cli migrate`** Chain `import` and `export` in one step to move between editors.
- **`onekeymap-cli view`** Inspect the actions and bindings stored in an existing universal keymap.

You can append `-h` or `--help` to any subcommand for detailed flag descriptions and examples.


---

## Configuration

OneKeymap can be configured via a config file at `~/.config/onekeymap/config.yaml`:

```yaml
# Default path for universal keymap
onekeymap: ~/.config/onekeymap/keymap.json

# Editor-specific config paths (optional, auto-detected by default)
editors:
  vscode:
    keymap_path: ~/Library/Application Support/Code/User/keybindings.json
  zed:
    keymap_path: ~/.config/zed/keymap.json
  intellij:
    keymap_path: ~/Library/Application Support/JetBrains/IntelliJIdea2024.1/keymaps/custom.xml
```

---

## 🧩 Supported Editors

| Editor | Import | Export | Notes |
|--------|--------|--------|-------|
| **VSCode** | ✅ | ✅ | Full support including `when` clauses |
| **Zed** | ✅ | ✅ | Full support including contexts |
| **IntelliJ IDEA** | ✅ | ✅ | Supports XML keymap format; shortcut coverage is still limited (see [Action Support Matrix](action-support-matrix.md)) |
| **Helix** | ❌ | ✅ | TOML configuration support; shortcut coverage is still limited (see [Action Support Matrix](action-support-matrix.md)) |
| **Vim/Neovim** | 🚧 | 🚧 | Planned |

---

## Contributing

Contributions are welcome! Check out the [Contributing Guide](CONTRIBUTING.md) for more details on how to get started. Here are some ways you can help:

- **Add Editor Support**: Implement a new editor plugin
- **Improve Mappings**: [Enhance the action mapping configuration](CONTRIBUTING.md#enhancing-the-action-mapping-configuration)
- **Report Bugs**: Open an issue with reproduction steps
- **Documentation**: Improve docs and examples

---

## Support

- **Issues**: [![GitHub Issues](https://img.shields.io/github/issues/xinnjie/onekeymap-cli?logo=github&label=Issues)](https://github.com/xinnjie/onekeymap-cli/issues)
- **Discussions**: [![Discord](https://img.shields.io/badge/Discord-Join%20the%20chat-5865F2?logo=discord&logoColor=white)](https://discord.com/invite/fW3TWuXj9A)

---

## How It Works

OneKeymap uses a **universal keymap format** that represents keyboard shortcuts in an editor-agnostic way. Here's the workflow:

```
┌─────────────┐
│   VSCode    │──┐
│  Keybindings│  │
└─────────────┘  │
                 │  Import
┌─────────────┐  │    ↓
│  IntelliJ   │──┼──────────────┐
│   Keymap    │  │              │
└─────────────┘  │    ┌─────────▼──────────┐
                 │    │  Universal Keymap  │
┌─────────────┐  │    │   (onekeymap.json) │
│     Zed     │──┘    └─────────┬──────────┘
│   Keymap    │                 │
└─────────────┘       Export    ↓
                 ┌───────────────────────┐
                 │  Any Supported Editor │
                 └───────────────────────┘
```

### Universal Keymap Format

Your keymap is stored in a clean, human-readable JSON format:

```json
{
  "keymaps": [
    {
      "action": "actions.edit.copy",
      "keys": "ctrl+c"
    },
    {
      "action": "actions.view.showCommandPalette",
      "keys": "ctrl+shift+p"
    },
    {
      "action": "actions.editor.quickFix",
      "keys": "ctrl+."
    }
  ]
}
```

### Action Mappings

OneKeymap maintains a comprehensive mapping that translates between editor-specific commands and universal actions. For example:

- `actions.edit.copy` maps to:
  - VSCode: `editor.action.clipboardCopyAction`
  - IntelliJ: `$Copy`
  - Zed: `editor::Copy`

This mapping layer handles context-specific behaviors, stateful toggles, and editor quirks automatically.

---

## ✨ Companion App: OneKeymap GUI

Prefer a polished interface? Take a look at [OneKeymap.app](https://onekeymap-landing-page.vercel.app/)—a paid GUI companion built on top of the CLI. `onekeymap-cli` will always remain free and open; the app is simply an optional bonus for those who enjoy visual workflows.

---
