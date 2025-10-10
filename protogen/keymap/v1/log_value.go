package keymapv1

import (
	"log/slog"
	"strings"
)

// compactKeyBinding returns a minimal view to avoid leaking full content in logs.
func compactKeyBinding(kb *Action) map[string]any {
	if kb == nil {
		return nil
	}

	var readables []string
	for _, b := range kb.GetBindings() {
		if b != nil && b.GetKeyChordsReadable() != "" {
			readables = append(readables, b.GetKeyChordsReadable())
		}
	}

	return map[string]any{
		"name":     kb.GetName(),
		"bindings": strings.Join(readables, " or "),
	}
}

func compactKeybindingsList(list []*Action) []map[string]any {
	if len(list) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(list))
	for _, kb := range list {
		out = append(out, compactKeyBinding(kb))
	}
	return out
}

func actionCount(ks *Keymap) int {
	if ks == nil {
		return 0
	}
	return len(ks.GetActions())
}

// LogValue implements slog.LogValuer for *ImportKeymapRequest.
func (req *ImportKeymapRequest) LogValue() slog.Value {
	if req == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.String("EditorType", req.GetEditorType().String()),
		slog.Int("SourceLen", len(req.GetSource())),
		slog.Int("BaseLen", len(req.GetBase())),
	)
}

// LogValue implements slog.LogValuer for *ImportKeymapResponse
// keymap: only print number of keybindings
// changes: for KeyBinding, only print id and key_chords_readable
func (resp *ImportKeymapResponse) LogValue() slog.Value {
	if resp == nil {
		return slog.Value{}
	}

	// build compact changes
	var addList []map[string]any
	var removeList []map[string]any
	var updateList []map[string]any

	if ch := resp.GetChanges(); ch != nil {
		for _, kb := range ch.GetAdd() {
			addList = append(addList, compactKeyBinding(kb))
		}
		for _, kb := range ch.GetRemove() {
			removeList = append(removeList, compactKeyBinding(kb))
		}
		for _, diff := range ch.GetUpdate() {
			updateList = append(updateList, map[string]any{
				"origin":  compactKeyBinding(diff.GetOrigin()),
				"updated": compactKeyBinding(diff.GetUpdated()),
			})
		}
	}

	return slog.GroupValue(
		slog.Int("ActionsCount", actionCount(resp.GetKeymap())),
		slog.Group("Changes",
			slog.Any("Add", addList),
			slog.Any("Remove", removeList),
			slog.Any("Update", updateList),
		),
	)
}

// LogValue implements slog.LogValuer for *GetKeymapRequest
// config: only print length to avoid logging full content
func (req *GetKeymapRequest) LogValue() slog.Value {
	if req == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.Int("ConfigLen", len(req.GetConfig())),
		slog.Bool("ReturnAll", req.GetReturnAll()),
	)
}

// LogValue implements slog.LogValuer for *GetKeymapResponse
// keymap: only compact keybindings {id, key_chords_readable}
func (resp *GetKeymapResponse) LogValue() slog.Value {
	if resp == nil {
		return slog.Value{}
	}
	ks := resp.GetKeymap()
	var compact []map[string]any
	if ks != nil {
		compact = compactKeybindingsList(ks.GetActions())
	}
	return slog.GroupValue(
		slog.Int("ActionsCount", actionCount(ks)),
		slog.Any("Actions", compact),
	)
}

// LogValue implements slog.LogValuer for *SaveKeymapRequest
// keymap: only compact keybindings {id, key_chords_readable}
func (req *SaveKeymapRequest) LogValue() slog.Value {
	if req == nil {
		return slog.Value{}
	}
	ks := req.GetKeymap()
	var compact []map[string]any
	if ks != nil {
		compact = compactKeybindingsList(ks.GetActions())
	}
	return slog.GroupValue(
		slog.Int("ActionsCount", actionCount(ks)),
		slog.Any("Actions", compact),
	)
}

// LogValue implements slog.LogValuer for *SaveKeymapResponse
// config: only print length to avoid logging full content
func (resp *SaveKeymapResponse) LogValue() slog.Value {
	if resp == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.Int("ConfigLen", len(resp.GetConfig())),
	)
}

// LogValue implements slog.LogValuer for *ExportKeymapRequest
// Only log diff_type and editor_type.
func (req *ExportKeymapRequest) LogValue() slog.Value {
	if req == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.String("EditorType", req.GetEditorType().String()),
		slog.String("DiffType", req.GetDiffType().String()),
		slog.Int("BaseLen", len(req.GetBase())),
		slog.String("Base", req.GetFilePath()),
	)
}

// LogValue implements slog.LogValuer for *ExportKeymapResponse
// Only log diff.
func (resp *ExportKeymapResponse) LogValue() slog.Value {
	if resp == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.String("Diff", resp.GetDiff()),
	)
}
