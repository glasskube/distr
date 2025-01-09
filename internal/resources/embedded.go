package resources

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed embedded
var embeddedFs embed.FS

func Get(filename string) ([]byte, error) {
	return embeddedFs.ReadFile(filename)
}

func GetTemplate(filename string) (*template.Template, error) {
	if bytes, err := Get("embedded/agent-base.yaml"); err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	} else if tmpl, err := template.New("agent").Parse(string(bytes)); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	} else {
		return tmpl, nil
	}
}
