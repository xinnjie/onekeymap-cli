package diff

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

// Differ defines the contract for comparing two objects and generating a diff.
type Differ interface {
	// Diff compares two objects and returns an ASCII Diff string.
	Diff(before, after any) (asciiDiff string, err error)
}

// NewJSONASCIIDiffer creates a new JSON differ that outputs ANSI-colored ASCII diffs.
func NewJSONASCIIDiffer() Differ {
	return &jsonASCIIDiffer{}
}

// NewJSONDiffer creates a new JSON differ that strips ANSI colors.
// It accepts nil, map[string]any, []any, or any strong-typed value which will be
// normalized via JSON round-trip.
func NewJSONDiffer() Differ {
	return &jsonDiffer{}
}

type jsonASCIIDiffer struct {
}

func (d *jsonASCIIDiffer) Diff(before, after any) (asciiDiff string, err error) {
	differ := gojsondiff.New()
	var diff gojsondiff.Diff

	// Normalize inputs to generic JSON values (map[string]any or []any)
	nb, err := normalizeToJSONValue(before)
	if err != nil {
		return "", err
	}
	na, err := normalizeToJSONValue(after)
	if err != nil {
		return "", err
	}

	objectBefore, objectBeforeOk := nb.(map[string]any)
	objectAfter, objectAfterOk := na.(map[string]any)
	arrayBefore, arrayBeforeOk := nb.([]any)
	arrayAfter, arrayAfterOk := na.([]any)
	switch {
	case objectBeforeOk && objectAfterOk:
		diff = differ.CompareObjects(objectBefore, objectAfter)
	case arrayBeforeOk && arrayAfterOk:
		diff = differ.CompareArrays(arrayBefore, arrayAfter)
	default:
		return "", errors.New("type mismatch: before and after are not both objects or arrays")
	}

	if !diff.Modified() {
		return "", nil
	}
	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       true,
	}

	formatter := formatter.NewAsciiFormatter(nb, config)
	diffString, err := formatter.Format(diff)
	if err != nil {
		return "", err
	}
	return diffString, nil
}

// jsonASCIIDiffer but strip ansi color codes.
type jsonDiffer struct {
}

func (d *jsonDiffer) Diff(before, after any) (asciiDiff string, err error) {
	asciiDiff, err = NewJSONASCIIDiffer().Diff(before, after)
	if err != nil {
		return "", err
	}
	return stripANSI(asciiDiff), nil
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRegexp.ReplaceAllString(s, "") }

// normalizeToJSONValue converts various input types to a generic JSON value (map[string]any or []any).
// Rules:
// - nil => []any{}
// - map[string]any or []any => returned as-is
// - other types => marshal to JSON then unmarshal into 'any'.
func normalizeToJSONValue(v any) (any, error) {
	if v == nil {
		// Default to empty array to align with callers expecting array top-level (e.g., Zed keymaps)
		return []any{}, nil
	}

	switch x := v.(type) {
	case map[string]any, []any:
		return x, nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		if rv.IsNil() {
			return map[string]any{}, nil
		}
	case reflect.Slice:
		if rv.IsNil() {
			return []any{}, nil
		}
	}

	// JSON round trip to convert strong types into generic maps/slices
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value to JSON: %w", err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value from JSON: %w", err)
	}
	if out == nil {
		switch rv.Kind() {
		case reflect.Map:
			return map[string]any{}, nil
		case reflect.Slice:
			return []any{}, nil
		}
		return []any{}, nil
	}
	return out, nil
}
