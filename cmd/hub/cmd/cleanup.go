package cmd

import (
	"github.com/spf13/cobra"
)

var CleanupCommand = &cobra.Command{
	Use: "cleanup",
}

func init() {
	RootCommand.AddCommand(CleanupCommand)
}
