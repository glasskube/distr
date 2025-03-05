package main

import "os"

func ScratchDir() string {
	if dir := os.Getenv("DISTR_AGENT_SCRATCH_DIR"); dir != "" {
		return dir
	}
	return "./scratch"
}
