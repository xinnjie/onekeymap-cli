package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
)

type viewFlags struct {
	file string
}

func NewCmdView() *cobra.Command {
	f := viewFlags{}
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View current OneKeymapSetting in a read-only TUI",
		RunE: viewRun(&f, func() (*mappings.MappingConfig, *slog.Logger) {
			return cmdMappingConfig, cmdLogger
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.file, "file", "", "Path to onekeymap.json (defaults to config value)")

	return cmd
}

func viewRun(
	f *viewFlags,
	dependencies func() (*mappings.MappingConfig, *slog.Logger),
) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		mappingConfig, logger := dependencies()
		path := f.file
		if path == "" {
			path = viper.GetString("onekeymap")
		}

		// Resolve absolute path for file watching
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}

		file, err := os.Open(absPath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to open onekeymap config: %w", err)
			}
			if err := ensureConfigFile(absPath); err != nil {
				return err
			}

			file, err = os.Open(absPath)
			if err != nil {
				return fmt.Errorf("failed to open onekeymap config: %w", err)
			}
		}
		defer func() { _ = file.Close() }()

		setting, err := keymap.Load(file, keymap.LoadOptions{})
		if err != nil {
			return fmt.Errorf("failed to parse onekeymap config: %w", err)
		}

		m := views.NewKeymapViewModel(setting, mappingConfig, absPath)
		p := tea.NewProgram(m)

		go watchFile(logger, absPath, p)

		_, err = p.Run()
		return err
	}
}

// watchFile monitors the file for changes and sends messages to the tea program.
// It watches the parent directory to handle atomic writes (used by editors like Helix, Vim, etc.)
// where the file is replaced via rename operations.
func watchFile(logger *slog.Logger, path string, p *tea.Program) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		p.Send(views.FileErrorMsg{Err: fmt.Errorf("failed to create watcher: %w", err)})
		return
	}
	defer watcher.Close()

	// Watch the parent directory to catch atomic writes (rename operations)
	dir := filepath.Dir(path)
	fileName := filepath.Base(path)
	logger.Debug("watching config file", "file", path, "dir", dir)

	if err := watcher.Add(dir); err != nil {
		p.Send(views.FileErrorMsg{Err: fmt.Errorf("failed to watch directory: %w", err)})
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only process events for our target file
			if filepath.Base(event.Name) != fileName {
				continue
			}

			// Handle different types of file events
			// Write: Direct file modification
			// Create: File created (after atomic write)
			// Rename: File renamed (atomic write completion)
			// Remove: File removed (part of atomic write process)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) ||
				event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
				logger.Debug("file event detected",
					"path", event.Name,
					"op", event.Op.String())
				p.Send(views.FileChangedMsg{Path: event.Name})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error("file watcher error", "error", err)
			p.Send(views.FileErrorMsg{Err: err})
		}
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

	if err := keymap.Save(file, keymap.Keymap{}, keymap.SaveOptions{}); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to initialize onekeymap config: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close onekeymap config: %w", err)
	}

	return nil
}
