package h

import "github.com/gin-gonic/gin"

type Lang = string

const (
	Lang_zh_CN = "zh_CN"
	Lang_en_US = "en_US"
)

func GetLanguage(c *gin.Context) Lang {
	l := c.Request.Header.Get("Accept-Language")
	if l == "" {
		return Lang_zh_CN
	}
	return l
}
