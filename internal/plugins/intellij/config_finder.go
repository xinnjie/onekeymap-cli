package intellij

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns default keymap config path for IntelliJ.
// NOTE: IntelliJ family has multiple editions/versions; precise discovery will
// be implemented later. For now, we return a not-supported error placeholder.
func (p *intellijPlugin) ConfigDetect(opts ...pluginapi.ConfigDetectOption) ([]string, error) {
	options := &pluginapi.ConfigDetectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if runtime.GOOS != "darwin" {
		return nil, errors.New("automatic path discovery is only supported on macOS for IntelliJ")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "JetBrains", "*", "keymaps"),
	}

	var keymapDirs []string
	for _, pat := range candidates {
		matches, _ := filepath.Glob(pat)
		for _, m := range matches {
			if fi, err := os.Stat(m); err == nil && fi.IsDir() {
				keymapDirs = append(keymapDirs, m)
			}
		}
	}
	if len(keymapDirs) == 0 {
		return nil, errors.New("could not locate JetBrains keymaps directory; please specify --output")
	}

	sort.Slice(keymapDirs, func(i, j int) bool {
		fi, _ := os.Stat(keymapDirs[i])
		fj, _ := os.Stat(keymapDirs[j])
		var ti, tj time.Time
		if fi != nil {
			ti = fi.ModTime()
		}
		if fj != nil {
			tj = fj.ModTime()
		}
		return ti.After(tj)
	})

	configPath := filepath.Join(keymapDirs[0], "Onekeymap.xml")

	return []string{configPath}, nil
}
