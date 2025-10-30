package diff_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinnjie/onekeymap-cli/internal/diff"
)

func TestGitDiffer_NoChange_ReturnsEqualDelta(t *testing.T) {
	d := diff.NewUnifiedDiffFormatDiffer()
	text := "hello world"
	got, err := d.Diff(bytes.NewBufferString(text), bytes.NewBufferString(text), "test")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	// go-diff-fmt prints only headers when there is no change, with a git header line
	want := "diff --git a/test b/test\n--- a/test\n+++ b/test\n"
	assert.Equal(t, want, got)
}

func TestGitDiffer_OnlyInsert(t *testing.T) {
	d := diff.NewUnifiedDiffFormatDiffer()
	before := "line1\n"
	after := "line1\nline2\n"
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after), "test")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := "diff --git a/test b/test\n--- a/test\n+++ b/test\n@@ -1 +1,2 @@\n line1\n+line2\n"
	assert.Equal(t, want, got)
}

func TestGitDiffer_OnlyDelete(t *testing.T) {
	d := diff.NewUnifiedDiffFormatDiffer()
	before := "abc"
	after := ""
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after), "test")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	// No trailing newline in input, expect explicit no-newline message
	want := "diff --git a/test b/test\n--- a/test\n+++ b/test\n@@ -1 +0,0 @@\n-abc\n\\ No newline at end of file\n"
	assert.Equal(t, want, got)
}

func TestGitDiffer_ReplaceMiddleChar(t *testing.T) {
	d := diff.NewUnifiedDiffFormatDiffer()
	before := "abc"
	after := "axc"
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after), "test")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := "diff --git a/test b/test\n--- a/test\n+++ b/test\n@@ -1 +1 @@\n-abc\n\\ No newline at end of file\n+axc\n\\ No newline at end of file\n"
	assert.Equal(t, want, got)
}
