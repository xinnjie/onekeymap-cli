package cmd

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	ij "github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/zed"
	"github.com/xinnjie/onekeymap-cli/pkg/pluginapi"
)

type devGenerateBaseFlags struct {
	IntelliJSourceDir string
	VSCodeSourceDir   string
	ZedSourceDir      string
	OutDir            string
	Editor            string
}

const (
	editorAll      = "all"
	editorIntelliJ = "intellij"
	editorVSCode   = "vscode"
	editorZed      = "zed"
)

func NewCmdDevGenerateBase() *cobra.Command {
	f := devGenerateBaseFlags{}
	cmd := &cobra.Command{
		Use:   "generateBase",
		Short: "Generate base keymap JSONs from editor-specific keymap files",
		Run:   devGenerateBaseRun(&f, func() (*slog.Logger, metrics.Recorder) { return cmdLogger, cmdRecorder }),
		Args:  cobra.ExactArgs(0),
	}
	cmd.Flags().
		StringVar(&f.IntelliJSourceDir, "intellij-source-dir", "chore/intellij", "Directory containing IntelliJ keymap XML files")
	cmd.Flags().
		StringVar(&f.VSCodeSourceDir, "vscode-source-dir", "chore/vscode", "Directory containing VSCode keymap JSON files")
	cmd.Flags().
		StringVar(&f.ZedSourceDir, "zed-source-dir", "chore/zed", "Directory containing Zed keymap JSON files")
	cmd.Flags().StringVar(&f.OutDir, "out-dir", "config/base", "Directory to write base JSON files")
	cmd.Flags().
		StringVar(&f.Editor, "editor", editorAll, "Editor to generate base keymap for (intellij, vscode, zed, or all)")
	return cmd
}

func devGenerateBaseRun(
	f *devGenerateBaseFlags,
	dependencies func() (*slog.Logger, metrics.Recorder),
) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger, recorder := dependencies()
		ctx := cmd.Context()

		mc, err := mappings.NewMappingConfig()
		if err != nil {
			logger.ErrorContext(ctx, "load mapping config", "error", err)
			os.Exit(1)
		}

		if err := os.MkdirAll(f.OutDir, 0o750); err != nil {
			logger.ErrorContext(ctx, "ensure out-dir", "dir", f.OutDir, "error", err)
			os.Exit(1)
		}

		if f.Editor == editorAll || f.Editor == editorIntelliJ {
			generateIntelliJBase(ctx, f, mc, logger, recorder)
		}

		if f.Editor == editorAll || f.Editor == editorVSCode {
			generateVSCodeBase(ctx, f, mc, logger, recorder)
		}

		if f.Editor == editorAll || f.Editor == editorZed {
			generateZedBase(ctx, f, mc, logger, recorder)
		}
	}
}

func generateIntelliJBase(
	ctx context.Context,
	f *devGenerateBaseFlags,
	mc *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) {
	plugin := ij.New(mc, logger, recorder)
	imp, err := plugin.Importer()
	if err != nil {
		logger.ErrorContext(ctx, "get intellij importer", "error", err)
		os.Exit(1)
	}

	tasks := []struct {
		name     string
		source   string
		outfile  string
		platform platform.Platform
	}{
		{name: "mac", source: "Mac OS X.xml", outfile: "intellij-mac.json", platform: platform.PlatformMacOS},
		{
			name:     "windows",
			source:   "$default.xml",
			outfile:  "intellij-windows.json",
			platform: platform.PlatformWindows,
		},
		{
			name:     "linux",
			source:   "Default for GNOME.xml",
			outfile:  "intellij-linux.json",
			platform: platform.PlatformLinux,
		},
	}

	for _, t := range tasks {
		logger.InfoContext(ctx, "generating intellij base", "platform", t.name, "src", t.source)
		xmlDoc, meta, err := loadAndFlattenIJXML(ctx, f.IntelliJSourceDir, t.source)
		if err != nil {
			logger.ErrorContext(ctx, "flatten xml", "src", t.source, "error", err)
			os.Exit(1)
		}

		// For macOS: convert 'control' to 'meta' for inherited actions only
		// This matches IntelliJ's runtime behavior: actions inherited from $default.xml
		// have control->command mapping, while actions explicitly defined in Mac OS X.xml
		// keep their control keys as-is (e.g., control SPACE for code completion)
		if t.platform == platform.PlatformMacOS && meta != nil {
			convertControlToMetaForMac(xmlDoc, meta)
		}

		buf := &bytes.Buffer{}
		if err := xml.NewEncoder(buf).Encode(xmlDoc); err != nil {
			logger.ErrorContext(ctx, "encode xml", "error", err)
			os.Exit(1)
		}

		km, err := imp.Import(ctx, bytes.NewReader(buf.Bytes()), pluginapi.PluginImportOption{})
		if err != nil {
			logger.ErrorContext(ctx, "import intellij xml", "src", t.source, "error", err)
			os.Exit(1)
		}

		outPath := filepath.Join(f.OutDir, t.outfile)
		fp, err := os.Create(outPath)
		if err != nil {
			logger.ErrorContext(ctx, "create output", "path", outPath, "error", err)
			os.Exit(1)
		}
		if err := keymap.Save(fp, km, keymap.SaveOptions{Platform: t.platform}); err != nil {
			_ = fp.Close()
			logger.ErrorContext(ctx, "write base json", "path", outPath, "error", err)
			os.Exit(1)
		}
		if err := fp.Close(); err != nil {
			logger.WarnContext(ctx, "close file", "path", outPath, "error", err)
		}
		logger.InfoContext(ctx, "wrote intellij base", "path", outPath)
	}
}

