package onekeymap

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/validateapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
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

	// No baseline provided: all imported keymaps are additions.
	if opts.Base == nil || len(opts.Base.GetKeybindings()) == 0 {
		changes := &importapi.KeymapChanges{}
		if len(setting.Keybindings) > 0 {
			changes.Add = append(changes.Add, setting.Keybindings...)
		}
		return &importapi.ImportResult{Setting: setting, Changes: changes}, nil
	}

	// Validate
	report, err := s.validate(ctx, setting, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// With baseline: compute changes via helper.
	changes, err := s.calculateChanges(opts.Base, setting)
	if err != nil {
		return nil, err
	}

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

	type kmList []*keymapv1.KeyBinding
	baseByAction := map[string]kmList{}
	newByAction := map[string]kmList{}
	basePair := map[string]*keymapv1.KeyBinding{}
	newPair := map[string]*keymapv1.KeyBinding{}

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
	adds := map[string]*keymapv1.KeyBinding{}
	removes := map[string]*keymapv1.KeyBinding{}
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
		changes.Add = make([]*keymapv1.KeyBinding, 0, len(adds))
		for _, v := range adds {
			changes.Add = append(changes.Add, v)
		}
		sort.Slice(changes.Add, func(i, j int) bool { return changes.Add[i].Id < changes.Add[j].Id })
	}
	if len(removes) > 0 {
		changes.Remove = make([]*keymapv1.KeyBinding, 0, len(removes))
		for _, v := range removes {
			changes.Remove = append(changes.Remove, v)
		}
		sort.Slice(changes.Remove, func(i, j int) bool { return changes.Remove[i].Id < changes.Remove[j].Id })
	}

	// Decorate metadata (Name/Description/Category) for all keybindings in changes
	decorate := func(kb *keymapv1.KeyBinding) {
		if kb == nil {
			return
		}
		if cfg := s.mappingConfig.FindByUniversalAction(kb.GetId()); cfg != nil {
			kb.Description = cfg.Description
			kb.Name = cfg.Name
			kb.Category = cfg.Category
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
