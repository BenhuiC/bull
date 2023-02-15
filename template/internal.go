package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/base.go.tpl
	base string
)

var InternalMap = map[string]*template.Template{
	"base.go": template.Must(template.New("base").Parse(base)),
}
