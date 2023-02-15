package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/root.go.tpl
	rootCmd string
	//go:embed temp/server.go.tpl
	serverCmd string
)

var CmdMap = map[string]*template.Template{
	"root.go":   template.Must(template.New("root").Parse(rootCmd)),
	"server.go": template.Must(template.New("server").Parse(serverCmd)),
}
