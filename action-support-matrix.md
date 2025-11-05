# Action Support Matrix

## Code

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Show documentation hover | ✅ | ✅ | ✅ | ✅ | ✅ | Show documentation hover | actions.hover.showHover |
| Parameter hints | ✅ | ✅ | ✅ | N/A | ✅ | Trigger Parameter Hints | actions.refactor.triggerParameterHint |

## Code.Go

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Go to bracket | ✅ | ✅ | ✅ | N/A | N/A | Go to bracket | actions.go.bracket |
| Call hierarchy | ✅ | ❌ (not supported yet, see [`Support 'Show Call Hierarchy' as an LSP action` issue](https://github.com/zed-industries/zed/issues/14203)) | ✅ | ✅ | N/A | Show call hierarchy | actions.go.callHierarchy |
| Go to definition | ✅ | ✅ | ❌ (There is not `Go to definition` in intellij, use `Go to declaration` instead) | ✅ | N/A | Go to definition | actions.go.definition |
| Peek definition | ✅ | ❌ (Peek is not supported yet, see [`Peek or Preview Definitions Inline` discussion](https://github.com/zed-industries/zed/discussions/28282)) | ✅ | N/A | N/A | Peek definition | actions.go.definitionPeek |
| Go to declaration | ✅ | ✅ | ✅ | N/A | N/A | Go to declaration or usages | actions.go.goToDeclaration |
| Go to super | ✅ | ❌ (not supported yet, no issue tracked) | ✅ | N/A | N/A | Go to super class/super method | actions.go.goToSuper |
| Go to test | ✅ | ❌ (not supported yet, see [`Go to test` discussion](https://github.com/zed-industries/zed/discussions/40859)) | ✅ | N/A | N/A | Go to test | actions.go.goToTest |
| Go to implementations | ✅ | ✅ | ✅ | N/A | N/A | Go to implementations, For an interface, this shows all the implementors of that interface and for abstract methods, this shows all concrete implementations of that method. | actions.go.implementations |
| Peek declaration | ✅ | ✅ | ✅ | N/A | N/A | Peek declaration | actions.go.peekDeclaration |
| Reference peek | ✅ | ❌ (Peek is not supported yet, see [`Peek or Preview Definitions Inline` discussion](https://github.com/zed-industries/zed/discussions/28282)) | ✅ | N/A | N/A | Show usages / reference search | actions.go.referencePeek |
| Go to references | ✅ | ✅ | ✅ | ✅ | N/A | Go to references | actions.go.references |
| Go to type definition | ✅ | ✅ | ✅ | N/A | N/A | Go to type definition | actions.go.typeDefinition |
| Peek type definition | ✅ | ❌ (Peek is not supported yet, see [`Peek or Preview Definitions Inline` discussion](https://github.com/zed-industries/zed/discussions/28282)) | ✅ | N/A | N/A | Peek type definition | actions.go.typeDefinitionPeek |
| Type hierarchy | ✅ | ❌ (Not supported yet, see [`Type hierarchy (class inheritance tree) support` discussion](https://github.com/zed-industries/zed/discussions/16348)) | ✅ | N/A | N/A | Show type hierarchy | actions.go.typeHierarchy |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Peek call hierarchy | ✅ | ❌ (not supported yet, see [`Support 'Show Call Hierarchy' as an LSP action` issue](https://github.com/zed-industries/zed/issues/14203)) | ❌ (Use `CallHierarchy` instead) | N/A | N/A | Peek call hierarchy | actions.go.callHierarchyPeek | Use `CallHierarchy` instead |
| Jump to next counterpart | N/A | N/A | N/A | ✅ | N/A | Jump to next counterpart, like switching between .cpp file and .h file | actions.go.jumpToNextCounterpart | - |
</details>

## Code.Refactor

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Extract to method | N/A | N/A | N/A | ✅ | N/A | Extract to method | action.refactor.extractMethod |
| Extract to variable | N/A | N/A | N/A | ✅ | N/A | Extract to variable | action.refactor.extractVariable |
| Code action | ✅ | ✅ | ✅ | N/A | ✅ | Code Action... | actions.refactor.codeAction |
| Organize imports | ✅ | ✅ | ✅ | N/A | N/A | Organize Imports | actions.refactor.organizeImports |
| Quick fix | ✅ | ❌ (not supported yet, no issue tracked) | ✅ | N/A | N/A | Quick Fix... | actions.refactor.quickFix |
| Refactor code | ✅ | ❌ (not supported yet, see [Code refactoring in Zed ](https://github.com/zed-industries/zed/discussions/8623)) | ✅ | N/A | ✅ | Refactor This... | actions.refactor.refactor |
| Rename symbol | ✅ | ✅ | ✅ | ✅ | ✅ | Rename | actions.refactor.rename |
| Generate codes | ✅ | ❌ (Use `Code action` instead) | ✅ | N/A | N/A | Generate code... (Getters, Setters, Constructors, hashCode/equals, toString) | actions.refactor.sourceAction |

## Code.Suggestion

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Next suggestion | ✅ | ✅ | ✅ | N/A | N/A | Show next inline suggestion | actions.edit.inlineSuggest.next |
| Previous suggestion | ✅ | ✅ | ✅ | N/A | N/A | Show previous inline suggestion | actions.edit.inlineSuggest.previous |
| Show inline suggestion | ✅ | ✅ | ✅ | N/A | N/A | Show inline suggestion | actions.edit.inlineSuggest.show |
| Show suggestions | ✅ | ✅ | ✅ | N/A | N/A | Trigger Suggest | actions.edit.suggest.show |

## Debug

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Restart debugging | ✅ | ✅ | ✅ | N/A | ✅ | Restart Debugging | actions.run.restartDebugging |
| Evaluate selection | ✅ | ✅ | ✅ | N/A | N/A | Send selection to REPL | actions.run.selectionToRepl |
| Start debugging | ✅ | ✅ | ✅ | N/A | ✅ | Start Debugging | actions.run.startDebugging |
| Stop debugging | ✅ | ✅ | ✅ | N/A | ✅ | Stop Debugging | actions.run.stopDebugging |
| Toggle breakpoint | ✅ | ✅ | ✅ | N/A | ✅ | Toggle Breakpoint | actions.run.toggleBreakpoint |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Show debug console | ✅ | ❌ (zed do not have debug console) | ❌ (intellij have debug output with DebugPanel) | N/A | N/A | Show Debug Output Console view | actions.view.showDebugOutputConsole | Not all editors have debug console |
</details>

## Debug.Step

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Continue | ✅ | ✅ | ✅ | N/A | ✅ | Continue | actions.run.continue |
| Run to cursor | ✅ | ✅ | ✅ | ✅ | N/A | Run to Cursor | actions.run.runToCursor |
| Step into | ✅ | ✅ | ✅ | N/A | ✅ | Step Into | actions.run.stepInto |
| Step out | ✅ | ✅ | ✅ | ✅ | ✅ | Step Out | actions.run.stepOut |
| Step over | ✅ | ✅ | ✅ | ✅ | ✅ | Step Over | actions.run.stepOver |

## Editor

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Find in file | ✅ | ✅ | ✅ | N/A | ✅ | Find in current file | actions.edit.find |
| Find in project | ✅ | ✅ | ✅ | ✅ | ✅ | Find in all files in the project | actions.edit.findInFiles |
| Format document | ✅ | ✅ | ✅ | N/A | N/A | Format Document | actions.edit.formatDocument |
| Format selection | ✅ | ✅ | ✅ | N/A | N/A | Format Selection | actions.edit.formatSelection |
| Replace in file | ✅ | ✅ | ✅ | N/A | N/A | Replace in current file | actions.edit.replace |
| Replace in project | ✅ | ✅ | ✅ | ✅ | N/A | Replace in all files in the project | actions.edit.replaceInFiles |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Re-Indent code | N/A | N/A | N/A | ✅ | N/A | Re-Indent code | actions.edit.reIndent | Use `FormatSelections` instead |
| Toggle word wrap | ✅ | ✅ | ❌ (intellij has a `Soft-Wrap` configuration in settings) | N/A | N/A | Toggle word wrap in the editor | actions.view.toggleWordWrap | - |
</details>

## Editor.Appearance

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Decrease font size | N/A | N/A | N/A | ✅ | N/A | Decrease font size | actions.appearance.decreaseFontSize |
| Increase font size | N/A | N/A | N/A | ✅ | N/A | Increase font size | actions.appearance.increaseFontSize |

## Editor.Clipboard

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Copy text | ✅ | ✅ | ✅ | N/A | ✅ | Copy selected text/file | actions.clipboard.copy |
| Copy file path | ✅ | ✅ | ✅ | N/A | N/A | Copy file path | actions.clipboard.copyFilePath |
| Cut text | ✅ | ✅ | ✅ | N/A | ✅ | Cut selected text/file | actions.clipboard.cut |
| Paste text | ✅ | ✅ | ✅ | ✅ | ✅ | Paste text/file | actions.clipboard.paste |

## Editor.Comment

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Toggle block comment | ✅ | ❌ (not supported yet, see [`Toggle block comment` discussion](https://github.com/zed-industries/zed/discussions/4751)) | ✅ | N/A | ✅ | Toggle block comment | actions.edit.toggleBlockComment |
| Toggle line comment | ✅ | ✅ | ✅ | N/A | ✅ | Toggle line comment | actions.edit.toggleLineComment |

## Editor.Cursor

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Undo cursor | ✅ | ✅ | ✅ | N/A | N/A | Undo last cursor operation | actions.edit.cursorUndo |

## Editor.Cursor.File

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Move to bottom | ✅ | ✅ | ✅ | N/A | ✅ | Move caret to text end | actions.cursor.moveToBottom |
| Select to bottom | ✅ | ✅ | ✅ | N/A | ✅ | Select from cursor to text end | actions.cursor.moveToBottomSelect |
| Move to top | ✅ | ✅ | ✅ | N/A | ✅ | Move caret to text start | actions.cursor.moveToTop |
| Select to top | ✅ | ✅ | ✅ | N/A | ✅ | Select from cursor to text start | actions.cursor.moveToTopSelect |
| Page down | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor down by one page | actions.cursor.pageDown |
| Select page down | ✅ | ✅ | ✅ | N/A | N/A | Select down by one page | actions.cursor.pageDownSelect |
| Page up | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor up by one page | actions.cursor.pageUp |
| Select page up | ✅ | ✅ | ✅ | N/A | N/A | Select up by one page | actions.cursor.pageUpSelect |

## Editor.Cursor.Line

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Move to line end | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor to the end of the line | actions.cursor.lineEnd |
| Select line end | ✅ | ✅ | ✅ | N/A | ✅ | Select from cursor to the end of the line | actions.cursor.lineEndSelect |
| Move to line start | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor to the beginning of the line | actions.cursor.lineStart |
| Select line start | ✅ | ✅ | ✅ | N/A | ✅ | Select from cursor to the beginning of the line | actions.cursor.lineStartSelect |

## Editor.Cursor.Multi

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Add cursor above | ✅ | ✅ | ✅ | N/A | ✅ | Add cursor above current line | actions.selection.addCursorAbove |
| Add cursor below | ✅ | ✅ | ✅ | N/A | ✅ | Add cursor below current line | actions.selection.addCursorBelow |
| Add cursors to ends | ✅ | ✅ | ✅ | N/A | N/A | Add cursors to the end of selected lines | actions.selection.addCursorsToLineEnds |
| Add next occurrence | ✅ | ✅ | ✅ | N/A | ✅ | Add next occurrence of selection to multicursor | actions.selection.addNextOccurrence |
| Add previous occurrence | ✅ | ✅ | ✅ | N/A | ✅ | Add previous occurrence of selection to multicursor | actions.selection.addPreviousOccurrence |
| Select all occurrences | ✅ | ✅ | ✅ | N/A | N/A | Select all occurrences of current selection | actions.selection.selectAllOccurrences |

## Editor.Cursor.Word

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Move to previous word | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor to the start of the previous word | actions.cursor.wordLeft |
| Select previous word | ✅ | ✅ | ✅ | N/A | ❌ (helix move cursor with selection by default) | Select to the start of the previous word | actions.cursor.wordLeftSelect |
| Move to previous subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | N/A | ✅ | Move cursor to the start of the previous subword (hump) | actions.cursor.wordPartLeft |
| Select previous subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | N/A | ❌ (helix move cursor with selection by default) | Select to the start of the previous subword (hump) | actions.cursor.wordPartLeftSelect |
| Move to next subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | N/A | ✅ | Move cursor to the end of the next subword (hump) | actions.cursor.wordPartRight |
| Select next subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | N/A | N/A | Select to the end of the next subword (hump) | actions.cursor.wordPartRightSelect |
| Move to next word | ✅ | ✅ | ✅ | N/A | ✅ | Move cursor to the end of the next word | actions.cursor.wordRight |
| Select next word | ✅ | ✅ | ✅ | N/A | ❌ (helix move cursor with selection by default) | Select to the end of the next word | actions.cursor.wordRightSelect |

## Editor.Folding

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Fold | ✅ | ✅ | ✅ | ✅ | N/A | Collapse the current code block | actions.fold.fold |
| Fold all | ✅ | ✅ | ✅ | N/A | N/A | Collapse all code blocks in the editor | actions.fold.foldAll |
| Fold recursively | ✅ | ✅ | ✅ | N/A | N/A | Collapse the current code block and its children recursively | actions.fold.foldRecursively |
| Toggle fold | ✅ | ✅ | ✅ | N/A | N/A | Toggle Fold | actions.fold.toggleFold |
| Unfold | ✅ | ✅ | ✅ | ✅ | N/A | Expand the current code block | actions.fold.unfold |
| Unfold all | ✅ | ✅ | ✅ | N/A | N/A | Expand all code blocks in the editor | actions.fold.unfoldAll |
| Unfold recursively | ✅ | ✅ | ✅ | N/A | N/A | Expand the current code block and its children recursively | actions.fold.unfoldRecursively |

## Editor.Line

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Delete line | ✅ | ✅ | ✅ | N/A | N/A | Delete line | actions.edit.deleteLines |
| Insert line after | ✅ | ✅ | ✅ | N/A | N/A | Insert a new line after the current line | actions.edit.insertLineAfter |
| Insert line before | ✅ | ✅ | ✅ | N/A | N/A | Insert a new line before the current line | actions.edit.insertLineBefore |
| Join lines | ✅ | ✅ | ✅ | N/A | N/A | Join lines | actions.edit.joinLines |
| Insert line break | ✅ | ✅ | ✅ | N/A | N/A | Insert line break | actions.edit.lineBreakInsert |
| Copy line down | ✅ | ✅ | ✅ | ✅ | ✅ | Copy current line down | actions.selection.copyLineDown |
| Copy line up | ✅ | ✅ | ❌ (not supported, no ticket tracked) | N/A | ✅ | Copy current line up | actions.selection.copyLineUp |
| Move line down | ✅ | ✅ | ✅ | ✅ | N/A | Move current line down | actions.selection.moveLineDown |
| Move line up | ✅ | ✅ | ✅ | ✅ | ✅ | Move current line up | actions.selection.moveLineUp |

## Editor.Selection

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Expand selection | ✅ | ✅ | ✅ | N/A | ✅ | Expand selection | actions.selection.expand |
| Select all | ✅ | ✅ | ✅ | N/A | ✅ | Select all text in the editor | actions.selection.selectAll |
| Shrink selection | ✅ | ✅ | ✅ | N/A | ✅ | Shrink selection | actions.selection.shrink |
| Toggle column selection | ✅ | ❌ (holding shift-option and perform a cursor drag to column select, see detail in [Add support for column selection mode issue](https://github.com/zed-industries/zed/issues/7215)) | ✅ | N/A | N/A | Toggle column selection. | actions.selection.toggleColumnSelectionMode |

## Editor.Word

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Delete previous word | ✅ | ✅ | ✅ | N/A | N/A | Delete to the start of the previous word | actions.edit.deleteWordLeft |
| Delete previous subword | ✅ | ✅ | ❌ (behaviour is controlled by `CamelHumps` setting) | N/A | N/A | Delete to the start of the previous subword (hump) | actions.edit.deleteWordPartLeft |
| Delete next subword | ✅ | ✅ | ❌ (behaviour is controlled by `CamelHumps` setting) | N/A | N/A | Delete to the end of the next subword (hump) | actions.edit.deleteWordPartRight |
| Delete next word | ✅ | ✅ | ✅ | N/A | N/A | Delete to the end of the next word | actions.edit.deleteWordRight |

## File

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Close file | ✅ | ✅ | ✅ | ✅ | ✅ | Close the active editor | actions.file.closeEditor |
| New file | ✅ | ✅ | ✅ | N/A | N/A | Create a new file | actions.file.newFile |
| Open file | ✅ | ✅ | ✅ | N/A | ✅ | Open file dialog | actions.file.openFile |
| Open recent | ✅ | ✅ | ✅ | ✅ | ✅ | Open Recent | actions.file.openRecent |
| Save file | ✅ | ✅ | ✅ | N/A | ✅ | Save current file | actions.file.save |
| Save all | ✅ | ✅ | ✅ | N/A | ✅ | Save all open files | actions.file.saveAll |
| Show in new window | ✅ | ❌ (Not support, no issue tracked) | ✅ | N/A | N/A | Show opened file in new window | actions.file.showOpenedFileInNewWindow |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Save as | ✅ | ✅ | ❌ (intellij do not have save as, you can use `Save file`.) | N/A | N/A | Save current file with a new name | actions.file.saveAs | Use `Save file` instead. Not all editors support `Save as`. |
</details>

## Navigation

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Find next | ✅ | ✅ | ✅ | ✅ | ✅ | Find Next | actions.edit.nextMatchFindAction |
| Find previous | ✅ | ✅ | ✅ | N/A | ✅ | Find Previous | actions.edit.previousMatchFindAction |
| Jump to Navigation Bar | ✅ | ❌ (Not support, no issue tracked) | ✅ | N/A | N/A | Jump to Navigation Bar | actions.go.breadcrumbsFocus |
| Find file | ✅ | ✅ | ✅ | N/A | N/A | Go to file | actions.go.fileFinder |
| Go to line | ✅ | ✅ | ✅ | N/A | N/A | Go to Line/Column | actions.go.line |
| Find symbol | ✅ | ✅ | ✅ | N/A | N/A | Go to symbol in workspace, across files in the workspace | actions.go.symbolFinder |
| Find symbol in editor | ✅ | ✅ | ✅ | N/A | N/A | Go to symbol in current open editor | actions.go.symbolFinderInEditor |

## Navigation.DirtyDiff

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Next change | ✅ | ✅ | ✅ | N/A | N/A | Go to next change | actions.go.nextChange |
| Previous change | ✅ | ✅ | ✅ | N/A | N/A | Go to previous change | actions.go.previousChange |

## Navigation.History

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Go back | ✅ | ✅ | ✅ | ✅ | N/A | Go to previous cursor location | actions.go.back |
| Go forward | ✅ | ✅ | ✅ | ✅ | N/A | Go to next cursor location | actions.go.forward |
| Go to last edit location | ✅ | ❌ (not supported yet, see [Implement "Go To Last Edit Location" issue](https://github.com/zed-industries/zed/issues/19731)) | ✅ | N/A | N/A | Go to last edit location | actions.go.lastEditLocation |

## Navigation.Problems

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Next problem | ✅ | ✅ | ✅ | N/A | N/A | Go to Next Problem (Error, Warning, Info) | actions.go.nextProblem |
| Previous problem | ✅ | ✅ | ✅ | N/A | N/A | Go to Previous Problem (Error, Warning, Info) | actions.go.previousProblem |

## Redo & Undo

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Redo action | ✅ | ✅ | ✅ | N/A | ✅ | Redo last undone action | actions.edit.redo |
| Undo action | ✅ | ✅ | ✅ | N/A | ✅ | Undo last action | actions.edit.undo |

## Run

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Configure tasks | ✅ | ✅ | ✅ | N/A | N/A | Configure Task Runner | actions.run.configureTaskRunner |
| Re-run task | ✅ | ✅ | ✅ | N/A | N/A | Re-run last Task | actions.run.reRunTask |
| Run build task | ✅ | N/A | ✅ | ✅ | N/A | Run the default build task | actions.run.runBuildTask |
| Run task | ✅ | ✅ | ✅ | N/A | N/A | Run Task | actions.run.runTask |

## Terminal

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| New terminal | ✅ | ✅ | ✅ | N/A | N/A | Create a new terminal | actions.terminal.new |

## Tools.Diff

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Compare files | ✅ | ✅ | ✅ | N/A | N/A | Compare two files | actions.diff.compareTwoFiles |
| Next change | ✅ | ✅ | ✅ | N/A | N/A | Go to next change in compare editor | actions.diff.nextChange |
| Previous change | ✅ | ✅ | ✅ | N/A | N/A | Go to previous change in compare editor | actions.diff.previousChange |

## Tools.Jupyter Notebook

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Edit cell | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Edit Cell | actions.notebook.cell.edit |
| Execute cell | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Execute Cell | actions.notebook.cell.execute |
| Execute and insert | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Execute Cell and Insert Below | actions.notebook.cell.executeAndInsertBelow |
| Execute and select | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Execute Cell and Select Below | actions.notebook.cell.executeAndSelectBelow |
| Insert above | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Insert Code Cell Above | actions.notebook.cell.insertCodeCellAbove |
| Insert below | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Insert Code Cell Below | actions.notebook.cell.insertCodeCellBelow |
| Move down | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Move Cell Down | actions.notebook.cell.moveDown |
| Move up | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Move Cell Up | actions.notebook.cell.moveUp |
| Quit edit | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Stop Editing Cell | actions.notebook.cell.quitEdit |
| Focus bottom | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Focus Bottom | actions.notebook.focusBottom |
| Focus top | ✅ | ❌ (Notebook not supported yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | N/A | Focus Top | actions.notebook.focusTop |

## Version Control

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Jump to Source | ✅ | N/A | ✅ | N/A | N/A | Jump to Source | action.git.jumpSource |
| Commit all | ✅ | ✅ | ✅ | ✅ | N/A | Commit All | actions.git.commitAll |
| Open changes | ✅ | ✅ | ✅ | N/A | N/A | Open all git changed files | actions.git.openChanges |
| Push changes | ✅ | ✅ | ✅ | ✅ | N/A | Push Changes | actions.git.push |
| Revert changes | ✅ | ✅ | ✅ | N/A | N/A | Revert Changes | actions.git.revert |
| Stage changes | ✅ | ✅ | ✅ | N/A | N/A | Stage Changes | actions.git.stage |
| Pull changes | ✅ | ✅ | ✅ | ✅ | N/A | Pull changes | actions.git.sync |
| Toggle blame | ❌ (vscode support toggle blame inline, see `Toggle blame inline`) | ✅ | ❌ (intellij can only toggle blame in actions) | N/A | N/A | Toggle Blame in left of editor | actions.git.toggleBlame |
| Toggle blame inline | ✅ | ✅ | N/A | N/A | N/A | Toggle blame inline, next to editor content | actions.git.toggleBlameInline |
| Toggle blame status bar | ✅ | ❌ (not supported yet, see [`Optional Git Blame in status bar instead of inline` discussion](https://github.com/zed-industries/zed/discussions/26127)) | N/A | N/A | N/A | Toggle blame in status bar | actions.git.toggleBlameStatusBar |
| Unstage changes | ✅ | ✅ | ✅ | N/A | N/A | Unstage Changes | actions.git.unstage |
| Accept current | ✅ | N/A | ✅ | N/A | N/A | Accept current change (keep left side) | actions.merge.acceptCurrent |
| Accept incoming | ✅ | N/A | ✅ | N/A | N/A | Accept incoming change (take right side) | actions.merge.acceptIncoming |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Blame hover | ❌ (vscode support blame inline, see `Toggle blame inline`) | ✅ | N/A | N/A | N/A | Show blame information on hover | actions.git.blameHover | - |
</details>

## View Management

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Open global settings | ✅ | ✅ | ✅ | N/A | ✅ | Open Global Settings | actions.view.openGlobalSettings |
| Open keyboard shortcuts | ✅ | ✅ | ❌ (intellij do not have open keyboard shortcuts, you can open `Keymap` in command palette searching for `Keymap` and then open it.) | N/A | N/A | Open Keyboard Shortcuts Settings | actions.view.openKeyboardShortcuts |
| Select theme | ✅ | ✅ | ✅ | N/A | ✅ | Select Theme | actions.view.selectTheme |
| Show command palette | ✅ | ✅ | ✅ | N/A | ✅ | Show Command Palette | actions.view.showCommandPalette |
| Toggle bottom dock | ✅ | ✅ | N/A | N/A | N/A | Toggle Bottom Dock visibility | actions.view.toggleBottomDock |
| Toggle right sidebar | ✅ | ✅ | N/A | N/A | N/A | Toggle Right Side Bar visibility | actions.view.toggleRightSideBar |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Toggle status bar | ✅ | ❌ (Not support, see [Add options to hide title and status bar issue](https://github.com/zed-industries/zed/issues/5120)) | N/A | N/A | N/A | Toggle Status Bar visibility | actions.view.toggleStatusBar | - |
</details>

## View Management.Pannels

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Show extensions | ✅ | ✅ | ✅ | N/A | N/A | Show Extensions view | actions.view.showExtensions |
| Show testing | ✅ | ❌ (zed do not have testing view) | ✅ | N/A | N/A | Show Testing view | actions.view.showTesting |
| Toggle debug panel | ✅ | ✅ | ✅ | N/A | N/A | Toggle Debug Panel | actions.view.toggleDebugPanel |
| Toggle file explorer | ✅ | ✅ | ✅ | N/A | ✅ | Toggle file explorer view | actions.view.toggleExplorer |
| Toggle output | ✅ | ❌ (zed do not have output view) | ✅ | N/A | N/A | Toggle Output view | actions.view.toggleOutput |
| Toggle problems | ✅ | ✅ | ✅ | N/A | ✅ | Toggle Problems view | actions.view.toggleProblems |
| Toggle search | ✅ | ✅ | ✅ | N/A | ✅ | Toggle Search view | actions.view.toggleSearch |
| Toggle source control | ✅ | ✅ | ✅ | N/A | ✅ | Toggle Source Control view | actions.view.toggleSourceControl |
| Toggle terminal | ✅ | ✅ | ✅ | N/A | N/A | Toggle Terminal view | actions.view.toggleTerminal |

## View Management.Split

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Focus next split | ✅ | ✅ | ✅ | N/A | ✅ | Focus next editor split | actions.view.focusNextSplit |
| Focus previous split | ✅ | ✅ | ✅ | N/A | ✅ | Focus previous editor split | actions.view.focusPreviousSplit |
| Split down | ✅ | ✅ | ✅ | N/A | N/A | Split editor to down | actions.view.splitDown |
| Split right | ✅ | ✅ | ✅ | N/A | N/A | Split editor to right | actions.view.splitRight |

<details>
<summary>Featured Actions</summary>

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID | Featured Reason |
|--------|--------|-----|----------|-------|-------|-------------|-----------|-----------------|
| Split left | ✅ | ✅ | ❌ (intellij do not have split left, use `Split right` instead.) | N/A | N/A | Split editor to left | actions.view.splitLeft | Not all editors support split left, use `Split right` instead. |
| Split up | ✅ | ✅ | ❌ (intellij do not have split up, use `Split down` instead.) | N/A | N/A | Split editor to up | actions.view.splitUp | Not all editors support split up, use `Split down` instead. |
</details>

## View Management.Tab

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Next tab | ✅ | ✅ | ✅ | ✅ | N/A | Switch to next tab | actions.tabSwitcher.next |
| Previous tab | ✅ | ✅ | ✅ | ✅ | N/A | Switch to previous tab | actions.tabSwitcher.previous |

## View Management.Window

| Action | VSCode | Zed | IntelliJ | Xcode | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------|-------------|-----------|
| Close window | ✅ | ✅ | ✅ | N/A | ✅ | Close the current window | actions.file.closeWindow |
| New window | ✅ | ✅ | N/A | ✅ | N/A | Open a new window | actions.file.newWindow |
| Maximize editor | ✅ | N/A | ✅ | N/A | N/A | Maximize editor (hide other windows) | actions.view.maximizeEditor |
| Toggle full screen | ✅ | ✅ | ✅ | N/A | N/A | Toggle full screen | actions.view.toggleFullScreen |