type genTask struct {
	name     string
	source   string
	outfile  string
	platform platform.Platform
}

func generateOnekeymapBase(
	ctx context.Context,
	sourceDir string,
	outDir string,
	imp pluginapi.PluginImporter,
	logger *slog.Logger,
	tasks []genTask,
) {
	for _, t := range tasks {
		logger.InfoContext(ctx, "generating base", "platform", t.name, "src", t.source)
		sourcePath := filepath.Join(sourceDir, t.source)
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			logger.ErrorContext(ctx, "open source", "path", sourcePath, "error", err)
			os.Exit(1)
		}

		km, err := imp.Import(ctx, sourceFile, pluginapi.PluginImportOption{})
		_ = sourceFile.Close()
		if err != nil {
			logger.ErrorContext(ctx, "import json", "src", t.source, "error", err)
			os.Exit(1)
		}

		outPath := filepath.Join(outDir, t.outfile)
		fp, err := os.Create(outPath)
		if err != nil {
			logger.ErrorContext(ctx, "create output", "path", outPath, "error", err)
			os.Exit(1)
		}
		if err := keymap.Save(fp, km, keymap.SaveOptions{Platform: t.platform}); err != nil {
			_ = fp.Close()
			logger.ErrorContext(ctx, "write base json", "path", outPath, "error", err)
			os.Exit(1)
		}
		if err := fp.Close(); err != nil {
			logger.WarnContext(ctx, "close file", "path", outPath, "error", err)
		}
		logger.InfoContext(ctx, "wrote base", "path", outPath)
	}
}

func generateVSCodeBase(
	ctx context.Context,
	f *devGenerateBaseFlags,
	mc *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) {
	plugin := vscode.New(mc, logger, recorder)
	imp, err := plugin.Importer()
	if err != nil {
		logger.ErrorContext(ctx, "get vscode importer", "error", err)
		os.Exit(1)
	}

	tasks := []genTask{
		{name: "mac", source: "macos.keybindings.json", outfile: "vscode-mac.json", platform: platform.PlatformMacOS},
		{
			name:     "windows",
			source:   "windows.keybindings.json",
			outfile:  "vscode-windows.json",
			platform: platform.PlatformWindows,
		},
		{
			name:     "linux",
			source:   "linux.keybindings.json",
			outfile:  "vscode-linux.json",
			platform: platform.PlatformLinux,
		},
	}
	generateOnekeymapBase(ctx, f.VSCodeSourceDir, f.OutDir, imp, logger, tasks)
}

func generateZedBase(
	ctx context.Context,
	f *devGenerateBaseFlags,
	mc *mappings.MappingConfig,
	logger *slog.Logger,
	recorder metrics.Recorder,
) {
	plugin := zed.New(mc, logger, recorder)
	imp, err := plugin.Importer()
	if err != nil {
		logger.ErrorContext(ctx, "get zed importer", "error", err)
		os.Exit(1)
	}

	tasks := []genTask{
		{name: "mac", source: "default-macos.json", outfile: "zed-mac.json", platform: platform.PlatformMacOS},
		{
			name:     "windows",
			source:   "default-windows.json",
			outfile:  "zed-windows.json",
			platform: platform.PlatformWindows,
		},
		{name: "linux", source: "default-linux.json", outfile: "zed-linux.json", platform: platform.PlatformLinux},
	}
	generateOnekeymapBase(ctx, f.ZedSourceDir, f.OutDir, imp, logger, tasks)
}

