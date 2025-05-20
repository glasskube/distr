package main

import (
	"os"

	"github.com/glasskube/distr/cmd/hub/cmd"
)

func main() {
	if err := cmd.RootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
