package updatecheck

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChecker_ShouldCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a temporary cache directory
	tmpDir := t.TempDir()

	checker := New("1.0.0", logger)
	checker.cacheDir = tmpDir

	// First check should return true (no cache file)
	if !checker.shouldCheck() {
		t.Error("expected shouldCheck to return true when cache file doesn't exist")
	}

	// Create cache file
	if err := checker.updateCheckTime(); err != nil {
		t.Fatalf("failed to update check time: %v", err)
	}

	// Immediately after, should return false
	if checker.shouldCheck() {
		t.Error("expected shouldCheck to return false immediately after update")
	}

	// Modify the cache file to be old
	cacheFile := filepath.Join(tmpDir, cacheFileName)
	oldTime := time.Now().Add(-checkInterval - 1*time.Hour)
	if err := os.Chtimes(cacheFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to modify cache file time: %v", err)
	}

	// Now should return true
	if !checker.shouldCheck() {
		t.Error("expected shouldCheck to return true when cache is old")
	}
}

func TestChecker_DevVersionSkipsCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	checker := New("dev", logger)

	if checker.currentVersion != "dev" {
		t.Fatalf("expected current version to be dev, got %s", checker.currentVersion)
	}

	// This should not panic or block
	ctx := context.Background()
	checker.CheckForUpdate(ctx)

	// Give goroutine a moment to potentially start
	time.Sleep(10 * time.Millisecond)
}

func TestChecker_FetchLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	checker := New("1.0.0", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release, err := checker.fetchLatestVersion(ctx)
	if err != nil {
		t.Logf("fetch failed (this is expected if network is unavailable): %v", err)
		return
	}

	if release.Version == "" {
		t.Error("expected non-empty version")
	}

	if release.HTMLURL == "" {
		t.Error("expected non-empty URL")
	}

	t.Logf("Latest version: %s, URL: %s", release.Version, release.HTMLURL)
}
