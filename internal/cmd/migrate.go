package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/views"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
)

var (
	migrateFrom   *string
	migrateTo     *string
	migrateInput  *string
	migrateOutput *string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate keymaps from one editor to another",
	RunE: func(cmd *cobra.Command, args []string) error {

		interactive, _ := cmd.Flags().GetBool("interactive")
		backup, _ := cmd.Flags().GetBool("backup")
		ctx := cmd.Context()

		if interactive {
			// In interactive mode, we can use the form to get missing values.
			if *migrateFrom == "" || *migrateTo == "" {
				model := views.NewMigrateFormModel(pluginRegistry, migrateFrom, migrateTo, migrateInput, migrateOutput)
				p := tea.NewProgram(model)
				if _, err := p.Run(); err != nil {
					logger.Error("failed to run interactive form", "error", err)
					return err
				}
			}
		}

		// After potentially running the form, `from` and `to` must be set.
		if *migrateFrom == "" || *migrateTo == "" {
			return errors.New("required flags 'from' and 'to' not set")
		}

		logger.Info("Migrating keymaps", "from", *migrateFrom, "to", *migrateTo)

		inputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(*migrateFrom))
		if !ok {
			logger.Error("failed to get input plugin", "from", *migrateFrom)
			return errors.New("failed to get input plugin")
		}
		outputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(*migrateTo))
		if !ok {
			logger.Error("failed to get output plugin", "to", *migrateTo)
			return errors.New("failed to get output plugin")
		}

		if *migrateInput == "" {
			v, err := inputPlugin.DefaultConfigPath()
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			*migrateInput = v[0]
		}

		if *migrateOutput == "" {
			v, err := outputPlugin.DefaultConfigPath()
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			*migrateOutput = v[0]
		}

		inputStream, err := os.Open(*migrateInput)
		if err != nil {
			logger.Error("failed to open input file", "error", err)
			return err
		}
		defer func() { _ = inputStream.Close() }()

		importOpts := importapi.ImportOptions{
			EditorType:  pluginapi.EditorType(*migrateFrom),
			InputStream: inputStream,
		}
		importResult, err := importService.Import(ctx, importOpts)
		if err != nil {
			logger.Error("migrate failed during import step", "error", err)
			return err
		}

		logger.Info("Import complete.")
		if importResult != nil {
			logger.Debug("Import Report", "report", importResult.Report)
		}

		if importResult == nil || importResult.Setting == nil {
			logger.Warn("No imported keymaps to export; aborting migrate")
			return nil
		}

		// Prepare base reader from existing output file if present for diff calculation
		var base io.Reader
		if f, err := os.Open(*migrateOutput); err == nil {
			defer func() { _ = f.Close() }()
			base = f
		} else if !os.IsNotExist(err) {
			logger.Warn("Failed to open existing output as base", "error", err)
		}

		// Export to memory buffer first for preview, optional confirmation, and then write
		var mem bytes.Buffer
		exportOpts := exportapi.ExportOptions{EditorType: pluginapi.EditorType(*migrateTo), Base: base}
		exportReport, err := exportService.Export(ctx, &mem, importResult.Setting, exportOpts)
		if err != nil {
			logger.Error("migrate failed during export step", "error", err)
			return err
		}

		// Show diff preview
		fmt.Println("================ Migrate Diff Preview ================")
		if exportReport != nil && strings.TrimSpace(exportReport.Diff) != "" {
			fmt.Println(exportReport.Diff)
		} else {
			fmt.Println("(no changes)")
		}
		fmt.Println("=====================================================")

		// Confirm before writing only when interactive
		if interactive {
			if !confirm(*migrateOutput) {
				logger.Info("Migration canceled; no changes were written.")
				return nil
			}
		}

		// Backup existing file if requested
		if backup {
			if backupPath, err := backupIfExists(*migrateOutput); err != nil {
				logger.Warn("Failed to backup existing file", "path", *migrateOutput, "error", err)
			} else if backupPath != "" {
				logger.Info("Created backup of existing config", "backup", backupPath)
			}
		}

		// Write buffer to the target file
		outputStream, err := os.OpenFile(*migrateOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			logger.Error("failed to create output file", "error", err)
			return err
		}
		defer func() { _ = outputStream.Close() }()
		if _, err := outputStream.Write(mem.Bytes()); err != nil {
			logger.Error("failed to write to output file", "error", err)
			return err
		}

		logger.Info("Migration complete!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateFrom = migrateCmd.Flags().String("from", "", "Source editor, valid values: vscode, zed")
	migrateTo = migrateCmd.Flags().String("to", "", "Target editor, valid values: vscode, zed")
	migrateInput = migrateCmd.Flags().String("input", "", "Path to source editor config")
	migrateOutput = migrateCmd.Flags().String("output", "", "Path to target editor config")

	migrateCmd.Flags().Bool("backup", false, "Create a backup of the target editor's keymap")
}
