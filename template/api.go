package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/common.go.tpl
	common string
	//go:embed temp/request.go.tpl
	request string
	//go:embed temp/response.go.tpl
	response string
	//go:embed temp/api.proto.tpl
	api string
	//go:embed temp/applogger.go.tpl
	applogger string
	//go:embed temp/exceptions.go.tpl
	exceptions string
	//go:embed temp/language.go.tpl
	language string
	//go:embed temp/serve.go.tpl
	serve string
)

var ApiMap = map[string]*template.Template{
	"h/applogger.go":  template.Must(template.New("applogger").Parse(applogger)),
	"h/exceptions.go": template.Must(template.New("exceptions").Parse(exceptions)),
	"h/language.go":   template.Must(template.New("language").Parse(language)),
	"h/common.go":     template.Must(template.New("common").Parse(common)),
	"h/request.go":    template.Must(template.New("request").Parse(request)),
	"h/response.go":   template.Must(template.New("response").Parse(response)),
	"serve.go":        template.Must(template.New("serve").Parse(serve)),
	"api.proto":       template.Must(template.New("api").Parse(api)),
}
