package importer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/xinnjie/onekeymap-cli/internal/dedup"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi" // Only for ValidationReport
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap/keybinding"
	"github.com/xinnjie/onekeymap-cli/pkg/api/platform"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/validateapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
	"github.com/xinnjie/onekeymap-cli/pkg/validate"
)

// importer is the default implementation of the Importer interface.
type importer struct {
	registry        *registry.Registry
	mappingConfig   *mappings.MappingConfig
	logger          *slog.Logger
	validator       *validateapi.Validator
	recorder        metrics.Recorder
	serviceReporter *metrics.ServiceReporter
}

// NewImporter creates a new default import service.
func NewImporter(
	registry *registry.Registry,
	config *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) importerapi.Importer {
	service := &importer{
		registry:      registry,
		mappingConfig: config,
		logger:        logger,
		validator: validateapi.NewValidator(
			validate.NewKeybindConflictRule(),
			validate.NewDanglingActionRule(config),
		),
		recorder:        recorder,
		serviceReporter: metrics.NewServiceReporter(recorder),
	}

	return service
}

// Import is the method implementation for the default service.
func (s *importer) Import(ctx context.Context, opts importerapi.ImportOptions) (*importerapi.ImportResult, error) {
	s.serviceReporter.ReportImportCall(ctx)

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
	res, err := importer.Import(ctx, opts.InputStream, pluginapi.PluginImportOption{})
	if err != nil {
		return nil, fmt.Errorf("failed to import config: %w", err)
	}
	setting := res.Keymap

	// Normalize: merge same-action entries and deduplicate identical bindings before downstream logic
	setting.Actions = dedup.Actions(setting.Actions)
	// Sort by action for determinism
	sort.Slice(setting.Actions, func(i, j int) bool {
		return setting.Actions[i].Name < setting.Actions[j].Name
	})

	s.logger.DebugContext(ctx, "imported from plugin", "actions", len(setting.Actions))

	// Check if Actions slice is nil (uninitialized), which indicates import failure
	// An empty slice (len=0) is valid for clearing all bindings
	if setting.Actions == nil {
		return nil, errors.New("failed to import config: no keybindings found")
	}

	report, err := s.validator.Validate(ctx, setting, opts.EditorType)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// No baseline provided: all imported keymaps are additions.
	if len(opts.Base.Actions) == 0 {
		changes := &importerapi.KeymapChanges{}
		if len(setting.Actions) > 0 {
			changes.Add = append(changes.Add, setting.Actions...)
		}
		return &importerapi.ImportResult{
			Setting:    setting,
			Changes:    changes,
			Report:     report,
			SkipReport: res.Report.SkipReport,
		}, nil
	}

	// If baseline provided, first union baseline chords into current setting so unchanged chords are retained.
	setting = unionWithBase(opts.Base, setting)
	setting.Actions = dedup.Actions(setting.Actions)

	// With baseline: compute changes via helper.
	changes := s.calculateChanges(opts.Base, setting)

	// Safety: ensure dedup on output as well
	setting.Actions = dedup.Actions(setting.Actions)
	return &importerapi.ImportResult{
		Setting:    setting,
		Changes:    changes,
		Report:     report,
		SkipReport: res.Report.SkipReport,
	}, nil
}

func (s *importer) calculateChanges(
	base keymap.Keymap,
	setting keymap.Keymap,
) *importerapi.KeymapChanges {
	baseIndex := s.buildActionIndex(base)
	newIndex := s.buildActionIndex(setting)

	adds, removes := s.calculateAddsAndRemoves(baseIndex.byPair, newIndex.byPair)
	updates := s.calculateUpdates(baseIndex.byAction, newIndex.byAction, adds, removes)

	changes := s.buildChangesResult(adds, removes, updates)

	return changes
}

type actionIndex struct {
	byAction map[string][]keymap.Action
	byPair   map[string]keymap.Action
}

func (s *importer) buildActionIndex(km keymap.Keymap) *actionIndex {
	idx := &actionIndex{
		byAction: map[string][]keymap.Action{},
		byPair:   map[string]keymap.Action{},
	}

	for _, kb := range km.Actions {
		if !hasValidChord(kb) {
			continue
		}
		idx.byAction[kb.Name] = append(idx.byAction[kb.Name], kb)
		idx.byPair[pairKey(kb)] = kb
	}

	return idx
}

