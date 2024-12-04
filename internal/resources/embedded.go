package resources

import (
	"embed"
)

//go:embed embedded
var embeddedFs embed.FS

func Get(filename string) ([]byte, error) {
	return embeddedFs.ReadFile(filename)
}
