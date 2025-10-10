package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/validateapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
	"google.golang.org/protobuf/proto"
)

// importService is the default implementation of the Importer interface.
type importService struct {
	registry      *plugins.Registry
	mappingConfig *mappings.MappingConfig
	logger        *slog.Logger
	validator     *validateapi.Validator
	recorder      metrics.Recorder
}

// NewImportService creates a new default import service.
func NewImportService(
	registry *plugins.Registry,
	config *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) importapi.Importer {
	service := &importService{
		registry:      registry,
		mappingConfig: config,
		logger:        logger,
		validator: validateapi.NewValidator(
			validateapi.NewKeybindConflictRule(),
			validateapi.NewDanglingActionRule(config),
		),
		recorder: recorder,
	}

	return service
}

// Import is the method implementation for the default service.
func (s *importService) Import(ctx context.Context, opts importapi.ImportOptions) (*importapi.ImportResult, error) {
	if opts.InputStream == nil {
		return nil, errors.New("input stream is required")
	}
	plugin, ok := s.registry.Get(opts.EditorType)
	if !ok {
		return nil, fmt.Errorf("no plugin found for editor type '%s'", opts.EditorType)
	}

	importer, err := plugin.Importer()
	if err != nil {
		return nil, fmt.Errorf("failed to get importer for %s: %w", opts.EditorType, err)
	}
	setting, err := importer.Import(ctx, opts.InputStream, pluginapi.PluginImportOption{})
	if err != nil {
		return nil, fmt.Errorf("failed to import config: %w", err)
	}

	setting = keymap.DecorateSetting(setting, s.mappingConfig)
	// Normalize: merge same-action entries and deduplicate identical bindings before downstream logic
	if setting != nil && len(setting.GetActions()) > 0 {
		setting.Actions = dedupKeyBindings(setting.GetActions())
	}

	s.recorder.RecordCommandProcessed(ctx, string(opts.EditorType), setting)

	// Sort by action for determinism
	sort.Slice(setting.GetActions(), func(i, j int) bool {
		return setting.GetActions()[i].GetName() < setting.GetActions()[j].GetName()
	})

	if setting == nil {
		return nil, errors.New("failed to import config: no keybindings found")
	}

	report, err := s.validator.Validate(ctx, setting, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// No baseline provided: all imported keymaps are additions.
	if len(opts.Base.GetActions()) == 0 {
		changes := &importapi.KeymapChanges{}
		if len(setting.GetActions()) > 0 {
			changes.Add = append(changes.Add, setting.GetActions()...)
		}
		return &importapi.ImportResult{Setting: setting, Changes: changes, Report: report}, nil
	}

	// If baseline provided, first union baseline chords into current setting so unchanged chords are retained.
	setting = unionWithBase(opts.Base, setting)
	setting.Actions = dedupKeyBindings(setting.GetActions())
	// Re-decorate after union so metadata (Name/Description/Category) and readable chords are present
	setting = keymap.DecorateSetting(setting, s.mappingConfig)

	// With baseline: compute changes via helper.
	changes := s.calculateChanges(opts.Base, setting)

	// Safety: ensure dedup on output as well
	setting.Actions = dedupKeyBindings(setting.GetActions())
	return &importapi.ImportResult{Setting: setting, Changes: changes, Report: report}, nil
}

func (s *importService) calculateChanges(
	base *keymapv1.Keymap,
	setting *keymapv1.Keymap,
) *importapi.KeymapChanges {
	baseIndex := s.buildActionIndex(base)
	newIndex := s.buildActionIndex(setting)

	adds, removes := s.calculateAddsAndRemoves(baseIndex.byPair, newIndex.byPair)
	updates := s.calculateUpdates(baseIndex.byAction, newIndex.byAction, adds, removes)

	changes := s.buildChangesResult(adds, removes, updates)
	s.decorateChanges(changes)

	return changes
}

type actionIndex struct {
	byAction map[string][]*keymapv1.Action
	byPair   map[string]*keymapv1.Action
}

func (s *importService) buildActionIndex(keymap *keymapv1.Keymap) *actionIndex {
	idx := &actionIndex{
		byAction: map[string][]*keymapv1.Action{},
		byPair:   map[string]*keymapv1.Action{},
	}

	for _, kb := range keymap.GetActions() {
		if !hasValidChord(kb) {
			continue
		}
		idx.byAction[kb.GetName()] = append(idx.byAction[kb.GetName()], kb)
		idx.byPair[pairKey(kb)] = kb
	}

	return idx
}

func (s *importService) calculateAddsAndRemoves(
	basePair map[string]*keymapv1.Action,
	newPair map[string]*keymapv1.Action,
) (map[string]*keymapv1.Action, map[string]*keymapv1.Action) {
	adds := map[string]*keymapv1.Action{}
	removes := map[string]*keymapv1.Action{}

	for k, v := range newPair {
		adds[k] = v
	}

	for k := range basePair {
		if _, ok := adds[k]; ok {
			delete(adds, k)
			continue
		}
		removes[k] = basePair[k]
	}

	return adds, removes
}

func (s *importService) calculateUpdates(
	baseByAction map[string][]*keymapv1.Action,
	newByAction map[string][]*keymapv1.Action,
	adds map[string]*keymapv1.Action,
	removes map[string]*keymapv1.Action,
) []importapi.KeymapDiff {
	var updates []importapi.KeymapDiff

	for action, beforeList := range baseByAction {
		afterList, ok := newByAction[action]
		if !ok || len(beforeList) != 1 || len(afterList) != 1 {
			continue
		}

		before := beforeList[0]
		after := afterList[0]
		if pairKey(before) == pairKey(after) {
			continue
		}

		updates = append(updates, importapi.KeymapDiff{Before: before, After: after})
		delete(adds, pairKey(after))
		delete(removes, pairKey(before))
	}

	return updates
}

func (s *importService) buildChangesResult(
	adds map[string]*keymapv1.Action,
	removes map[string]*keymapv1.Action,
	updates []importapi.KeymapDiff,
) *importapi.KeymapChanges {
	changes := &importapi.KeymapChanges{Update: updates}

	if len(adds) > 0 {
		changes.Add = make([]*keymapv1.Action, 0, len(adds))
		for _, v := range adds {
			changes.Add = append(changes.Add, v)
		}
		sort.Slice(changes.Add, func(i, j int) bool {
			return changes.Add[i].GetName() < changes.Add[j].GetName()
		})
	}

	if len(removes) > 0 {
		changes.Remove = make([]*keymapv1.Action, 0, len(removes))
		for _, v := range removes {
			changes.Remove = append(changes.Remove, v)
		}
		sort.Slice(changes.Remove, func(i, j int) bool {
			return changes.Remove[i].GetName() < changes.Remove[j].GetName()
		})
	}

	return changes
}

func (s *importService) decorateChanges(changes *importapi.KeymapChanges) {
	for _, kb := range changes.Add {
		s.decorateAction(kb)
	}
	for _, kb := range changes.Remove {
		s.decorateAction(kb)
	}
	for i := range changes.Update {
		s.decorateAction(changes.Update[i].Before)
		s.decorateAction(changes.Update[i].After)
	}
}

func (s *importService) decorateAction(kb *keymapv1.Action) {
	if kb == nil {
		return
	}

	if cfg := s.mappingConfig.FindByUniversalAction(kb.GetName()); cfg != nil {
		if kb.GetActionConfig() == nil {
			kb.ActionConfig = &keymapv1.ActionConfig{}
		}
		kb.ActionConfig.Description = cfg.Description
		kb.ActionConfig.DisplayName = cfg.Name
		kb.ActionConfig.Category = cfg.Category
	}

	for _, b := range kb.GetBindings() {
		if b != nil && b.GetKeyChords() != nil && len(b.GetKeyChords().GetChords()) > 0 {
			if formatted, err := keymap.NewKeyBinding(b).Format(platform.PlatformMacOS, "+"); err == nil {
				b.KeyChordsReadable = formatted
			}
		}
	}
}

func hasValidChord(kb *keymapv1.Action) bool {
	if kb == nil {
		return false
	}
	for _, b := range kb.GetBindings() {
		if b != nil && b.GetKeyChords() != nil && len(b.GetKeyChords().GetChords()) > 0 {
			return true
		}
	}
	return false
}

// unionWithBase merges baseline and imported settings per action id,
// preserving baseline bindings and adding new imported bindings. Metadata from Decorate is preserved.
func unionWithBase(base *keymapv1.Keymap, imported *keymapv1.Keymap) *keymapv1.Keymap {
	if imported == nil {
		return base
	}
	if base == nil || len(base.GetActions()) == 0 {
		return imported
	}
	// index existing results by action id
	out := &keymapv1.Keymap{Actions: []*keymapv1.Action{}}
	byID := make(map[string]*keymapv1.Action)

	// start with baseline (so Before reflects baseline order/first occurrence)
	for _, kb := range base.GetActions() {
		if kb == nil {
			continue
		}
		// shallow copy action binding shell; reuse bindings slice ref (safe; will be appended below possibly)
		ab := &keymapv1.Action{
			Name: kb.GetName(),
		}
		// Only clone ActionConfig if it exists
		if kb.GetActionConfig() != nil {
			cloned := proto.Clone(kb.GetActionConfig())
			if ac, ok := cloned.(*keymapv1.ActionConfig); ok {
				ab.ActionConfig = ac
			}
		}
		// copy bindings
		for _, b := range kb.GetBindings() {
			if b != nil {
				ab.Bindings = append(ab.Bindings, b)
			}
		}
		byID[ab.GetName()] = ab
		out.Actions = append(out.Actions, ab)
	}
	// merge imported bindings into corresponding actions (or create new action entries)
	for _, kb := range imported.GetActions() {
		if kb == nil {
			continue
		}
		existing, ok := byID[kb.GetName()]
		if !ok {
			// add as new action
			ab := &keymapv1.Action{
				Name: kb.GetName(),
			}
			// Only clone ActionConfig if it exists
			if kb.GetActionConfig() != nil {
				cloned := proto.Clone(kb.GetActionConfig())
				if ac, ok := cloned.(*keymapv1.ActionConfig); ok {
					ab.ActionConfig = ac
				}
			}
			for _, b := range kb.GetBindings() {
				if b != nil {
					ab.Bindings = append(ab.Bindings, b)
				}
			}
			byID[ab.GetName()] = ab
			out.Actions = append(out.Actions, ab)
			continue
		}
		// union bindings
		for _, nb := range kb.GetBindings() {
			if nb == nil {
				continue
			}
			dup := false
			for _, eb := range existing.GetBindings() {
				if proto.Equal(eb.GetKeyChords(), nb.GetKeyChords()) {
					dup = true
					break
				}
			}
			if !dup {
				existing.Bindings = append(existing.Bindings, nb)
			}
		}
	}
	return out
}
