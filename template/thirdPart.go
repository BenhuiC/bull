package template

import "text/template"
import _ "embed"

var (
	//go:embed temp/thirdErrors.proto.tpl
	errors string
	//go:embed temp/thirdGoGo.proto.tpl
	gogo string
	//go:embed temp/thirdApiAnnotation.proto.tpl
	annotations string
	//go:embed temp/thirdApiClient.proto.tpl
	client string
	//go:embed temp/thirdApiFieldBehavior.proto.tpl
	fieldBehavior string
	//go:embed temp/thirdApiHttp.proto.tpl
	http string
	//go:embed temp/thirdApiHttpBody.proto.tpl
	httpBody string
	//go:embed temp/thirdDescriptor.proto.tpl
	descriptor string
	//go:embed temp/thirdValidate.proto.tpl
	validate string
)

var ThirdPartMap = map[string]*template.Template{
	"errors/errors.proto":            template.Must(template.New("errors").Parse(errors)),
	"gogoproto/gogo.proto":           template.Must(template.New("gogo").Parse(gogo)),
	"google/api/annotations.proto":   template.Must(template.New("annotations").Parse(annotations)),
	"google/api/client.proto":        template.Must(template.New("client").Parse(client)),
	"google/api/fieldBehavior.proto": template.Must(template.New("fieldBehavior").Parse(fieldBehavior)),
	"google/api/http.proto":          template.Must(template.New("http").Parse(http)),
	"google/api/httpBody.proto":      template.Must(template.New("httpBody").Parse(httpBody)),
	"protobuf/descriptor.proto":      template.Must(template.New("descriptor").Parse(descriptor)),
	"validate/validate.proto":        template.Must(template.New("validate").Parse(validate)),
}
