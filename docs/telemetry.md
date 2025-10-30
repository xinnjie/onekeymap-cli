# Telemetry Data Collection

OneKeymap CLI collects anonymous telemetry data to help improve the tool. This document explains what data is collected, how it's used, and how you can control telemetry.

## Data Collection Overview

Telemetry is **disabled by default**. When you first run the CLI in interactive mode, you'll be prompted to opt-in to telemetry. You can also enable or disable it manually in your configuration.

## What Data is Collected

### 1. Usage Metrics

#### Unknown Actions
- Metric: `onekeymap.unknown_action_total`
- Description: Count of unrecognized editor actions encountered during import
- Attributes:
  - `editor_type`: The editor type (e.g., "vscode", "intellij")
  - `action`: The unrecognized action name
- Purpose: Identify missing keymap mappings that should be added

#### Import Operations
- Metric: `onekeymap.import_total`
- Description: Total count of keymap import operations
- Purpose: Understand how frequently users import keymaps

#### Export Operations
- Metric: `onekeymap.export_total`
- Description: Total count of keymap export operations
- Purpose: Understand how frequently users export keymaps

### 2. System Information

The following system information is automatically included with all metrics:

#### Service Identification
- Service Name: `onekeymap-cli`
- Service Version: Your installed CLI version
- Operating System: Your OS (macOS, Windows, Linux)

## What Data is NOT Collected

- **No Personal Information**: Names, email addresses, or other personal data
- **No File Contents**: Actual keymap configurations or file contents
- **No File Paths**: Local file paths or directory structures
- **No Keystrokes**: Individual key presses or keyboard shortcuts
- **No Network Activity**: Other network requests or browsing data

## How Data is Used

Collected data helps us:

1. Improve Compatibility: Unknown action reports help identify missing editor action mappings
2. Understand Usage Patterns: Import/export frequencies guide development priorities

## Configuration

### Enable/Disable Telemetry

You can control telemetry through your configuration file or command-line flags:

**Configuration file** (`~/.config/onekeymap/config.yaml`):
```yaml
telemetry:
  enabled: true  # or false to disable
```

**Command-line flag**:
```bash
onekeymap --telemetry=1 [command] # or --telemetry=0 to disable
```

## Privacy Commitment

- All telemetry data is **anonymous** and **aggregated**
- No personally identifiable information is collected
- You have full control over enabling/disabling telemetry
- Data collection is transparent and documented in this file
- The tool functions identically whether telemetry is enabled or disabled
