package updatecheck_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/xinnjie/onekeymap-cli/internal/updatecheck"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("creates checker with valid version", func(t *testing.T) {
		checker := updatecheck.New("1.0.0", logger)
		if checker == nil {
			t.Fatal("expected non-nil checker")
		}
	})

	t.Run("creates checker with dev version", func(t *testing.T) {
		checker := updatecheck.New("dev", logger)
		if checker == nil {
			t.Fatal("expected non-nil checker")
		}
	})
}

func TestChecker_CheckForUpdateMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("dev version returns empty string", func(t *testing.T) {
		checker := updatecheck.New("dev", logger)
		ctx := context.Background()
		msg := checker.CheckForUpdateMessage(ctx)

		if msg != "" {
			t.Errorf("expected empty message for dev version, got %s", msg)
		}
	})

	t.Run("returns result for normal version", func(t *testing.T) {
		checker := updatecheck.New("1.0.0", logger)
		ctx := context.Background()

		// This should return empty string since network check will likely fail in test environment
		msg := checker.CheckForUpdateMessage(ctx)

		// We don't assert the exact content since it depends on network availability
		// Just ensure it doesn't panic and returns a string
		if msg != "" {
			t.Logf("Update message received: %s", msg)
		}
	})
}

func TestChecker_CheckForUpdateAsync(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("dev version returns empty string via channel", func(t *testing.T) {
		checker := updatecheck.New("dev", logger)
		ctx := context.Background()
		msgChan := checker.CheckForUpdateAsync(ctx)

		// Should receive empty string for dev version
		msg := <-msgChan
		if msg != "" {
			t.Errorf("expected empty message for dev version, got %s", msg)
		}

		// Channel should be closed
		_, ok := <-msgChan
		if ok {
			t.Error("expected channel to be closed")
		}
	})

	t.Run("normal version returns result via channel", func(t *testing.T) {
		checker := updatecheck.New("1.0.0", logger)
		ctx := context.Background()
		msgChan := checker.CheckForUpdateAsync(ctx)

		// Should receive a result (likely empty string due to network/cache)
		msg := <-msgChan

		// We don't assert the exact content since it depends on network availability
		// Just ensure we get a response and channel closes properly
		_ = msg

		// Channel should be closed
		_, ok := <-msgChan
		if ok {
			t.Error("expected channel to be closed")
		}
	})
}
