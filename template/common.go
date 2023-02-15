package template

import (
	_ "embed"
	"text/template"
)

var (
	//go:embed temp/gitignore.tpl
	gitignore string
	//go:embed temp/main.go.tpl
	main string
	//go:embed temp/app.yaml.tpl
	configFile string
	//go:embed temp/Makefile.tpl
	makefile string
)

var CommonMap = map[string]*template.Template{
	".gitignore": template.Must(template.New("gitignore").Parse(gitignore)),
	"main.go":    template.Must(template.New("main").Parse(main)),
	"app.yaml":   template.Must(template.New("config").Parse(configFile)),
	"Makefile":   template.Must(template.New("makefile").Parse(makefile)),
}
