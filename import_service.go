package onekeymap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/platform"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/validateapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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

// unionWithBase merges baseline and imported settings per action id,
// preserving baseline bindings and adding new imported bindings. Metadata from Decorate is preserved.
func unionWithBase(base *keymapv1.Keymap, imported *keymapv1.Keymap) *keymapv1.Keymap {
	if imported == nil {
		return base
	}
	if base == nil || len(base.GetKeybindings()) == 0 {
		return imported
	}
	// index existing results by action id
	out := &keymapv1.Keymap{Keybindings: []*keymapv1.Action{}}
	byID := make(map[string]*keymapv1.Action)

	// start with baseline (so Before reflects baseline order/first occurrence)
	for _, kb := range base.GetKeybindings() {
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
		out.Keybindings = append(out.Keybindings, ab)
	}
	// merge imported bindings into corresponding actions (or create new action entries)
	for _, kb := range imported.GetKeybindings() {
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
			out.Keybindings = append(out.Keybindings, ab)
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
	if setting != nil && len(setting.GetKeybindings()) > 0 {
		setting.Keybindings = dedupKeyBindings(setting.GetKeybindings())
	}

	s.recorder.RecordCommandProcessed(ctx, string(opts.EditorType), setting)

	// Sort by action for determinism
	if setting != nil && len(setting.GetKeybindings()) > 0 {
		sort.Slice(setting.GetKeybindings(), func(i, j int) bool {
			return setting.GetKeybindings()[i].GetName() < setting.GetKeybindings()[j].GetName()
		})
	}

	if setting == nil {
		return nil, errors.New("failed to import config: no keybindings found")
	}

	// Validate
	report, err := s.validate(ctx, setting, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// No baseline provided: all imported keymaps are additions.
	if opts.Base == nil || len(opts.Base.GetKeybindings()) == 0 {
		changes := &importapi.KeymapChanges{}
		if len(setting.GetKeybindings()) > 0 {
			changes.Add = append(changes.Add, setting.GetKeybindings()...)
		}
		return &importapi.ImportResult{Setting: setting, Changes: changes, Report: report}, nil
	}

	// If baseline provided, first union baseline chords into current setting so unchanged chords are retained.
	if opts.Base != nil {
		setting = unionWithBase(opts.Base, setting)
		setting.Keybindings = dedupKeyBindings(setting.GetKeybindings())
		// Re-decorate after union so metadata (Name/Description/Category) and readable chords are present
		setting = keymap.DecorateSetting(setting, s.mappingConfig)
	}

	// With baseline: compute changes via helper.
	changes := s.calculateChanges(opts.Base, setting)

	// Safety: ensure dedup on output as well
	setting.Keybindings = dedupKeyBindings(setting.GetKeybindings())
	return &importapi.ImportResult{Setting: setting, Changes: changes, Report: report}, nil
}

func (s *importService) validate(
	ctx context.Context,
	setting *keymapv1.Keymap,
	opts importapi.ImportOptions,
) (*keymapv1.ValidationReport, error) {
	if setting == nil {
		return &keymapv1.ValidationReport{
			SourceEditor: string(opts.EditorType),
			Summary: &keymapv1.Summary{
				MappingsProcessed: 0,
				MappingsSucceeded: 0,
			},
			Issues:   make([]*keymapv1.ValidationIssue, 0),
			Warnings: make([]*keymapv1.ValidationIssue, 0),
		}, nil
	}

	return s.validator.Validate(ctx, setting, opts)
}

func (s *importService) calculateChanges(
	base *keymapv1.Keymap,
	setting *keymapv1.Keymap,
) *importapi.KeymapChanges {
	if base == nil || setting == nil {
		return &importapi.KeymapChanges{}
	}

	type kmList []*keymapv1.Action
	baseByAction := map[string]kmList{}
	newByAction := map[string]kmList{}
	basePair := map[string]*keymapv1.Action{}
	newPair := map[string]*keymapv1.Action{}

	// hasValidChord returns true if the action has at least one binding with non-empty chords
	hasValidChord := func(kb *keymapv1.Action) bool {
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

	for _, kb := range base.GetKeybindings() {
		if kb == nil || !hasValidChord(kb) {
			// Treat base actions with no valid chords as absent for change calculation
			continue
		}
		baseByAction[kb.GetName()] = append(baseByAction[kb.GetName()], kb)
		basePair[pairKey(kb)] = kb
	}
	for _, kb := range setting.GetKeybindings() {
		if kb == nil || !hasValidChord(kb) {
			// Likewise, ignore empty actions in new setting when computing diffs
			continue
		}
		newByAction[kb.GetName()] = append(newByAction[kb.GetName()], kb)
		newPair[pairKey(kb)] = kb
	}

	// Initial adds/removes by exact (action,keybinding) pair.
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

	// Updates: when exactly one per action in both baseline and new but pair differs.
	var updates []importapi.KeymapDiff
	for action, beforeList := range baseByAction {
		afterList, ok := newByAction[action]
		if !ok {
			continue
		}
		if len(beforeList) == 1 && len(afterList) == 1 {
			before := beforeList[0]
			after := afterList[0]
			if pairKey(before) != pairKey(after) {
				updates = append(updates, importapi.KeymapDiff{Before: before, After: after})
				// Remove from adds/removes to avoid double counting
				delete(adds, pairKey(after))
				delete(removes, pairKey(before))
			}
		}
	}

	changes := &importapi.KeymapChanges{Update: updates}
	if len(adds) > 0 {
		changes.Add = make([]*keymapv1.Action, 0, len(adds))
		for _, v := range adds {
			changes.Add = append(changes.Add, v)
		}
		sort.Slice(changes.Add, func(i, j int) bool { return changes.Add[i].GetName() < changes.Add[j].GetName() })
	}
	if len(removes) > 0 {
		changes.Remove = make([]*keymapv1.Action, 0, len(removes))
		for _, v := range removes {
			changes.Remove = append(changes.Remove, v)
		}
		sort.Slice(
			changes.Remove,
			func(i, j int) bool { return changes.Remove[i].GetName() < changes.Remove[j].GetName() },
		)
	}

	// Decorate metadata (Name/Description/Category) for all keybindings in changes
	decorate := func(kb *keymapv1.Action) {
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
	for _, kb := range changes.Add {
		decorate(kb)
	}
	for _, kb := range changes.Remove {
		decorate(kb)
	}
	for i := range changes.Update {
		decorate(changes.Update[i].Before)
		decorate(changes.Update[i].After)
	}

	return changes
}
