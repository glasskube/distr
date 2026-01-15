package resources

import (
	"embed"
	"fmt"
	"io/fs"
	"text/template"

	"github.com/distr-sh/distr/internal/util"
)

var (
	//go:embed embedded
	embeddedFs embed.FS
	fsys       = util.Require(fs.Sub(embeddedFs, "embedded"))
	templates  = map[string]*template.Template{}
)

func Get(name string) ([]byte, error) {
	return fs.ReadFile(fsys, name)
}

func GetTemplate(name string) (*template.Template, error) {
	if tmpl, ok := templates[name]; ok {
		return tmpl, nil
	} else if tmpl, err := template.ParseFS(fsys, name); err != nil {
		return nil, fmt.Errorf("failed to parse template %v: %w", name, err)
	} else {
		templates[name] = tmpl
		return tmpl, nil
	}
}
