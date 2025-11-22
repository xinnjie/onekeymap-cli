package vscode

import (
	"encoding/json"
	"reflect"
	"sort"

	mappings2 "github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

// FindByVSCodeActionWithArgs searches for a mapping by VSCode command, when clause, and args.
// strict exact matching for both `command` and `when` clauses.
// If strict matching fails, fall back progressively to wildcard `when`, mismatched `when`, and command-only matches.
func (i *vscodeImporter) FindByVSCodeActionWithArgs(
	command, when string,
	args map[string]interface{},
) *mappings2.ActionMappingConfig {
	buckets := newCandidateBuckets()
	for _, mapping := range i.mappingConfig.Mappings {
		i.appendMappingCandidates(buckets, mapping, command, when, args)
	}

	if m := pickCandidate(buckets.exactWhen); m != nil {
		return m
	}
	if m := pickCandidate(buckets.wildcard); m != nil {
		return m
	}
	if m := pickCandidate(buckets.ignoreWhen); m != nil {
		i.logger.Debug(
			"Falling back to command+args match (ignoring when)",
			"command", command,
			"when", when,
			"args", args,
		)
		return m
	}
	if args == nil {
		if m := pickCandidate(buckets.commandOnly); m != nil {
			i.logger.Debug("Falling back to command only match", "command", command, "when", when)
			return m
		}
	}
	return nil
}

type candidate struct {
	mapping          *mappings2.ActionMappingConfig
	enabledForImport bool
}

func newCandidate(mapping mappings2.ActionMappingConfig, enabled bool) *candidate {
	mCopy := mapping
	return &candidate{mapping: &mCopy, enabledForImport: enabled}
}

type bucketKind int

const (
	bucketNone bucketKind = iota
	bucketExactWhen
	bucketWildcard
	bucketIgnoreWhen
	bucketCommandOnly
)

type candidateBuckets struct {
	exactWhen   []*candidate
	wildcard    []*candidate
	ignoreWhen  []*candidate
	commandOnly []*candidate

	seenExact       map[string]struct{}
	seenWildcard    map[string]struct{}
	seenIgnoreWhen  map[string]struct{}
	seenCommandOnly map[string]struct{}
}

func newCandidateBuckets() *candidateBuckets {
	return &candidateBuckets{
		seenExact:       make(map[string]struct{}),
		seenWildcard:    make(map[string]struct{}),
		seenIgnoreWhen:  make(map[string]struct{}),
		seenCommandOnly: make(map[string]struct{}),
	}
}

func (b *candidateBuckets) add(kind bucketKind, mappingID string, cand *candidate) {
	switch kind {
	case bucketExactWhen:
		if _, ok := b.seenExact[mappingID]; ok {
			return
		}
		b.seenExact[mappingID] = struct{}{}
		b.exactWhen = append(b.exactWhen, cand)
	case bucketWildcard:
		if _, ok := b.seenWildcard[mappingID]; ok {
			return
		}
		b.seenWildcard[mappingID] = struct{}{}
		b.wildcard = append(b.wildcard, cand)
	case bucketIgnoreWhen:
		if _, ok := b.seenIgnoreWhen[mappingID]; ok {
			return
		}
		b.seenIgnoreWhen[mappingID] = struct{}{}
		b.ignoreWhen = append(b.ignoreWhen, cand)
	case bucketCommandOnly:
		if _, ok := b.seenCommandOnly[mappingID]; ok {
			return
		}
		b.seenCommandOnly[mappingID] = struct{}{}
		b.commandOnly = append(b.commandOnly, cand)
	}
}

func pickCandidate(cands []*candidate) *mappings2.ActionMappingConfig {
	if len(cands) == 0 {
		return nil
	}
	var (
		enabled  []*candidate
		disabled []*candidate
	)
	for _, c := range cands {
		if c.enabledForImport {
			enabled = append(enabled, c)
		} else {
			disabled = append(disabled, c)
		}
	}
	// Prefer enabled configs, fall back to disabled if no enabled ones
	chooseFrom := enabled
	if len(chooseFrom) == 0 {
		chooseFrom = disabled
	}
	sort.Slice(chooseFrom, func(i, j int) bool {
		return chooseFrom[i].mapping.ID < chooseFrom[j].mapping.ID
	})
	return chooseFrom[0].mapping
}

func (i *vscodeImporter) appendMappingCandidates(
	buckets *candidateBuckets,
	mapping mappings2.ActionMappingConfig,
	command, when string,
	args map[string]interface{},
) {
	if len(mapping.VSCode) == 0 {
		return
	}

	for _, vc := range mapping.VSCode {
		if vc.Command != command {
			continue
		}
		// Skip configs that are explicitly disabled for import
		if vc.DisableImport {
			continue
		}

		bucket := determineBucket(vc, when, args)
		if bucket == bucketNone {
			continue
		}

		// All remaining configs are enabled for import
		cand := newCandidate(mapping, true)
		buckets.add(bucket, mapping.ID, cand)
	}
}

func determineBucket(vc mappings2.VscodeMappingConfig, when string, args map[string]interface{}) bucketKind {
	if equalArgs(vc.Args, args) {
		if vc.When == "" {
			return bucketWildcard
		}
		if vc.When == when {
			return bucketExactWhen
		}
		return bucketIgnoreWhen
	}
	if args == nil {
		return bucketCommandOnly
	}
	return bucketNone
}

// equalArgs compares two args maps by canonical JSON encoding to avoid type
// mismatches (e.g., YAML int vs JSON float64) and map iteration order issues.
func equalArgs(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	ab, err1 := json.Marshal(a)
	bb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		// As a conservative fallback, use reflect.DeepEqual
		return reflect.DeepEqual(a, b)
	}
	return string(ab) == string(bb)
}
