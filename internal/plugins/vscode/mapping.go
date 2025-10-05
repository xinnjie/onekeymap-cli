package vscode

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/xinnjie/onekeymap-cli/internal/mappings"
)

// FindByVSCodeActionWithArgs searches for a mapping by VSCode command, when clause, and args.
// strict exact matching for both `command` and `when` clauses
// if strict matching fails, it will try to find a matching action by `command` only, return the command if only one action matches the `command`, else return nil.
func (i *vscodeImporter) FindByVSCodeActionWithArgs(
	command, when string,
	args map[string]interface{},
) *mappings.ActionMappingConfig {
	// We collect candidates first, then decide deterministically to avoid
	// any map-iteration induced non-determinism.
	type candidate struct {
		m         *mappings.ActionMappingConfig
		forImport bool // effective: explicit ForImport or single VSCode config in that action
		explicit  bool // true only if explicitly set by config
	}

	var exactWhenArgs []*candidate     // command + args match, when exactly equals (non-empty)
	var wildcardWhenArgs []*candidate  // command + args match, when is empty in config (wildcard)
	var cmdArgsIgnoreWhen []*candidate // command + args match, but when differs (ignore when)
	var cmdOnly []*candidate           // command only (used when args == nil)

	seenExact := make(map[string]struct{})
	seenWildcard := make(map[string]struct{})
	seenCmdArgs := make(map[string]struct{})
	seenCmdOnly := make(map[string]struct{})

	for _, mapping := range i.mappingConfig.Mappings {
		// If this action mapping has any explicit ForImport entries, we will
		// only consider those entries for import selection and ignore others.
		hasExplicitForImport := false
		for _, vc := range mapping.VSCode {
			if vc.ForImport {
				hasExplicitForImport = true
				break
			}
		}

		for _, vc := range mapping.VSCode {
			if vc.Command != command {
				continue
			}

			effectiveForImport := vc.ForImport || len(mapping.VSCode) == 1
			explicitForImport := vc.ForImport

			// Gating: if any entry within the same action mapping is marked
			// forImport=true, then non-forImport entries from the same mapping
			// are not eligible for import.
			if hasExplicitForImport && !explicitForImport {
				continue
			}

			argsBothNil := (vc.Args == nil && args == nil)
			argsBothNonNilAndEqual := (vc.Args != nil && args != nil && equalArgs(vc.Args, args))
			argsMatch := argsBothNil || argsBothNonNilAndEqual

			whenExact := (vc.When != "" && vc.When == when)
			whenWildcard := (vc.When == "")

			if argsMatch {
				if whenExact {
					if _, ok := seenExact[mapping.ID]; !ok {
						m := mapping
						exactWhenArgs = append(
							exactWhenArgs,
							&candidate{m: &m, forImport: effectiveForImport, explicit: explicitForImport},
						)
						seenExact[mapping.ID] = struct{}{}
					}
					continue
				}
				if whenWildcard {
					if _, ok := seenWildcard[mapping.ID]; !ok {
						m := mapping
						wildcardWhenArgs = append(
							wildcardWhenArgs,
							&candidate{m: &m, forImport: effectiveForImport, explicit: explicitForImport},
						)
						seenWildcard[mapping.ID] = struct{}{}
					}
					continue
				}
				// args match but when doesn't — keep as command+args ignoring when
				if _, ok := seenCmdArgs[mapping.ID]; !ok {
					m := mapping
					cmdArgsIgnoreWhen = append(
						cmdArgsIgnoreWhen,
						&candidate{m: &m, forImport: effectiveForImport, explicit: explicitForImport},
					)
					seenCmdArgs[mapping.ID] = struct{}{}
				}
				continue
			}

			// If args are not provided by incoming binding, allow command-only fallback.
			if args == nil {
				if _, ok := seenCmdOnly[mapping.ID]; !ok {
					m := mapping
					cmdOnly = append(
						cmdOnly,
						&candidate{m: &m, forImport: effectiveForImport, explicit: explicitForImport},
					)
					seenCmdOnly[mapping.ID] = struct{}{}
				}
			}
		}
	}

	// Helper to pick deterministically with ForImport preference (explicit over implicit), then by ID
	pick := func(cands []*candidate) *mappings.ActionMappingConfig {
		if len(cands) == 0 {
			return nil
		}
		var explicitForImport, implicitForImport, others []*candidate
		for _, c := range cands {
			switch {
			case c.forImport && c.explicit:
				explicitForImport = append(explicitForImport, c)
			case c.forImport:
				implicitForImport = append(implicitForImport, c)
			default:
				others = append(others, c)
			}
		}
		chooseFrom := cands
		switch {
		case len(explicitForImport) > 0:
			chooseFrom = explicitForImport
		case len(implicitForImport) > 0:
			chooseFrom = implicitForImport
		default:
			chooseFrom = others
		}
		sort.Slice(chooseFrom, func(i, j int) bool { return chooseFrom[i].m.ID < chooseFrom[j].m.ID })
		return chooseFrom[0].m
	}

	if m := pick(exactWhenArgs); m != nil {
		return m
	}
	if m := pick(wildcardWhenArgs); m != nil {
		return m
	}
	if m := pick(cmdArgsIgnoreWhen); m != nil {
		i.logger.Debug(
			"Falling back to command+args match (ignoring when)",
			"command",
			command,
			"when",
			when,
			"args",
			args,
		)
		return m
	}
	if args == nil {
		if m := pick(cmdOnly); m != nil {
			i.logger.Debug("Falling back to command only match", "command", command, "when", when)
			return m
		}
	}
	return nil
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
