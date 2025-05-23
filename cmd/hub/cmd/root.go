package cmd

import (
	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/spf13/cobra"
)

var RootCommand = &cobra.Command{
	Use:     "distr",
	Version: buildconfig.Version(),
}
