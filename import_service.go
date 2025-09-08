package onekeymap

import (
	"context"
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
func unionWithBase(base *keymapv1.KeymapSetting, imported *keymapv1.KeymapSetting) *keymapv1.KeymapSetting {
	if imported == nil {
		return base
	}
	if base == nil || len(base.GetKeybindings()) == 0 {
		return imported
	}
	// index existing results by action id
	out := &keymapv1.KeymapSetting{Keybindings: []*keymapv1.ActionBinding{}}
	byID := make(map[string]*keymapv1.ActionBinding)

	// start with baseline (so Before reflects baseline order/first occurrence)
	for _, kb := range base.GetKeybindings() {
		if kb == nil {
			continue
		}
		// shallow copy action binding shell; reuse bindings slice ref (safe; will be appended below possibly)
		ab := &keymapv1.ActionBinding{Id: kb.GetId(), Name: kb.GetName(), Description: kb.GetDescription(), Category: kb.GetCategory()}
		// copy bindings
		for _, b := range kb.GetBindings() {
			if b != nil {
				ab.Bindings = append(ab.Bindings, b)
			}
		}
		byID[ab.Id] = ab
		out.Keybindings = append(out.Keybindings, ab)
	}
	// merge imported bindings into corresponding actions (or create new action entries)
	for _, kb := range imported.GetKeybindings() {
		if kb == nil {
			continue
		}
		existing, ok := byID[kb.GetId()]
		if !ok {
			// add as new action
			ab := &keymapv1.ActionBinding{Id: kb.GetId(), Name: kb.GetName(), Description: kb.GetDescription(), Category: kb.GetCategory()}
			for _, b := range kb.GetBindings() {
				if b != nil {
					ab.Bindings = append(ab.Bindings, b)
				}
			}
			byID[ab.Id] = ab
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
func NewImportService(registry *plugins.Registry, config *mappings.MappingConfig, logger *slog.Logger, recorder metrics.Recorder) importapi.Importer {
	service := &importService{
		registry:      registry,
		mappingConfig: config,
		logger:        logger,
		validator:     validateapi.NewValidator(validateapi.NewKeybindConflictRule(), validateapi.NewDanglingActionRule(config)),
		recorder:      recorder,
	}

	return service
}

// Import is the method implementation for the default service.
func (s *importService) Import(ctx context.Context, opts importapi.ImportOptions) (*importapi.ImportResult, error) {
	if opts.InputStream == nil {
		return nil, fmt.Errorf("input stream is required")
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
	if setting != nil && len(setting.Keybindings) > 0 {
		setting.Keybindings = dedupKeyBindings(setting.Keybindings)
	}

	s.recorder.RecordCommandProcessed(ctx, string(opts.EditorType), setting)

	// Sort by action for determinism
	if setting != nil && len(setting.Keybindings) > 0 {
		sort.Slice(setting.Keybindings, func(i, j int) bool {
			return setting.Keybindings[i].Id < setting.Keybindings[j].Id
		})
	}

	if setting == nil {
		return nil, nil
	}

	// Validate
	report, err := s.validate(ctx, setting, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// No baseline provided: all imported keymaps are additions.
	if opts.Base == nil || len(opts.Base.GetKeybindings()) == 0 {
		changes := &importapi.KeymapChanges{}
		if len(setting.Keybindings) > 0 {
			changes.Add = append(changes.Add, setting.Keybindings...)
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
	changes, err := s.calculateChanges(opts.Base, setting)
	if err != nil {
		return nil, err
	}

	// Safety: ensure dedup on output as well
	setting.Keybindings = dedupKeyBindings(setting.Keybindings)
	return &importapi.ImportResult{Setting: setting, Changes: changes, Report: report}, nil
}

func (s *importService) validate(ctx context.Context, setting *keymapv1.KeymapSetting, opts importapi.ImportOptions) (*keymapv1.ValidationReport, error) {
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

func (s *importService) calculateChanges(base *keymapv1.KeymapSetting, setting *keymapv1.KeymapSetting) (*importapi.KeymapChanges, error) {
	if base == nil || setting == nil {
		return &importapi.KeymapChanges{}, nil
	}

	type kmList []*keymapv1.ActionBinding
	baseByAction := map[string]kmList{}
	newByAction := map[string]kmList{}
	basePair := map[string]*keymapv1.ActionBinding{}
	newPair := map[string]*keymapv1.ActionBinding{}

	for _, kb := range base.GetKeybindings() {
		if kb == nil {
			continue
		}
		baseByAction[kb.GetId()] = append(baseByAction[kb.GetId()], kb)
		basePair[pairKey(kb)] = kb
	}
	for _, kb := range setting.GetKeybindings() {
		if kb == nil {
			continue
		}
		newByAction[kb.GetId()] = append(newByAction[kb.GetId()], kb)
		newPair[pairKey(kb)] = kb
	}

	// Initial adds/removes by exact (action,keybinding) pair.
	adds := map[string]*keymapv1.ActionBinding{}
	removes := map[string]*keymapv1.ActionBinding{}
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
		changes.Add = make([]*keymapv1.ActionBinding, 0, len(adds))
		for _, v := range adds {
			changes.Add = append(changes.Add, v)
		}
		sort.Slice(changes.Add, func(i, j int) bool { return changes.Add[i].Id < changes.Add[j].Id })
	}
	if len(removes) > 0 {
		changes.Remove = make([]*keymapv1.ActionBinding, 0, len(removes))
		for _, v := range removes {
			changes.Remove = append(changes.Remove, v)
		}
		sort.Slice(changes.Remove, func(i, j int) bool { return changes.Remove[i].Id < changes.Remove[j].Id })
	}

	// Decorate metadata (Name/Description/Category) for all keybindings in changes
	decorate := func(kb *keymapv1.ActionBinding) {
		if kb == nil {
			return
		}
		if cfg := s.mappingConfig.FindByUniversalAction(kb.GetId()); cfg != nil {
			kb.Description = cfg.Description
			kb.Name = cfg.Name
			kb.Category = cfg.Category
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

	return changes, nil
}
