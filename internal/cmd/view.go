package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type viewFlags struct {
	file string
}

func NewCmdView() *cobra.Command {
	f := viewFlags{}
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View current OneKeymapSetting in a read-only TUI",
		RunE: viewRun(&f, func() *mappings.MappingConfig {
			return cmdMappingConfig
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.file, "file", "", "Path to onekeymap.json (defaults to config value)")

	return cmd
}

func viewRun(f *viewFlags, dependencies func() *mappings.MappingConfig) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		mappingConfig := dependencies()
		path := f.file
		if path == "" {
			path = viper.GetString("onekeymap")
		}
		file, err := os.Open(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to open onekeymap config: %w", err)
			}
			if err := ensureConfigFile(path); err != nil {
				return err
			}

			file, err = os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open onekeymap config: %w", err)
			}
		}
		defer func() { _ = file.Close() }()

		setting, err := keymap.Load(file)
		if err != nil {
			return fmt.Errorf("failed to parse onekeymap config: %w", err)
		}

		m := views.NewKeymapViewModel(setting, mappingConfig)
		_, err = tea.NewProgram(m).Run()
		return err
	}
}

// Create empty config file if it doesn't exist
func ensureConfigFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to initialize onekeymap config: %w", err)
	}

	if err := keymap.Save(file, &keymapv1.Keymap{}); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to initialize onekeymap config: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close onekeymap config: %w", err)
	}

	return nil
}
