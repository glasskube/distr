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
	if bytes, err := Get(filename); err != nil {
		return nil, fmt.Errorf("failed to read template file %v: %w", filename, err)
	} else if tmpl, err := template.New("").Parse(string(bytes)); err != nil {
		return nil, fmt.Errorf("failed to parse template %v: %w", filename, err)
	} else {
		return tmpl, nil
	}
}
