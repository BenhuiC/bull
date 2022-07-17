package template

import "text/template"

var ApiMap = map[string]*template.Template{
	"common":   template.Must(template.New("common").Parse(Common())),
	"request":  template.Must(template.New("request").Parse(Request())),
	"response": template.Must(template.New("response").Parse(Response())),
	"api":      template.Must(template.New("api").Parse(Api())),
}

// Common projectDir/api/common.go
func Common() string {
	return `
package h

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	LogSkipPaths = map[string]bool{
		"/health":  true,
		"/metrics": true,
	}
)

func MidRecovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err interface{}) {
		// 和AbortDirect配合，支持通过panic方式直接返回错误
		if e, ok := err.(Response); ok {
			c.JSON(http.StatusOK, e)
			return
		}
		R(c, CodeServerError, "ServerError", nil)
		logger.With(loggerFields(c)...).With(zap.Stack("stacks")).Errorf("panic: %v", err)
		// h.Error(c, err)
	})
}

func MidSetRequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader(RequestIDHeaderKey)
		if id == "" {
			id = uuid.NewString()
		}
		ctx.Set(RequestIDHeaderKey, id)
		ctx.Next()
	}
}

func MidLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()
		path := ctx.Request.URL.Path

		// Process request
		ctx.Next()

		// Log only when path is not being skipped
		if ok := LogSkipPaths[path]; !ok {
			raw := ctx.Request.URL.Path
			if ctx.Request.URL.RawQuery != "" {
				raw = raw + "?" + ctx.Request.URL.RawQuery
			}
			Latency := time.Since(start)
			Infof(ctx, "%s %s %d %dus", ctx.Request.Method, raw, ctx.Writer.Status(), Latency.Microseconds())
		}
	}
}
`
}

// Request projectDir/api/request.go
func Request() string {
	return `
package h

import "github.com/gin-gonic/gin"

const (
	RequestIDHeaderKey = "X-Request-Id"
)

func GetRequestID(c *gin.Context) string {
	if id, exist := c.Get(RequestIDHeaderKey); exist {
		return id.(string)
	}
	return ""
}
`
}

// Response projectDir/api/response.go
func Response() string {
	return `
package h

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func R(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   message,
		Data:      data,
		CreatedAt: time.Now(),
		RequestID: GetRequestID(c),
	})
}

func RR(c *gin.Context, data interface{}) {
	R(c, CodeOK, "success", data)
}

func RE(c *gin.Context, msg string) {
	R(c, CodeExceptionDefault, msg, nil)
}
` + "type Response struct {\n\tCode      int         `json:\"code\"`\n\tMessage   string      `json:\"message\"`\n\tData      interface{} `json:\"data,omitempty\"`\n\tCreatedAt time.Time   `json:\"createdAt\"`\n\tRequestID string      `json:\"requestID,omitempty\"`\n}"
}

// Api projectDir/api/api.proto
func Api() string {
	return `
syntax = "proto3";
package api;

option go_package="{{ .ProjectName }}";

import "google/api/annotations.proto";
import "gogoproto/gogo.proto";
import "google/protobuf/wrappers.proto";

service {{ .ProjectName}}Service {
  // 探活
  rpc Ping(Empty) returns (Pong){
    option(google.api.http) ={
      get: "/ping"
    };
  }
}

message Empty {}

message Pong {
  Pong string =1;
}
`
}
