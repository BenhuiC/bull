package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/config.go.tpl
	config string
	//go:embed temp/logger.go.tpl
	logger string
	//go:embed temp/extendLogger.go.tpl
	extendLogger string
	//go:embed temp/util.go.tpl
	util string
	//go:embed temp/workerContext.go.tpl
	workerCtx string
	//go:embed temp/workerMeta.go.tpl
	workerMeta string
	//go:embed temp/workerRunner.go.tpl
	workerRunner string
	//go:embed temp/workerStatus.go.tpl
	workerStatus string
	//go:embed temp/worker.go.tpl
	worker string
)

var PkgMap = map[string]*template.Template{
	"config/config.go":       template.Must(template.New("config").Parse(config)),
	"log/logger.go":          template.Must(template.New("logger").Parse(logger)),
	"log/extendLogger.go":    template.Must(template.New("extendLogger").Parse(extendLogger)),
	"util/util.go":           template.Must(template.New("util").Parse(util)),
	"worker/workerCtx.go":    template.Must(template.New("workerCtx").Parse(workerCtx)),
	"worker/workerMeta.go":   template.Must(template.New("workerMeta").Parse(workerMeta)),
	"worker/workerStatus.go": template.Must(template.New("workerStatus").Parse(workerStatus)),
	"worker/worker.go":       template.Must(template.New("worker").Parse(worker)),
}
