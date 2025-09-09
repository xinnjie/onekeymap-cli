package diff

import (
	"io"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// UnifiedDiffFormatDiffer generates a diff in a format similar to `git diff`.
type UnifiedDiffFormatDiffer struct{}

// NewUnifiedDiffFormatDiffer creates a new GitDiffer.
func NewUnifiedDiffFormatDiffer() *UnifiedDiffFormatDiffer {
	return &UnifiedDiffFormatDiffer{}
}

// Diff compares two texts and returns a git-style diff string.
func (d *UnifiedDiffFormatDiffer) Diff(before, after io.Reader) (string, error) {
	beforeBytes, err := io.ReadAll(before)
	if err != nil {
		return "", err
	}

	afterBytes, err := io.ReadAll(after)
	if err != nil {
		return "", err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(beforeBytes), string(afterBytes), false)
	patches := dmp.PatchMake(string(beforeBytes), diffs)
	patchText := dmp.PatchToText(patches)
	return patchText, nil
}
