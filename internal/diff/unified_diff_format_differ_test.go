package diff

import (
	"bytes"
	"testing"
)

func TestGitDiffer_NoChange_ReturnsEqualDelta(t *testing.T) {
	d := NewUnifiedDiffFormatDiffer()
	text := "hello world"
	got, err := d.Diff(bytes.NewBufferString(text), bytes.NewBufferString(text))
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := ""
	if got != want {
		t.Fatalf("unexpected delta for identical inputs\nwant: %q\n got: %q", want, got)
	}
}

func TestGitDiffer_OnlyInsert(t *testing.T) {
	d := NewUnifiedDiffFormatDiffer()
	before := ""
	after := "abc"
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after))
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := "@@ -0,0 +1,3 @@\n+abc\n"
	if got != want {
		t.Fatalf("unexpected delta for insert\nwant: %q\n got: %q", want, got)
	}
}

func TestGitDiffer_OnlyDelete(t *testing.T) {
	d := NewUnifiedDiffFormatDiffer()
	before := "abc"
	after := ""
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after))
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := "@@ -1,3 +0,0 @@\n-abc\n"
	if got != want {
		t.Fatalf("unexpected delta for delete\nwant: %q\n got: %q", want, got)
	}
}

func TestGitDiffer_ReplaceMiddleChar(t *testing.T) {
	d := NewUnifiedDiffFormatDiffer()
	before := "abc"
	after := "axc"
	got, err := d.Diff(bytes.NewBufferString(before), bytes.NewBufferString(after))
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	want := "@@ -1,3 +1,3 @@\n a\n-b\n+x\n c\n"
	if got != want {
		t.Fatalf("unexpected delta for replace\nwant: %q\n got: %q", want, got)
	}
}
