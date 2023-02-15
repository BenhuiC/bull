package h

import (
	"fmt"
)

const (
	CodeServerError = "serverErr"
	CodeOK          = "success"
)

type Exception struct {
	Code      string          `yaml:"code"`
	Default   string          `yaml:"default"`
	Languages map[Lang]string `yaml:"languages"`
}

func (e Exception) Error() string {
	return fmt.Sprintf("Code:%s ,Message: %s", e.Code, e.Default)
}

var (
	CannotAccessErr = Exception{
		Code:    "error.data.cannot-access",
		Default: "No permission!",
		Languages: map[Lang]string{
			Lang_zh_CN: "无权限！",
			Lang_en_US: "No permission!",
		},
	}
	SetCannotUploadErr = Exception{
		Code:    "error.data.datset-cannot-upload",
		Default: "No more data can be added to the data set!",
		Languages: map[Lang]string{
			Lang_zh_CN: "数据集无法再添加数据！",
			Lang_en_US: "No more data can be added to the data set!",
		},
	}
	ParamErr = Exception{
		Code:    "error.data.invalid-param",
		Default: "The data parameter is wrong! %v",
		Languages: map[Lang]string{
			Lang_zh_CN: "数据参数错误！%v",
			Lang_en_US: "The data parameter is wrong!",
		},
	}
	NameConflictErr = Exception{
		Code:    "error.data.name-conflict",
		Default: "Repeat the name!",
		Languages: map[Lang]string{
			Lang_zh_CN: "重复名称！",
			Lang_en_US: "Repeat the name!",
		},
	}
	NotFoundErr = Exception{
		Code:    "error.data.not-found",
		Default: "The data was not found!",
		Languages: map[Lang]string{
			Lang_zh_CN: "该数据未找到！",
			Lang_en_US: "The data was not found!",
		},
	}
	TextUnknownError = Exception{
		Code:    "error.data.unknown-error",
		Default: "unknown error: %v",
		Languages: map[Lang]string{
			Lang_zh_CN: "未知错误：%v",
			Lang_en_US: "",
		},
	}
	ServiceCallErr = Exception{
		Code:    "error.data.service-call",
		Default: "serviceCallErr",
		Languages: map[Lang]string{
			Lang_zh_CN: "服务异常",
			Lang_en_US: "service err",
		},
	}
)

func ParseException(lang Lang, buErr Exception, vals ...interface{}) (string, string) {
	msg, ok := buErr.Languages[lang]
	if !ok {
		msg = buErr.Default
	}
	if len(vals) > 0 {
		msg = fmt.Sprintf(msg, vals...)
	}
	return buErr.Code, msg
}
