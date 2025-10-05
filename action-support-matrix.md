# Action Support Matrix

| Action | VSCode | Zed | IntelliJ | Helix | Description | Action ID |
|--------|--------|-----|----------|-------|-------------|-----------|
| Copy text | ✅ | ✅ | ✅ | ✅ | Copy selected text/file | actions.clipboard.copy |
| Copy file path | ✅ | ✅ | ✅ | N/A | Copy file path | actions.clipboard.copyFilePath |
| Cut text | ✅ | ✅ | ✅ | ✅ | Cut selected text/file | actions.clipboard.cut |
| Paste text | ✅ | ✅ | ✅ | ✅ | Paste text/file | actions.clipboard.paste |
| Move to line end | ✅ | ✅ | ✅ | ✅ | Move cursor to the end of the line | actions.cursor.lineEnd |
| Select line end | ✅ | ✅ | ✅ | ✅ | Select from cursor to the end of the line | actions.cursor.lineEndSelect |
| Move to line start | ✅ | ✅ | ✅ | ✅ | Move cursor to the beginning of the line | actions.cursor.lineStart |
| Select line start | ✅ | ✅ | ✅ | ✅ | Select from cursor to the beginning of the line | actions.cursor.lineStartSelect |
| Move to bottom | ✅ | ✅ | ✅ | ✅ | Move caret to text end | actions.cursor.moveToBottom |
| Select to bottom | ✅ | ✅ | ✅ | ✅ | Select from cursor to text end | actions.cursor.moveToBottomSelect |
| Move to top | ✅ | ✅ | ✅ | ✅ | Move caret to text start | actions.cursor.moveToTop |
| Select to top | ✅ | ✅ | ✅ | ✅ | Select from cursor to text start | actions.cursor.moveToTopSelect |
| Page down | ✅ | ✅ | ✅ | ✅ | Move cursor down by one page | actions.cursor.pageDown |
| Select page down | ✅ | ✅ | ✅ | N/A | Select down by one page | actions.cursor.pageDownSelect |
| Page up | ✅ | ✅ | ✅ | ✅ | Move cursor up by one page | actions.cursor.pageUp |
| Select page up | ✅ | ✅ | ✅ | N/A | Select up by one page | actions.cursor.pageUpSelect |
| Move to previous word | ✅ | ✅ | ✅ | ✅ | Move cursor to the start of the previous word | actions.cursor.wordLeft |
| Select previous word | ✅ | ✅ | ✅ | ❌ (helix move cursor with selection by default) | Select to the start of the previous word | actions.cursor.wordLeftSelect |
| Move to previous subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | ✅ | Move cursor to the start of the previous subword (hump) | actions.cursor.wordPartLeft |
| Select previous subword | ✅ | ✅ | ❌ (intellij need to turn on `CamelHumps` setting) | ❌ (helix move cursor with selection by default) | Select to the start of the previous subword (hump) | actions.cursor.wordPartLeftSelect |
| Move to next subword | ✅ | ✅ | N/A | ✅ | Move cursor to the end of the next subword (hump) | actions.cursor.wordPartRight |
| Select next subword | ✅ | ✅ | N/A | N/A | Select to the end of the next subword (hump) | actions.cursor.wordPartRightSelect |
| Move to next word | ✅ | ✅ | ✅ | ✅ | Move cursor to the end of the next word | actions.cursor.wordRight |
| Select next word | ✅ | ✅ | ✅ | ❌ (helix move cursor with selection by default) | Select to the end of the next word | actions.cursor.wordRightSelect |
| Compare files | ✅ | ✅ | ✅ | N/A | Compare two files | actions.diff.compareTwoFiles |
| Next change | ✅ | ✅ | ✅ | N/A | Go to next change in compare editor | actions.diff.nextChange |
| Previous change | ✅ | ✅ | ✅ | N/A | Go to previous change in compare editor | actions.diff.previousChange |
| Undo cursor | ✅ | ✅ | ✅ | N/A | Undo last cursor operation | actions.edit.cursorUndo |
| Delete line | ✅ | ✅ | ✅ | N/A | Delete line | actions.edit.deleteLines |
| Delete previous word | ✅ | ✅ | ✅ | N/A | Delete to the start of the previous word | actions.edit.deleteWordLeft |
| Delete previous subword | ✅ | ✅ | N/A | N/A | Delete to the start of the previous subword (hump) | actions.edit.deleteWordPartLeft |
| Delete next subword | ✅ | ✅ | N/A | N/A | Delete to the end of the next subword (hump) | actions.edit.deleteWordPartRight |
| Find in file | ✅ | ✅ | ✅ | ✅ | Find in current file | actions.edit.find |
| Find in project | ✅ | ✅ | ✅ | ✅ | Find in all files in the project | actions.edit.findInFiles |
| Format document | ✅ | ✅ | ✅ | N/A | Format Document | actions.edit.formatDocument |
| Format selection | ✅ | ✅ | ✅ | N/A | Format Selection | actions.edit.formatSelection |
| Next suggestion | ✅ | ✅ | ✅ | N/A | Show next inline suggestion | actions.edit.inlineSuggest.next |
| Previous suggestion | ✅ | ✅ | ✅ | N/A | Show previous inline suggestion | actions.edit.inlineSuggest.previous |
| Show inline suggestion | ✅ | ✅ | ✅ | N/A | Show inline suggestion | actions.edit.inlineSuggest.show |
| Insert line after | ✅ | ✅ | ✅ | N/A | Insert a new line after the current line | actions.edit.insertLineAfter |
| Insert line before | ✅ | ✅ | ✅ | N/A | Insert a new line before the current line | actions.edit.insertLineBefore |
| Join lines | ✅ | ✅ | ✅ | N/A | Join lines | actions.edit.joinLines |
| Insert line break | ✅ | ✅ | ✅ | N/A | Insert line break | actions.edit.lineBreakInsert |
| Find next | ✅ | ✅ | ✅ | ✅ | Find Next | actions.edit.nextMatchFindAction |
| Find previous | ✅ | ✅ | ✅ | ✅ | Find Previous | actions.edit.previousMatchFindAction |
| Redo action | ✅ | ✅ | ✅ | ✅ | Redo last undone action | actions.edit.redo |
| Replace in file | ✅ | ✅ | ✅ | N/A | Replace in current file | actions.edit.replace |
| Replace in project | ✅ | ✅ | ✅ | N/A | Replace in all files in the project | actions.edit.replaceInFiles |
| Show suggestions | ✅ | ✅ | ✅ | N/A | Trigger Suggest | actions.edit.suggest.show |
| Toggle block comment | ✅ | N/A | ✅ | ✅ | Toggle block comment | actions.edit.toggleBlockComment |
| Toggle line comment | ✅ | ✅ | ✅ | ✅ | Toggle line comment | actions.edit.toggleLineComment |
| Undo action | ✅ | ✅ | ✅ | ✅ | Undo last action | actions.edit.undo |
| Close file | ✅ | ✅ | ✅ | ✅ | Close the active editor | actions.file.closeEditor |
| Close window | ✅ | ✅ | ✅ | ✅ | Close the current window | actions.file.closeWindow |
| New file | ✅ | ✅ | ✅ | N/A | Create a new file | actions.file.newFile |
| New window | ✅ | ✅ | N/A | N/A | Open a new window | actions.file.newWindow |
| Open file | ✅ | ✅ | ✅ | ✅ | Open file dialog | actions.file.openFile |
| Open recent | ✅ | ✅ | ✅ | ✅ | Open Recent | actions.file.openRecent |
| Save file | ✅ | ✅ | ✅ | ✅ | Save current file | actions.file.save |
| Save all | ✅ | ✅ | ✅ | ✅ | Save all open files | actions.file.saveAll |
| Save as | ✅ | ✅ | N/A | N/A | Save current file with a new name | actions.file.saveAs |
| Show in new window | ✅ | N/A | ✅ | N/A | Show opened file in new window | actions.file.showOpenedFileInNewWindow |
| Fold | ✅ | ✅ | ✅ | N/A | Collapse the current code block | actions.fold.fold |
| Fold all | ✅ | ✅ | ✅ | N/A | Collapse all code blocks in the editor | actions.fold.foldAll |
| Fold recursively | ✅ | ✅ | ✅ | N/A | Collapse the current code block and its children recursively | actions.fold.foldRecursively |
| Toggle fold | ✅ | ✅ | ✅ | N/A | Toggle Fold | actions.fold.toggleFold |
| Unfold | ✅ | ✅ | ✅ | N/A | Expand the current code block | actions.fold.unfold |
| Unfold all | ✅ | ✅ | ✅ | N/A | Expand all code blocks in the editor | actions.fold.unfoldAll |
| Unfold recursively | ✅ | ✅ | ✅ | N/A | Expand the current code block and its children recursively | actions.fold.unfoldRecursively |
| Commit all | ✅ | ✅ | ✅ | N/A | Commit All | actions.git.commitAll |
| Open changes | ✅ | ✅ | ✅ | N/A | Open all git changed files | actions.git.openChanges |
| Push changes | ✅ | ✅ | ✅ | N/A | Push Changes | actions.git.push |
| Revert changes | ✅ | ✅ | ✅ | N/A | Revert Changes | actions.git.revert |
| Stage changes | ✅ | ✅ | ✅ | N/A | Stage Changes | actions.git.stage |
| Sync changes | ✅ | ✅ | ✅ | N/A | Sync (Pull, Push) | actions.git.sync |
| Toggle blame | N/A | ✅ | N/A | N/A | Toggle Blame | actions.git.toggleBlame |
| Unstage changes | ✅ | ✅ | ✅ | N/A | Unstage Changes | actions.git.unstage |
| Go back | ✅ | ✅ | ✅ | N/A | Go to previous cursor location | actions.go.back |
| Go to bracket | ✅ | ✅ | ✅ | N/A | Go to Bracket | actions.go.bracket |
| Focus breadcrumbs | ✅ | N/A | ✅ | N/A | Jump to Navigation Bar | actions.go.breadcrumbsFocus |
| Call hierarchy | ✅ | N/A | ✅ | N/A | Show Call Hierarchy | actions.go.callHierarchy |
| Find file | ✅ | ✅ | ✅ | N/A | Go to File | actions.go.fileFinder |
| Go forward | ✅ | ✅ | ✅ | N/A | Go to next cursor location | actions.go.forward |
| Go to declaration | ✅ | ✅ | ✅ | N/A | Go to Declaration or Usages | actions.go.goToDeclaration |
| Go to super | ✅ | N/A | ✅ | N/A | Go to Super Class/Super Method | actions.go.goToSuper |
| Go to test | ✅ | N/A | ✅ | N/A | Go to Test | actions.go.goToTest |
| Go to implementations | ✅ | ✅ | ✅ | N/A | Go to Implementations | actions.go.implementations |
| Go to last edit location | ✅ | N/A | ✅ | N/A | Go to last edit location | actions.go.lastEditLocation |
| Go to line | ✅ | ✅ | ✅ | N/A | Go to Line/Column | actions.go.line |
| Next change | ✅ | ✅ | ✅ | N/A | Go to Next Change | actions.go.nextChange |
| Next problem | ✅ | ✅ | ✅ | N/A | Go to Next Problem (Error, Warning, Info) | actions.go.nextProblem |
| Peek declaration | ✅ | ✅ | ✅ | N/A | Peek Declaration | actions.go.peekDeclaration |
| Previous change | ✅ | ✅ | ✅ | N/A | Go to Previous Change | actions.go.previousChange |
| Previous problem | ✅ | ✅ | ✅ | N/A | Go to Previous Problem (Error, Warning, Info) | actions.go.previousProblem |
| Reference peek | ✅ | N/A | ✅ | N/A | Show Usages / Reference Search | actions.go.referencePeek |
| Go to references | ✅ | ✅ | ✅ | N/A | Go to References | actions.go.references |
| Find symbol | ✅ | ✅ | ✅ | N/A | Go to Symbol in Workspace | actions.go.symbolFinder |
| Find symbol in editor | ✅ | ✅ | ✅ | N/A | Go to Symbol in Editor | actions.go.symbolFinderInEditor |
| Go to type definition | ✅ | ✅ | ✅ | N/A | Go to Type Definition | actions.go.typeDefinition |
| Type hierarchy | ✅ | N/A | ✅ | N/A | Show Type Hierarchy | actions.go.typeHierarchy |
| Show hover | ✅ | ✅ | ✅ | ✅ | Show Hover | actions.hover.showHover |
| Accept current | ✅ | N/A | ✅ | N/A | Accept current change (keep left side) | actions.merge.acceptCurrent |
| Accept incoming | ✅ | N/A | ✅ | N/A | Accept incoming change (take right side) | actions.merge.acceptIncoming |
| Edit cell | ✅ | ❌ (zed do not support jupyter notebook yet see [issue](https://github.com/zed-industries/zed/discussions/25936)) | ✅ | N/A | Edit Cell | actions.notebook.cell.edit |
| Execute cell | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Execute Cell | actions.notebook.cell.execute |
| Execute and insert | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Execute Cell and Insert Below | actions.notebook.cell.executeAndInsertBelow |
| Execute and select | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Execute Cell and Select Below | actions.notebook.cell.executeAndSelectBelow |
| Insert above | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Insert Code Cell Above | actions.notebook.cell.insertCodeCellAbove |
| Insert below | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Insert Code Cell Below | actions.notebook.cell.insertCodeCellBelow |
| Move down | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Move Cell Down | actions.notebook.cell.moveDown |
| Move up | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Move Cell Up | actions.notebook.cell.moveUp |
| Quit edit | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Stop Editing Cell | actions.notebook.cell.quitEdit |
| Focus bottom | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Focus Bottom | actions.notebook.focusBottom |
| Focus top | ✅ | ❌ (jupyter notebook yet) | ✅ | N/A | Focus Top | actions.notebook.focusTop |
| Code action | ✅ | ✅ | ✅ | ✅ | Code Action... | actions.refactor.codeAction |
| Organize imports | ✅ | ✅ | ✅ | N/A | Organize Imports | actions.refactor.organizeImports |
| Quick fix | ✅ | N/A | N/A | N/A | Quick Fix... | actions.refactor.quickFix |
| Refactor code | ✅ | N/A | ✅ | ✅ | Refactor This... | actions.refactor.refactor |
| Rename symbol | ✅ | ✅ | ✅ | ✅ | Rename | actions.refactor.rename |
| Parameter hints | ✅ | ✅ | ✅ | ✅ | Trigger Parameter Hints | actions.refactor.triggerParameterHint |
| Configure tasks | ✅ | ✅ | ✅ | N/A | Configure Task Runner | actions.run.configureTaskRunner |
| Continue | ✅ | ✅ | ✅ | ✅ | Continue | actions.run.continue |
| Re-run task | ✅ | ✅ | ✅ | N/A | Re-run last Task | actions.run.reRunTask |
| Restart debugging | ✅ | ✅ | ✅ | ✅ | Restart Debugging | actions.run.restartDebugging |
| Run task | ✅ | ✅ | ✅ | N/A | Run Task | actions.run.runTask |
| Run to cursor | ✅ | ✅ | ✅ | N/A | Run to Cursor | actions.run.runToCursor |
| Evaluate selection | ✅ | ✅ | ✅ | N/A | Send selection to REPL | actions.run.selectionToRepl |
| Start debugging | ✅ | ✅ | ✅ | ✅ | Start Debugging | actions.run.startDebugging |
| Step into | ✅ | ✅ | ✅ | ✅ | Step Into | actions.run.stepInto |
| Step out | ✅ | ✅ | ✅ | ✅ | Step Out | actions.run.stepOut |
| Step over | ✅ | ✅ | ✅ | ✅ | Step Over | actions.run.stepOver |
| Stop debugging | ✅ | ✅ | ✅ | ✅ | Stop Debugging | actions.run.stopDebugging |
| Toggle breakpoint | ✅ | ✅ | ✅ | ✅ | Toggle Breakpoint | actions.run.toggleBreakpoint |
| Add cursor above | ✅ | ✅ | ✅ | ✅ | Add cursor above current line | actions.selection.addCursorAbove |
| Add cursor below | ✅ | ✅ | ✅ | ✅ | Add cursor below current line | actions.selection.addCursorBelow |
| Add cursors to ends | ✅ | N/A | ✅ | N/A | Add cursors to the end of selected lines | actions.selection.addCursorsToLineEnds |
| Add next occurrence | ✅ | ✅ | ✅ | ✅ | Add next occurrence of selection to multicursor | actions.selection.addNextOccurrence |
| Add previous occurrence | ✅ | ✅ | ✅ | ✅ | Add previous occurrence of selection to multicursor | actions.selection.addPreviousOccurrence |
| Copy line down | ✅ | ✅ | ✅ | ✅ | Copy current line down | actions.selection.copyLineDown |
| Copy line up | ✅ | ✅ | N/A | ✅ | Copy current line up | actions.selection.copyLineUp |
| Expand selection | ✅ | ✅ | ✅ | ✅ | Expand selection | actions.selection.expand |
| Move line down | ✅ | ✅ | ✅ | N/A | Move current line down | actions.selection.moveLineDown |
| Move line up | ✅ | ✅ | ✅ | ✅ | Move current line up | actions.selection.moveLineUp |
| Select all | ✅ | ✅ | ✅ | ✅ | Select all text in the editor | actions.selection.selectAll |
| Select all occurrences | ✅ | ✅ | ✅ | N/A | Select all occurrences of current selection | actions.selection.selectAllOccurrences |
| Shrink selection | ✅ | ✅ | ✅ | ✅ | Shrink selection | actions.selection.shrink |
| Toggle column mode | ✅ | N/A | ✅ | N/A | Toggle column selection mode | actions.selection.toggleColumnSelectionMode |
| Next tab | ✅ | ✅ | ✅ | N/A | Switch to next tab | actions.tabSwitcher.next |
| Previous tab | ✅ | ✅ | ✅ | N/A | Switch to previous tab | actions.tabSwitcher.previous |
| New terminal | ✅ | ✅ | ✅ | N/A | Create a new terminal | actions.terminal.new |
| Run build task | ✅ | N/A | ✅ | N/A | Run the default build task | actions.terminal.runBuildTask |
| Focus next split | ✅ | ✅ | ✅ | ✅ | Focus next editor split | actions.view.focusNextSplit |
| Focus previous split | ✅ | ✅ | ✅ | ✅ | Focus previous editor split | actions.view.focusPreviousSplit |
| Maximize editor | ✅ | N/A | ✅ | N/A | Maximize editor (hide other windows) | actions.view.maximizeEditor |
| Open global settings | ✅ | ✅ | ✅ | ✅ | Open Global Settings | actions.view.openGlobalSettings |
| Select theme | ✅ | ✅ | ✅ | ✅ | Select Theme | actions.view.selectTheme |
| Show command palette | ✅ | ✅ | ✅ | ✅ | Show Command Palette | actions.view.showCommandPalette |
| Show debug console | ✅ | N/A | ❌ (intellij have debug output with DebugPanel) | N/A | Show Debug Output Console view | actions.view.showDebugOutputConsole |
| Show extensions | ✅ | ✅ | N/A | N/A | Show Extensions view | actions.view.showExtensions |
| Show testing | ✅ | N/A | ✅ | N/A | Show Testing view | actions.view.showTesting |
| Split down | ✅ | ✅ | N/A | N/A | Split editor to down | actions.view.splitDown |
| Split left | ✅ | ✅ | N/A | N/A | Split editor to left | actions.view.splitLeft |
| Split right | ✅ | ✅ | ✅ | N/A | Split editor to right | actions.view.splitRight |
| Split up | ✅ | ✅ | N/A | N/A | Split editor to up | actions.view.splitUp |
| Toggle bottom dock | ✅ | ✅ | N/A | N/A | Toggle Bottom Dock visibility | actions.view.toggleBottomDock |
| Toggle debug panel | ✅ | ✅ | ✅ | N/A | Toggle Debug Panel | actions.view.toggleDebugPanel |
| Toggle file explorer | ✅ | ✅ | ✅ | ✅ | Toggle file explorer view | actions.view.toggleExplorer |
| Toggle full screen | ✅ | ✅ | ✅ | N/A | Toggle full screen | actions.view.toggleFullScreen |
| Toggle output | ✅ | N/A | ✅ | N/A | Toggle Output view | actions.view.toggleOutput |
| Toggle problems | ✅ | ✅ | ✅ | ✅ | Toggle Problems view | actions.view.toggleProblems |
| Toggle right sidebar | ✅ | ✅ | N/A | N/A | Toggle Right Side Bar visibility | actions.view.toggleRightSideBar |
| Toggle search | ✅ | ✅ | ✅ | ✅ | Toggle Search view | actions.view.toggleSearch |
| Toggle source control | ✅ | ✅ | ✅ | ✅ | Toggle Source Control view | actions.view.toggleSourceControl |
| Toggle status bar | ✅ | N/A | N/A | N/A | Toggle Status Bar visibility | actions.view.toggleStatusBar |
| Toggle terminal | ✅ | ✅ | ✅ | N/A | Toggle Terminal view | actions.view.toggleTerminal |
| Toggle word wrap | ✅ | ✅ | ❌ (intellij need configure in settings) | N/A | Toggle word wrap in the editor | actions.view.toggleWordWrap |
