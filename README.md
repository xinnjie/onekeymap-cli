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

OneKeymap CLI is a powerful command-line tool that lets you import, export, and synchronize keyboard shortcuts between VSCode, IntelliJ IDEA, Zed, Xcode, Helix, and more. Stop reconfiguring keybindings every time you switch editors—maintain one universal keymap and deploy it everywhere.

| Import Keymap | View Keymap | Export Keymap |
|---|---|---|
| [<img src="assets/onekeymap-cli-import.gif" width="600" alt="Import Demo">](https://asciinema.org/a/748300) | [<img src="assets/onekeymap-cli-view.gif" width="600" alt="View Demo">](https://asciinema.org/a/ZNqGYNMKs0jVh5qH6Smv6ysni) | [<img src="assets/onekeymap-cli-export.gif" width="600" alt="Export Demo">](https://asciinema.org/a/748319) |
| Auto-detects your editor's keymap file and lets you know keybindings to import. | View keybindings by category and navigate with arrow keys. | Shows a diff of your target keymap file before exporting. |

---

## 🚀 Quick Start

### Installation

#### macOS

- **Homebrew**
  ```bash
  brew tap xinnjie/onekeymap
  brew install onekeymap-cli
  ```

#### Linux

- **Debian/Ubuntu (.deb)**
  ```bash
  wget https://github.com/xinnjie/onekeymap-cli/releases/download/v0.5.1/onekeymap-cli_0.5.1_x86_64.deb
  sudo dpkg -i onekeymap-cli_0.5.1_x86_64.deb
  ```
- **Fedora/RHEL/CentOS (.rpm)**
  ```bash
  wget https://github.com/xinnjie/onekeymap-cli/releases/download/v0.5.1/onekeymap-cli_0.5.1_x86_64.rpm
  sudo rpm -i onekeymap-cli_0.5.1_x86_64.rpm
  ```
- **Alpine (.apk)**
  ```bash
  wget https://github.com/xinnjie/onekeymap-cli/releases/download/v0.5.1/onekeymap-cli_0.5.1_x86_64.apk
  sudo apk add --allow-untrusted onekeymap-cli_0.5.1_x86_64.apk
  ```

#### Windows
- **Scoop**
  ```powershell
  scoop bucket add xinnjie https://github.com/xinnjie/scoop-bucket
  scoop install onekeymap-cli
  ```

- **Winget**
  ```powershell
  winget install xinnjie.onekeymap-cli
  ```

- **Zip Archive**
  Download `onekeymap-cli_Windows_<arch>.zip` from [GitHub Releases](https://github.com/xinnjie/onekeymap-cli/releases), extract it, and add the directory to your PATH, or run:
  ```powershell
  Expand-Archive -Path .\onekeymap-cli_*.zip -DestinationPath "$Env:USERPROFILE\onekeymap-cli"
  setx PATH "$Env:USERPROFILE\onekeymap-cli;$Env:PATH"
  ```

#### Cross-platform (Linux/MacOS/Windows)

- **Go Install**
  ```bash
  go install github.com/xinnjie/onekeymap-cli/cmd/onekeymap-cli@latest
  ```

- **From Release:**
Download the latest binary from [GitHub Releases](https://github.com/xinnjie/onekeymap-cli/releases/latest).

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

<details>
<summary><strong>Configuration</strong></summary>

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

</details>

---

## 🧩 Supported Editors & Actions

| Editor | Import | Export | Notes |
|--------|--------|--------|-------|
| **VSCode** | ✅ | ✅ |  |
| **Zed** | ✅ | ✅ |  |
| **IntelliJ IDEA** | ✅ | ✅ |  |
| **Xcode** | ✅ | ✅ | shortcut coverage is still limited (see [Action Support Matrix](action-support-matrix.md)) |
| **Helix** | ❌ | ✅ | TOML configuration support; shortcut coverage is still limited (see [Action Support Matrix](action-support-matrix.md)) |
| **Vim/Neovim** | 🚧 | 🚧 | Planned |

> See all supported actions: [action-support-matrix.md](action-support-matrix.md)

---

## Contributing

Contributions are welcome! Check out the [Contributing Guide](CONTRIBUTING.md) for more details on how to get started. Here are some ways you can help:

- Add Editor Support: Implement a new editor plugin
- Improve Mappings: [Enhance the action mapping configuration](CONTRIBUTING.md#enhancing-the-action-mapping-configuration)
- Report Bugs: Open an issue with reproduction steps
- Documentation: Improve docs and examples

For development setup and building from source, see the [Development section in CONTRIBUTING.md](CONTRIBUTING.md#development).

---

## Support

- Issues: [![GitHub Issues](https://img.shields.io/github/issues/xinnjie/onekeymap-cli?logo=github&label=Issues)](https://github.com/xinnjie/onekeymap-cli/issues)
- Discussions: [![Discord](https://img.shields.io/badge/Discord-Join%20the%20chat-5865F2?logo=discord&logoColor=white)](https://discord.com/invite/fW3TWuXj9A)

---

<details>
<summary>How It Works</summary>

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
</details>

---

## ✨ Companion App: OneKeymap GUI

Prefer a polished interface? Take a look at [OneKeymap.app](https://www.onekeymap.com/)—a paid GUI companion built on top of the CLI. `onekeymap-cli` will always remain free and open; the app is simply an optional bonus for those who enjoy visual workflows.

[![OneKeymap GUI screenshot](assets/onekeymap-app-hero.png)](https://www.onekeymap.com/)

---
