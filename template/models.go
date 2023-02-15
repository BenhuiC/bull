package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/models.go.tpl
	models string
)

var ModelMap = map[string]*template.Template{
	"model.go": template.Must(template.New("model").Parse(models)),
}
