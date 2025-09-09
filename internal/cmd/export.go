package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/views"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

var (
	toFlag       *string
	exportInput  *string
	exportOutput *string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a universal keymap to an editor's format",
	RunE: func(cmd *cobra.Command, args []string) error {
		onekeymapPlaceHolder := viper.GetString("onekeymap")
		err := prepareExportInputFlags(cmd, onekeymapPlaceHolder)
		if err != nil {
			return err
		}
		logger.Info("Exporting config", "to", *toFlag, "from", *exportInput)

		file, err := os.Open(*exportInput)
		if err != nil {
			logger.Error("Failed to open input file", "error", err)
			return err
		}
		defer func() {
			if err := file.Close(); err != nil {
				logger.Error("Failed to close input file", "error", err)
			}
		}()

		setting, err := keymap.Load(file)
		if err != nil {
			logger.Error("Failed to load config file", "error", err)
			return err
		}

		opts := exportapi.ExportOptions{EditorType: pluginapi.EditorType(*toFlag)}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(*exportOutput), 0o755); err != nil {
			logger.Error("Failed to create output directory", "dir", filepath.Dir(*exportOutput), "error", err)
			return err
		}

		// Prepare base reader from existing output file if present for diff calculation
		var base io.Reader
		if f, err := os.Open(*exportOutput); err == nil {
			defer func() { _ = f.Close() }()
			base = f
		} else if !os.IsNotExist(err) {
			// Non-ENOENT error opening base file; log and continue without base
			logger.Warn("Failed to open existing output as base", "error", err)
		}
		opts.Base = base
		opts.DiffType = keymapv1.ExportKeymapRequest_ASCII_DIFF

		// Export to memory buffer first for preview, optional confirmation, and then write
		var mem bytes.Buffer
		report, err := exportService.Export(cmd.Context(), &mem, setting, opts)
		if err != nil {
			logger.Error("export failed", "error", err)
			return err
		}

		// Show diff preview
		fmt.Println("================ Export Diff Preview ================")
		if report != nil && strings.TrimSpace(report.Diff) != "" {
			fmt.Println(report.Diff)
		} else {
			fmt.Println("(no diff available)")
		}
		fmt.Println("=====================================================")

		// TODO(xinnjie): optimize skip writing when output is the same. Whether same or not should not rely on report.Diff

		// Confirm before writing only when interactive
		if *interactive {
			if !confirm(*exportOutput) {
				fmt.Println("Export canceled; no changes were written.")
				return nil
			}
		}

		// Backup existing file if requested
		if *backup {
			if backupPath, err := backupIfExists(*exportOutput); err != nil {
				logger.Warn("Failed to backup existing file", "path", *exportOutput, "error", err)
			} else if backupPath != "" {
				logger.Info("Created backup of existing config", "backup", backupPath)
			}
		}

		// Write buffer to the target file
		outputFile, err := os.OpenFile(*exportOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			logger.Error("Failed to create output file", "error", err)
			return err
		}
		defer func() {
			if err := outputFile.Close(); err != nil {
				logger.Error("Failed to close output file", "error", err)
			}
		}()
		if _, err := outputFile.Write(mem.Bytes()); err != nil {
			logger.Error("Failed to write to output file", "error", err)
			return err
		}

		logger.Info("Successfully exported keymap", "to", *toFlag, "output", *exportOutput)
		return nil
	},
}

func prepareExportInputFlags(cmd *cobra.Command, onekeymapPlaceholder string) error {
	if *interactive {
		needSelectEditor := !cmd.Flags().Changed("to") || *toFlag == ""
		needInput := !cmd.Flags().Changed("input") || *exportInput == ""
		needOutput := !cmd.Flags().Changed("output") || *exportOutput == ""

		if needSelectEditor || needInput || needOutput {
			if err := runExportForm(pluginRegistry, toFlag, exportInput, exportOutput, onekeymapPlaceholder, needSelectEditor, needInput, needOutput); err != nil {
				return err
			}
		}
	} else {
		if *toFlag == "" {
			return fmt.Errorf("flag --to is required")
		}
		if !cmd.Flags().Changed("input") && onekeymapPlaceholder != "" {
			*exportInput = onekeymapPlaceholder
		}
	}

	// Validate selected editor plugin exists
	p, ok := pluginRegistry.Get(pluginapi.EditorType(*toFlag))
	if !ok {
		logger.Error("Editor not found", "editor", *toFlag)
		return fmt.Errorf("editor %s not found", *toFlag)
	}

	// Determine input
	if *exportInput == "" {
		// Fallback to config default in case it wasn't provided
		*exportInput = onekeymapPlaceholder
	}
	// Determine output path
	if *exportOutput == "" {
		if v, err := p.DefaultConfigPath(); err == nil {
			*exportOutput = v[0]
		}
	}
	return nil
}

func runExportForm(pluginRegistry *plugins.Registry, to, input, output *string, onekeymapConfigPlaceHolder string,
	needSelectEditor, needInput, needOutput bool) error {
	m, err := views.NewOutputFormModel(pluginRegistry, needSelectEditor, needInput, needOutput, to, input, output, onekeymapConfigPlaceHolder)
	if err != nil {
		return err
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		if errors.Is(err, tea.ErrProgramKilled) {
			os.Exit(0)
		}
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(exportCmd)
	toFlag = exportCmd.Flags().String("to", "", "Target editor to export to")
	exportInput = exportCmd.Flags().String("input", "", "Path to the source onekeymap.json file")
	exportOutput = exportCmd.Flags().String("output", "", "Optional: Path to the target editor's config file")

	_ = exportCmd.RegisterFlagCompletionFunc("to", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return pluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
	})
}