func (s *importer) calculateAddsAndRemoves(
	basePair map[string]keymap.Action,
	newPair map[string]keymap.Action,
) (map[string]keymap.Action, map[string]keymap.Action) {
	adds := map[string]keymap.Action{}
	removes := map[string]keymap.Action{}

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

func (s *importer) calculateUpdates(
	baseByAction map[string][]keymap.Action,
	newByAction map[string][]keymap.Action,
	adds map[string]keymap.Action,
	removes map[string]keymap.Action,
) []importerapi.KeymapDiff {
	var updates []importerapi.KeymapDiff

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

		updates = append(updates, importerapi.KeymapDiff{Before: before, After: after})
		delete(adds, pairKey(after))
		delete(removes, pairKey(before))
	}

	return updates
}

func (s *importer) buildChangesResult(
	adds map[string]keymap.Action,
	removes map[string]keymap.Action,
	updates []importerapi.KeymapDiff,
) *importerapi.KeymapChanges {
	changes := &importerapi.KeymapChanges{Update: updates}

	if len(adds) > 0 {
		changes.Add = make([]keymap.Action, 0, len(adds))
		for _, v := range adds {
			changes.Add = append(changes.Add, v)
		}
		sort.Slice(changes.Add, func(i, j int) bool {
			return changes.Add[i].Name < changes.Add[j].Name
		})
	}

	if len(removes) > 0 {
		changes.Remove = make([]keymap.Action, 0, len(removes))
		for _, v := range removes {
			changes.Remove = append(changes.Remove, v)
		}
		sort.Slice(changes.Remove, func(i, j int) bool {
			return changes.Remove[i].Name < changes.Remove[j].Name
		})
	}

	return changes
}

// pairKey builds a stable identifier for an action by name and normalized key bindings
func pairKey(action keymap.Action) string {
	if len(action.Bindings) == 0 {
		return action.Name + "\x00"
	}
	// Format each binding
	parts := make([]string, 0, len(action.Bindings))
	for _, b := range action.Bindings {
		if len(b.KeyChords) == 0 {
			continue
		}
		parts = append(parts, b.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"}))
	}
	// Simple insertion sort
	for i := 1; i < len(parts); i++ {
		j := i
		for j > 0 && parts[j] < parts[j-1] {
			parts[j], parts[j-1] = parts[j-1], parts[j]
			j--
		}
	}
	// Join with NUL to avoid ambiguity
	sig := action.Name + "\x00"
	for _, p := range parts {
		sig += p + "\x00"
	}
	return sig
}

func hasValidChord(action keymap.Action) bool {
	for _, b := range action.Bindings {
		if len(b.KeyChords) > 0 {
			return true
		}
	}
	return false
}

// unionWithBase merges baseline and imported settings per action id,
// preserving baseline bindings and adding new imported bindings.
func unionWithBase(base keymap.Keymap, imported keymap.Keymap) keymap.Keymap {
	if len(imported.Actions) == 0 {
		return base
	}
	if len(base.Actions) == 0 {
		return imported
	}
	// index existing results by action id
	out := keymap.Keymap{Actions: []keymap.Action{}}
	byID := make(map[string]*keymap.Action)

	// start with baseline (so Before reflects baseline order/first occurrence)
	for _, kb := range base.Actions {
		// copy action
		ab := keymap.Action{
			Name:     kb.Name,
			Bindings: append([]keybinding.Keybinding{}, kb.Bindings...),
		}
		out.Actions = append(out.Actions, ab)
		byID[ab.Name] = &out.Actions[len(out.Actions)-1]
	}
	// merge imported bindings into corresponding actions (or create new action entries)
	for _, kb := range imported.Actions {
		existing, ok := byID[kb.Name]
		if !ok {
			// add as new action
			ab := keymap.Action{
				Name:     kb.Name,
				Bindings: append([]keybinding.Keybinding{}, kb.Bindings...),
			}
			out.Actions = append(out.Actions, ab)
			byID[ab.Name] = &out.Actions[len(out.Actions)-1]
			continue
		}
		// union bindings
		for _, nb := range kb.Bindings {
			if len(nb.KeyChords) == 0 {
				continue
			}
			dup := false
			nbStr := nb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
			for _, eb := range existing.Bindings {
				ebStr := eb.String(keybinding.FormatOption{Platform: platform.PlatformMacOS, Separator: "+"})
				if ebStr == nbStr {
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