func loadAndFlattenIJXML(_ context.Context, dir, file string) (*ij.KeymapXML, *keymapMetadata, error) {
	visited := map[string]bool{}
	var chain []*ij.KeymapXML
	n := file
	for {
		if visited[n] {
			return nil, nil, fmt.Errorf("cyclic parent reference: %s", n)
		}
		visited[n] = true
		p := filepath.Join(dir, n)
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", p, err)
		}
		var doc ij.KeymapXML
		if err := xml.Unmarshal(data, &doc); err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", p, err)
		}
		chain = append(chain, &doc)
		par := strings.TrimSpace(doc.Parent)
		if par == "" {
			break
		}
		n = par + ".xml"
	}

	acc := &ij.KeymapXML{}
	order := []string{}
	actions := map[string]*ij.ActionXML{}
	// Track which actions come from the top-level file (not inherited)
	topLevelActions := make(map[string]bool)

	// Mark actions from the top-level file (first in chain, before we reverse)
	if len(chain) > 0 {
		for _, a := range chain[0].Actions {
			topLevelActions[a.ID] = true
		}
	}

	for i := len(chain) - 1; i >= 0; i-- {
		for _, a := range chain[i].Actions {
			ax := actions[a.ID]
			if ax == nil {
				na := a
				na.KeyboardShortcuts = nil
				actions[a.ID] = &na
				order = append(order, a.ID)
				ax = &na
			}
			seen := map[string]struct{}{}
			for _, ks := range ax.KeyboardShortcuts {
				seen[ks.First+"\x00"+ks.Second] = struct{}{}
			}
			for _, ks := range a.KeyboardShortcuts {
				k := ks.First + "\x00" + ks.Second
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = struct{}{}
				ax.KeyboardShortcuts = append(ax.KeyboardShortcuts, ks)
			}
		}
	}

	for _, id := range order {
		if ax, ok := actions[id]; ok {
			acc.Actions = append(acc.Actions, *ax)
		}
	}

	meta := &keymapMetadata{
		topLevelActions: topLevelActions,
	}
	return acc, meta, nil
}

// keymapMetadata contains metadata about the flattened keymap
type keymapMetadata struct {
	topLevelActions map[string]bool // actions defined in top-level file (not inherited)
}

// convertControlToMetaForMac transforms 'control' modifier to 'meta' for inherited actions only.
// This matches IntelliJ's runtime behavior on macOS:
// - Actions inherited from $default.xml have 'control' auto-mapped to Command (meta)
// - Actions explicitly defined in Mac OS X.xml keep their 'control' keys as-is
// For example: inherited "$Copy" with "control C" becomes "meta C",
// but explicit "CodeCompletion" with "control SPACE" stays "control SPACE"
func convertControlToMetaForMac(doc *ij.KeymapXML, meta *keymapMetadata) {
	if doc == nil || meta == nil {
		return
	}

	for i := range doc.Actions {
		actionID := doc.Actions[i].ID
		// Only convert actions that are NOT explicitly defined in the top-level file
		// (i.e., they are inherited from parent keymaps like $default.xml)
		if meta.topLevelActions[actionID] {
			continue // Skip actions explicitly defined in Mac OS X.xml
		}

		for j := range doc.Actions[i].KeyboardShortcuts {
			ks := &doc.Actions[i].KeyboardShortcuts[j]
			ks.First = replaceControlWithMeta(ks.First)
			ks.Second = replaceControlWithMeta(ks.Second)
		}
	}
}

// replaceControlWithMeta replaces "control" modifier with "meta" in a keystroke string.
// Handles both lowercase "control" and uppercase "CONTROL" to be safe.
func replaceControlWithMeta(keystroke string) string {
	if keystroke == "" {
		return keystroke
	}

	result := strings.ReplaceAll(keystroke, "control ", "meta ")
	result = strings.ReplaceAll(result, "CONTROL ", "meta ")

	if strings.HasSuffix(result, "control") {
		result = strings.TrimSuffix(result, "control") + "meta"
	}
	if strings.HasSuffix(result, "CONTROL") {
		result = strings.TrimSuffix(result, "CONTROL") + "meta"
	}

	return result
}
