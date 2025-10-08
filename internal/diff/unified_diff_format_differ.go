package diff

import (
	"fmt"
	"io"

	"github.com/mrk21/go-diff-fmt/difffmt"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const defaultContextLines = 3

// UnifiedDiffFormatDiffer generates a diff in a format similar to `git diff`.
type UnifiedDiffFormatDiffer struct{}

// NewUnifiedDiffFormatDiffer creates a new GitDiffer.
func NewUnifiedDiffFormatDiffer() *UnifiedDiffFormatDiffer {
	return &UnifiedDiffFormatDiffer{}
}

// Diff compares two texts and returns a git-style diff string.
func (d *UnifiedDiffFormatDiffer) Diff(before, after io.Reader, filePath string) (string, error) {
	beforeBytes, err := io.ReadAll(before)
	if err != nil {
		return "", err
	}

	afterBytes, err := io.ReadAll(after)
	if err != nil {
		return "", err
	}

	// Use diff-match-patch line-mode diff to get stable unified output
	dmp := diffmatchpatch.New()
	textA := string(beforeBytes)
	textB := string(afterBytes)
	runesA, runesB, lineArray := dmp.DiffLinesToRunes(textA, textB)
	diffs := dmp.DiffMainRunes(runesA, runesB, false)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// Convert to hunks and format as Unified Diff using go-diff-fmt
	lineDiffs := difffmt.MakeLineDiffsFromDMP(diffs)
	hunks := difffmt.MakeHunks(lineDiffs, defaultContextLines)
	unifiedFmt := difffmt.NewUnifiedFormat(difffmt.UnifiedFormatOption{
		ColorMode: difffmt.ColorNone,
	})
	// Provide git-style paths for headers
	targetA := difffmt.NewDiffTarget(fmt.Sprintf("a/%s", filePath))
	targetB := difffmt.NewDiffTarget(fmt.Sprintf("b/%s", filePath))

	// Standard unified diff format do not include `diff --git ...` header, but client requires it , so add manually
	return fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath) + unifiedFmt.Sprint(targetA, targetB, hunks), nil
}
