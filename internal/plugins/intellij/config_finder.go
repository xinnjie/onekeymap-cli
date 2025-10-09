package intellij

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

// ConfigDetect returns default keymap config path for IntelliJ IDEA Ultimate.
func (p *intellijPlugin) ConfigDetect(opts pluginapi.ConfigDetectOptions) (paths []string, installed bool, err error) {
	return detectConfigForIDE("IntelliJ IDEA", "IntelliJIdea*", "idea", opts)
}

// detectConfigForIDE is a helper function to detect config paths for a specific JetBrains IDE.
func detectConfigForIDE(
	appNamePrefix, dirPattern, commandName string,
	opts pluginapi.ConfigDetectOptions,
) (paths []string, installed bool, err error) {
	if runtime.GOOS != "darwin" {
		return nil, false, fmt.Errorf(
			"automatic path discovery is only supported on macOS for IntelliJ, %w",
			pluginapi.ErrNotSupported,
		)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, err
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "JetBrains", dirPattern, "keymaps"),
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
		return nil, false, fmt.Errorf("could not locate %s keymaps directory", appNamePrefix)
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

	if opts.Sandbox {
		installed = false
	} else {
		// Outside of sandbox, `exec.LookPath` is the most reliable way to see if the command is in the user's PATH.
		_, err := exec.LookPath(commandName)
		installed = err == nil
	}

	return []string{configPath}, installed, nil
}
