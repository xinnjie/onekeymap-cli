# OneKeymap Action Mappings

This directory contains the action mapping definitions that serve as the "brain" of the OneKeymap system. These YAML files define how actions from different editors are translated to and from the universal OneKeymap format.

## Overview

The Action Mappings System is the core component that enables OneKeymap to translate keyboard shortcuts between different editors. Each mapping file defines a set of actions with their corresponding implementations across supported editors.

## File Structure

```
action_mappings/
├── README.md          # This documentation
├── editor.yaml        # Core text editing operations
├── navigation.yaml    # Cursor movement and navigation
└── workspace.yaml     # Workspace-level operations
```

## Mapping File Format

Each YAML file follows this structure:

```yaml
mappings:
  - id: "actions.category.actionName"
    description: "Human-readable description of what this action does"
    vscode:
      command: "vscode.command.name"
      when: "contextCondition"
    zed:
      action: "zed::Action"
      context: "ContextName"
    intellij:
      action: "ActionName"
      context: "ContextName"
    vim:
      command: "vim_command"
      mode: "normal|insert|visual"
```

### Field Definitions

- **`id`**: Unique universal action identifier following the pattern `actions.category.actionName`
- **`description`**: Clear, human-readable explanation of what the action does
- **Editor-specific sections**:
  - **VSCode**: `command` (the VSCode command ID) and optional `when` clause
  - **Zed**: `action` (the Zed action name) and `context` (when the action is available)
  - **IntelliJ**: `action` (IntelliJ action ID) and optional `context`
  - **Vim**: `command` (vim command/key sequence) and `mode` (vim mode where applicable)

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

## Categories

### Editor Actions (`editor.yaml`)
Core text editing operations such as:
- Copy, cut, paste
- Undo, redo
- Find, replace
- Format document
- Comment/uncomment
- Line operations (duplicate, delete, indent)

### Navigation Actions (`navigation.yaml`)
Cursor movement and code navigation:
- Word, line, file navigation
- Go to definition, references
- Error navigation
- Bracket matching
- Scrolling operations

### Workspace Actions (`workspace.yaml`)
Project and workspace-level operations:
- File management (open, save, close)
- Search across files
- Panel and sidebar management
- Terminal operations
- Editor splitting and focus

## Adding New Mappings

To add a new action mapping:

1. **Choose the appropriate file** based on the action category
2. **Create a unique ID** following the naming convention: `actions.category.actionName`
3. **Add comprehensive mappings** for all supported editors
4. **Test the mapping** by running the validation tools
5. **Update documentation** if introducing new categories

## Validation and Testing

### Automatic Validation
The system automatically validates:
- **Unique IDs**: No duplicate action IDs across all files
- **YAML syntax**: All files must be valid YAML
- **Required fields**: Each mapping must have an ID and description

### Testing Your Mappings
```bash
# Load and validate all mappings
cd onekeymap
go run test_mappings.go

# Test specific conversions
go test ./internal/mappings -v
```

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

#### Vim
- Use standard vim commands and key sequences
- Specify the correct mode (normal, insert, visual, etc.)
- Consider vim plugin compatibility for advanced features

### Quality Guidelines
1. **Test thoroughly**: Verify mappings work in actual editors
2. **Consider context**: Ensure `when` clauses and contexts are appropriate
3. **Maintain consistency**: Similar actions should have similar patterns
4. **Document edge cases**: Add comments for complex mappings

## Troubleshooting

### Common Issues

#### Mapping Not Found
- Verify the action ID exists in the mapping files
- Check for typos in editor-specific command names
- Ensure the target editor section is properly defined

#### Context Conflicts
- Review `when` clauses in VSCode mappings
- Check context specifications for Zed and IntelliJ
- Verify vim mode specifications

#### Conversion Failures
- Run the test suite to identify specific failures
- Check the conversion reports for detailed error information
- Validate YAML syntax using external tools

### Debugging Tools
```bash
# Validate YAML syntax
yamllint action_mappings/

# Test specific editor conversion
go run test_mappings.go | grep "vscode\|zed"

# Generate detailed conversion reports
go test ./internal/mappings -v -run TestConversion
```

## Contributing

When contributing new mappings:

1. **Follow the existing patterns** and naming conventions
2. **Test with actual editors** to ensure functionality
3. **Include comprehensive descriptions**
4. **Add mappings for all supported editors** when possible
5. **Update this README** if adding new categories or patterns

## Future Enhancements

Planned improvements to the mapping system:
- Support for key sequence mappings (multi-key combinations)
- Conditional mappings based on file types or project settings
- Plugin-specific mappings for popular editor extensions
- Automated mapping discovery from editor configurations
- Visual mapping editor and validator tools

## References

- [OneKeymap Design Document](../doc/design.md)
- [CLI Usage Guide](../doc/cli.md)
- [Plugin Development Guide](../doc/plugins.md)
