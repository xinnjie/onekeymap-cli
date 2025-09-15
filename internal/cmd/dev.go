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
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run:  devRun(&f),
		Args: cobra.ExactArgs(0),
	}

	return cmd
}

func devRun(f *devFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// Empty implementation - this is a parent command for dev subcommands
	}
}
