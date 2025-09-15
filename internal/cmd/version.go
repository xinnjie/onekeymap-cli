package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("OneKeymap CLI\n")
		cmd.Printf("Version:    %s\n", version)
		cmd.Printf("Built:      %s\n", buildTime)
		cmd.Printf("Git Commit: %s\n", gitCommit)
		cmd.Printf("Go Version: %s\n", "go1.23.5") // Could be dynamic if needed
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
