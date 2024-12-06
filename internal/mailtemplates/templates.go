package mailtemplates

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*
var embeddedFS embed.FS

var templates *template.Template

func init() {
	if fsys, err := fs.Sub(embeddedFS, "templates"); err != nil {
		panic(err)
	} else {
		templates = template.Must(parse(fsys, "*.html", "fragments/*.html", "css/*.css"))
	}
}

func parse(fsys fs.FS, patterns ...string) (*template.Template, error) {
	var t *template.Template = template.New("")
	for _, p := range patterns {
		if files, err := fs.Glob(fsys, p); err != nil {
			return nil, err
		} else {
			for _, file := range files {
				if ft, err := template.ParseFS(fsys, file); err != nil {
					return nil, err
				} else if _, err := t.AddParseTree(file, ft.Tree); err != nil {
					return nil, err
				}
			}
		}
	}
	return t, nil
}

func Welcome() (*template.Template, any) {
	return templates.Lookup("welcome.html"), nil
}
