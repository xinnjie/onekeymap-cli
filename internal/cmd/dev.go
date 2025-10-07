package cmd

import (
	"github.com/spf13/cobra"
)

type devFlags struct {
	// No flags for dev command currently
}

func NewCmdDev() *cobra.Command {
	f := devFlags{}
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Commands for development",
		Long:  `Commands for development, may be unstable and change without notice.`,
		Run:   devRun(&f),
		Args:  cobra.ExactArgs(0),
	}

	return cmd
}

func devRun(_ *devFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// Empty implementation - this is a parent command for dev subcommands
	}
}
