//go:build darwin

package intellij

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

func TestConfigDetect_Darwin(t *testing.T) {
	mappingConfig, err := mappings.NewTestMappingConfig()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	plugin := New(mappingConfig, logger, metrics.NewNoop())

	t.Run("no keymap directories -> error", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		_, _, err := plugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not locate IntelliJ IDEA keymaps directory")
	})

	t.Run("choose most recently modified keymaps dir", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		oldDir := filepath.Join(home, "Library", "Application Support", "JetBrains", "IntelliJIdea2023.2", "keymaps")
		newDir := filepath.Join(home, "Library", "Application Support", "JetBrains", "IntelliJIdea2024.1", "keymaps")

		for _, d := range []string{oldDir, newDir} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", d, err)
			}
		}

		// Ensure differing mod times: oldDir older, newDir newer
		oldTime := time.Now().Add(-2 * time.Hour)
		newTime := time.Now().Add(-30 * time.Minute)
		if err := os.Chtimes(oldDir, oldTime, oldTime); err != nil {
			t.Fatalf("chtimes oldDir: %v", err)
		}
		if err := os.Chtimes(newDir, newTime, newTime); err != nil {
			t.Fatalf("chtimes newDir: %v", err)
		}

		p, _, err := plugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
		require.NoError(t, err)
		expected := []string{filepath.Join(newDir, "Onekeymap.xml")}
		assert.Equal(t, expected, p)
	})
}
