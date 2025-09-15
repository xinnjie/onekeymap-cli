package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/views"
)

type viewFlags struct {
	file string
}

func NewCmdView() *cobra.Command {
	f := viewFlags{}
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View current OneKeymapSetting in a read-only TUI",
		RunE:  viewRun(&f),
		Args:  cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.file, "file", "", "Path to onekeymap.json (defaults to config value)")

	return cmd
}

func viewRun(f *viewFlags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		path := f.file
		if path == "" {
			path = viper.GetString("onekeymap")
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open onekeymap config: %w", err)
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
